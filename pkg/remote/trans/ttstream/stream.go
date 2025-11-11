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
	"time"

	"github.com/bytedance/gopkg/lang/mcache"
	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/gopkg/protocol/ttheader"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
	ktransport "github.com/cloudwego/kitex/transport"
)

var (
	_ streaming.ClientStream          = (*clientStream)(nil)
	_ streaming.ServerStream          = (*serverStream)(nil)
	_ streaming.CloseCallbackRegister = (*stream)(nil)
)

var defaultRstException = thrift.NewApplicationException(13, "rst")

// newBasicStream is a common function creating stream basic fields,
// pls use newClientStream or newServerStream to create the real stream exposing to users
func newBasicStream(ctx context.Context, writer streamWriter, smeta streamFrame) *stream {
	s := new(stream)
	s.ctx = ctx
	s.rpcInfo = rpcinfo.GetRPCInfo(ctx)
	s.streamFrame = smeta
	s.writer = writer
	// Lazy initialization: allocate wheader and wtrailer only when needed
	// This avoids unnecessary allocations for streams that don't send headers/trailers
	return s
}

// streamFrame define a basic stream frame
type streamFrame struct {
	sid     int32
	method  string
	meta    IntHeader
	header  streaming.Header // key:value, key is full name
	trailer streaming.Trailer
}

const (
	streamSigNone     int32 = 0
	streamSigActive   int32 = 1
	streamSigInactive int32 = -1
	streamSigCancel   int32 = -2
)

const (
	streamStateActive          int32 = 0 // when stream is created, init state is active
	streamStateHalfCloseLocal  int32 = 1
	streamStateHalfCloseRemote int32 = 2
	streamStateInactive        int32 = 3
)

const (
	streamFlagHeaderSent uint8 = 1 << iota
	streamFlagTrailerSent
)

// stream is used to process frames and expose user APIs
type stream struct {
	streamFrame
	ctx      context.Context
	rpcInfo  rpcinfo.RPCInfo
	reader   *streamReader
	writer   streamWriter
	wheader  streaming.Header  // lazily initialized on first SetHeader/writeHeader call
	wtrailer streaming.Trailer // lazily initialized on first SetTrailer/writeTrailer call

	recvTimeout   time.Duration
	closeCallback []func(error)
	flags         uint8 // tracks if header/trailer have been sent
}

func (s *stream) Service() string {
	if len(s.header) == 0 {
		return ""
	}
	return s.header[ttheader.HeaderIDLServiceName]
}

func (s *stream) Method() string {
	return s.method
}

func (s *stream) TransportProtocol() ktransport.Protocol {
	return ktransport.TTHeaderStreaming
}

// SendMsg send a message to peer.
// In order to avoid underlying execution errors when the context passed in by the user does not
// contain information related to this RPC, the context specified when creating the stream is used
// here, and the context passed in by the user is ignored.
func (s *stream) SendMsg(ctx context.Context, msg any) (err error) {
	// encode payload
	payload, err := EncodePayload(s.ctx, msg)
	if err != nil {
		return err
	}
	// tracing send size
	ri := s.rpcInfo
	if ri != nil && ri.Stats() != nil {
		if rpcStats := rpcinfo.AsMutableRPCStats(ri.Stats()); rpcStats != nil {
			rpcStats.IncrSendSize(uint64(len(payload)))
		}
	}
	// send data frame
	return s.writeFrame(dataFrameType, nil, nil, payload)
}

func (s *stream) RecvMsg(ctx context.Context, data any) error {
	nctx := s.ctx
	if s.recvTimeout > 0 {
		var cancel context.CancelFunc
		nctx, cancel = context.WithTimeout(nctx, s.recvTimeout)
		defer cancel()
	}
	payload, err := s.reader.output(nctx)
	if err != nil {
		return err
	}
	err = DecodePayload(nctx, payload, data)
	// payload will not be access after decode
	mcache.Free(payload)

	// tracing recv size
	ri := s.rpcInfo
	if ri != nil && ri.Stats() != nil {
		if rpcStats := rpcinfo.AsMutableRPCStats(ri.Stats()); rpcStats != nil {
			rpcStats.IncrRecvSize(uint64(len(payload)))
		}
	}
	return err
}

func (s *stream) RegisterCloseCallback(cb func(error)) {
	s.closeCallback = append(s.closeCallback, cb)
}

func (s *stream) setRecvTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}
	s.recvTimeout = timeout
}

func (s *stream) runCloseCallback(exception error) {
	if len(s.closeCallback) > 0 {
		for _, cb := range s.closeCallback {
			cb(exception)
		}
	}
	_ = s.writer.CloseStream(s.sid)
}

func (s *stream) writeFrame(ftype int32, header streaming.Header, trailer streaming.Trailer, payload []byte) (err error) {
	fr := newFrame(streamFrame{sid: s.sid, method: s.method, header: header, trailer: trailer}, ftype, payload)
	return s.writer.WriteFrame(fr)
}

// writeTrailer send trailer to peer
// if exception is not nil, trailer frame should carry a payload
func (s *stream) sendTrailer(exception error) (err error) {
	// Check if trailer was already sent using flag
	if s.flags&streamFlagTrailerSent != 0 {
		return fmt.Errorf("stream trailer already sent")
	}

	// Mark as sent
	s.flags |= streamFlagTrailerSent

	wtrailer := s.wtrailer
	// Clear for GC (optional, helps reduce memory footprint)
	s.wtrailer = nil

	var payload []byte
	if exception != nil {
		payload, err = EncodeException(context.Background(), s.method, s.sid, exception)
		if err != nil {
			return err
		}
	}
	err = s.writeFrame(trailerFrameType, nil, wtrailer, payload)
	return err
}

func (s *stream) sendRst(exception error, cancelPath string) (err error) {
	var payload []byte
	if exception != nil {
		payload, err = EncodeException(context.Background(), s.method, s.sid, exception)
		if err != nil {
			return err
		}
	}
	var header streaming.Header
	if cancelPath != "" {
		header = make(streaming.Header)
		header[ttheader.HeaderTTStreamCancelPath] = cancelPath
	}
	return s.writeFrame(rstFrameType, header, nil, payload)
}

// === Frame OnRead callback

func (s *stream) onReadDataFrame(fr *Frame) (err error) {
	s.reader.input(context.Background(), fr.payload)
	return nil
}
