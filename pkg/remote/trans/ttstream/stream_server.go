/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ttstream

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/gopkg/protocol/ttheader"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/streaming"
	"github.com/cloudwego/kitex/pkg/transmeta"
)

var _ ServerStreamMeta = (*serverStream)(nil)

func newServerStream(ctx context.Context, writer streamWriter, smeta streamFrame) *serverStream {
	s := newBasicStream(ctx, writer, smeta)
	s.reader = newStreamReader()
	return &serverStream{s}
}

type serverStream struct {
	*stream
}

func (s *serverStream) SetHeader(hd streaming.Header) error {
	return s.writeHeader(hd)
}

func (s *serverStream) SendHeader(hd streaming.Header) error {
	if err := s.writeHeader(hd); err != nil {
		return err
	}
	return s.stream.sendHeader()
}

func (s *serverStream) SetTrailer(tl streaming.Trailer) error {
	return s.writeTrailer(tl)
}

func (s *serverStream) RecvMsg(ctx context.Context, req any) error {
	return s.stream.RecvMsg(ctx, req)
}

// SendMsg should send left header first
func (s *serverStream) SendMsg(ctx context.Context, res any) error {
	if st := atomic.LoadInt32(&s.state); st == streamStateInactive {
		return s.ctx.Err()
	}
	// todo: consider SendHeader and status change
	if s.wheader != nil {
		if err := s.sendHeader(); err != nil {
			return err
		}
	}
	return s.stream.SendMsg(ctx, res)
}

// CloseSend by serverStream will be called after server handler returned
// after CloseSend stream cannot be access again
func (s *serverStream) CloseSend(exception error) error {
	err := s.closeSend(exception)
	if err != nil {
		return err
	}
	return s.close(nil)
}

// closeRecv called only when server receiving Trailer Frame
func (s *serverStream) closeRecv(exception error) error {
	// client->server data transferring has stopped
	atomic.CompareAndSwapInt32(&s.state, streamStateActive, streamStateHalfCloseRemote)
	select {
	case s.headerSig <- streamSigInactive:
	default:
	}
	select {
	case s.trailerSig <- streamSigInactive:
	default:
	}
	s.reader.close(exception)
	return nil
}

func (s *serverStream) close(exception error) error {
	// support cascading cancel
	// we must cancel the ctx first before change the state change
	// otherwise Recv/Send will get the expected exception
	s.cancelFunc(exception)
	if atomic.SwapInt32(&s.state, streamStateInactive) == streamSigInactive {
		return nil
	}

	select {
	case s.headerSig <- streamSigInactive:
	default:
	}
	select {
	case s.trailerSig <- streamSigInactive:
	default:
	}

	s.reader.close(exception)
	s.runCloseCallback(exception)
	return nil
}

// === serverStream onRead callback

func (s *serverStream) onReadHeaderFrame(fr *Frame) error {
	if s.header != nil {
		return errUnexpectedHeader.newBuilder().withSide(serverSide).withCause(fmt.Errorf("stream[%d] already set header", s.sid))
	}
	s.header = fr.header
	select {
	case s.headerSig <- streamSigActive:
	default:
		return errUnexpectedHeader.newBuilder().withSide(serverSide).withCause(fmt.Errorf("stream[%d] already set header", s.sid))
	}
	klog.Debugf("stream[%s] read header: %v", s.method, fr.header)
	return nil
}

func (s *serverStream) onReadTrailerFrame(fr *Frame) (err error) {
	var exception error
	// when server-side returns non-biz error, it will be wrapped as ApplicationException stored in trailer frame payload
	if len(fr.payload) > 0 {
		// exception is type of (*thrift.ApplicationException)
		_, _, err = thrift.UnmarshalFastMsg(fr.payload, nil)
		exception = errApplicationException.newBuilder().withSide(serverSide).withCause(err)
	} else if len(fr.trailer) > 0 {
		// when server-side returns biz error, payload is empty and biz error information is stored in trailer frame header
		bizErr, err := transmeta.ParseBizStatusErr(fr.trailer)
		if err != nil {
			exception = errIllegalBizErr.newBuilder().withSide(serverSide).withCause(err)
		} else if bizErr != nil {
			// bizErr is independent of rpc exception handling
			exception = bizErr
		}
	}
	s.trailer = fr.trailer
	select {
	case s.trailerSig <- streamSigActive:
	default:
	}
	select {
	case s.headerSig <- streamSigNone:
		// if trailer arrived, we should return unblock stream.Header()
	default:
	}

	klog.Debugf("stream[%d] recv trailer: %v, exception: %v", s.sid, s.trailer, exception)
	// server-side stream recv trailer, we only need to close recv but still can send data
	return s.closeRecv(exception)
}

func (s *serverStream) onReadRstFrame(fr *Frame) (err error) {
	var rstEx *Exception
	var appEx *thrift.ApplicationException
	if len(fr.payload) > 0 {
		// exception is type of (*thrift.ApplicationException)
		appEx, err = unmarshalException(fr.payload)
		if err != nil {
			klog.Errorf("KITEX: stream[%d] unmarshal Exception in rst frame failed, err: %v", s.sid, err)
			appEx = defaultRstException
		}
	} else {
		klog.Infof("KITEX: stream[%d] recv rst frame without payload", s.sid)
		appEx = defaultRstException
	}

	// distract via information
	var via string
	if fr.header != nil {
		// cascading link initial node
		hdrVia, ok := fr.header[ttheader.HeaderVia]
		if ok {
			via = hdrVia
		}
	}
	rstEx = errUpstreamCancel.newBuilder().withSide(serverSide)
	if via != "" {
		rstEx = rstEx.setOrAppendVia(via).withCauseAndTypeId(appEx, appEx.TypeId())
	} else {
		rstEx = rstEx.withCauseAndTypeId(appEx, appEx.TypeId())
	}
	s.trailer = fr.trailer
	select {
	case s.trailerSig <- streamSigActive:
	default:
	}
	select {
	case s.headerSig <- streamSigNone:
		// if trailer arrived, we should return unblock stream.Header()
	default:
	}

	klog.Infof("stream[%d] recv header: %v, exception: %v", s.sid, fr.header, rstEx)
	// when receiving rst frame, we should close stream and there is no need to send rst frame
	return s.close(rstEx)
}

func (s *serverStream) closeTest(exception error, via string) error {
	// support cascading cancel
	// we must cancel the ctx first before change the state change
	// otherwise Recv/Send will get the expected exception
	s.cancelFunc(exception)
	if atomic.SwapInt32(&s.state, streamStateInactive) == streamSigInactive {
		return nil
	}

	select {
	case s.headerSig <- streamSigInactive:
	default:
	}
	select {
	case s.trailerSig <- streamSigInactive:
	default:
	}

	s.reader.close(exception)
	s.runCloseCallback(exception)

	s.reader.close(exception)
	if err := s.sendRst(exception, via); err != nil {
		return err
	}
	s.runCloseCallback(exception)
	return nil
}
