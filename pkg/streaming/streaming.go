/*
 * Copyright 2021 CloudWeGo Authors
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

// Package streaming interface
package streaming

import (
	"context"
	"io"

	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
)

// Stream both client and server stream
//
// On client side, once RecvMsg or SendMsg returns a non-nil error, the stream should be considered
// finished and no further calls should be made on it.
// On server side, RecvMsg returning io.EOF only indicates the client-to-server direction is done;
// the server may still call SendMsg. After RecvMsg returns io.EOF, the caller should not make
// further RecvMsg calls, but SendMsg is not affected.
// After SendMsg returns a non-nil error or RecvMsg returns any error other than io.EOF,
// the stream is aborted.
// If subsequent calls are made regardless, they will return the same error.
//
// Deprecated: It's only for gRPC, use ClientStream or ServerStream instead.
type Stream interface {
	// SetHeader sets the header metadata. It may be called multiple times.
	// When call multiple times, all the provided metadata will be merged.
	// All the metadata will be sent out when one of the following happens:
	//  - ServerStream.SendHeader() is called;
	//  - The first response is sent out;
	//  - An RPC status is sent out (error or success).
	SetHeader(metadata.MD) error
	// SendHeader sends the header metadata.
	// The provided md and headers set by SetHeader() will be sent.
	// It fails if called multiple times.
	SendHeader(metadata.MD) error
	// SetTrailer sets the trailer metadata which will be sent with the RPC status.
	// When called more than once, all the provided metadata will be merged.
	SetTrailer(metadata.MD)
	// Header is used for client side stream to receive header from server.
	Header() (metadata.MD, error)
	// Trailer is used for client side stream to receive trailer from server.
	Trailer() metadata.MD
	// Context the stream context.Context
	Context() context.Context
	// RecvMsg receives a message from the peer.
	// It blocks until a message is received or an error occurs.
	// On client side, it returns io.EOF when the stream is done.
	// On server side, it returns io.EOF when the client-to-server direction is done (i.e. the client
	// called Close()); the server may still call SendMsg after that.
	// After RecvMsg returns a non-nil error (including io.EOF), the caller should not make further
	// RecvMsg calls. If subsequent calls are made regardless, they will return the same error.
	// For non-EOF errors, the stream is aborted entirely.
	// It is not concurrent-safe.
	RecvMsg(m interface{}) error
	// SendMsg sends a message to the peer.
	// It blocks until the message is sent or an error occurs.
	// After SendMsg returns a non-nil error, the stream is aborted and
	// subsequent calls will return the same error.
	// It is not concurrent-safe.
	SendMsg(m interface{}) error
	// not concurrent-safety with SendMsg
	io.Closer
}

// WithDoFinish should be implemented when:
// (1) you want to wrap a stream in client middleware, and
// (2) you want to manually call streaming.FinishStream(stream, error) to record the end of stream
// Note: the DoFinish should be reentrant, better with a sync.Once.
type WithDoFinish interface {
	DoFinish(error)
}

// CloseCallbackRegister register a callback when stream closed.
type CloseCallbackRegister interface {
	RegisterCloseCallback(cb func(error))
}

// Args endpoint request
type Args struct {
	ServerStream ServerStream
	ClientStream ClientStream
	// for gRPC compatible
	Stream Stream
}

// Result endpoint response
type Result struct {
	ServerStream ServerStream
	ClientStream ClientStream
	// for gRPC compatible
	Stream Stream
}

type GRPCStreamGetter interface {
	GetGRPCStream() Stream
}
