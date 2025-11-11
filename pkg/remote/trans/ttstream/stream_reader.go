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
	"errors"
	"io"
	"sync"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote/trans/ttstream/container"
)

var streamReaderPool sync.Pool

// streamReader is an abstraction layer for stream level IO operations
type streamReader struct {
	pipe      *container.Pipe[streamMsg]
	cache     [1]streamMsg
	exception error // once has exception, the stream should not work normally again
}

type streamMsg struct {
	payload   []byte
	exception error
}

func newStreamReader() *streamReader {
	var sio *streamReader
	if v := streamReaderPool.Get(); v != nil {
		sio = v.(*streamReader)
		sio.exception = nil // reset state
	} else {
		sio = new(streamReader)
	}
	sio.pipe = container.NewPipe[streamMsg]()
	return sio
}

func newStreamReaderWithCtxDoneCallback(callback container.CtxDoneCallback) *streamReader {
	var sio *streamReader
	if v := streamReaderPool.Get(); v != nil {
		sio = v.(*streamReader)
		sio.exception = nil // reset state
	} else {
		sio = new(streamReader)
	}
	sio.pipe = container.NewPipe[streamMsg](container.WithCtxDoneCallback(callback))
	return sio
}

func recycleStreamReader(sio *streamReader) {
	if sio == nil {
		return
	}
	// Clear references to help GC
	sio.pipe = nil
	sio.exception = nil
	streamReaderPool.Put(sio)
}

func (s *streamReader) input(ctx context.Context, payload []byte) {
	err := s.pipe.Write(ctx, streamMsg{payload: payload})
	if err != nil {
		klog.Errorf("stream pipe input failed: %v", err)
	}
}

// output would return err in the following scenarios:
// - pipe finished: container.ErrPipeEOF, container.ErrPipeCanceled
// - ctx Done() triggered: ctx.Err()
// - trailer frame contains err: streamMsg.exception
func (s *streamReader) output(ctx context.Context) (payload []byte, err error) {
	if s.exception != nil {
		return nil, s.exception
	}

	n, err := s.pipe.Read(ctx, s.cache[:])
	if err != nil {
		if errors.Is(err, container.ErrPipeEOF) {
			err = io.EOF
		}
		s.exception = err
		return nil, s.exception
	}
	if n == 0 {
		s.exception = io.EOF
		return nil, s.exception
	}
	msg := s.cache[0]
	if msg.exception != nil {
		s.exception = msg.exception
		return nil, s.exception
	}
	return msg.payload, nil
}

func (s *streamReader) close(exception error) {
	if exception != nil {
		_ = s.pipe.Write(context.Background(), streamMsg{exception: exception})
	}
	s.pipe.Close()
	// Clear callback to avoid closure captures preventing GC (similar to grpc fix #1886)
	s.pipe.ClearCallback()
	// Note: We don't recycle streamReader here because the stream may still hold a reference
	// The streamReader will be GC'd when the stream is GC'd
}
