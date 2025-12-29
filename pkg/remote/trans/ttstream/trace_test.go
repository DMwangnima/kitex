//go:build !windows

/*
 * Copyright 2025 CloudWeGo Authors
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
	"io"
	"sync"
	"testing"

	"github.com/cloudwego/kitex/internal/test"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/streaming"
)

type mockStreamTracer struct {
	t        *testing.T
	eventBuf []stats.Event
	mu       sync.Mutex
	finishCh chan struct{}
}

func newMockStreamTracer() *mockStreamTracer {
	return &mockStreamTracer{
		finishCh: make(chan struct{}, 1),
	}
}

func (tracer *mockStreamTracer) Start(ctx context.Context) context.Context {
	return ctx
}

func (tracer *mockStreamTracer) Finish(ctx context.Context) {
	tracer.finishCh <- struct{}{}
}

func (tracer *mockStreamTracer) SetT(t *testing.T) {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	tracer.t = t
}

func (tracer *mockStreamTracer) ReportStreamEvent(ctx context.Context, ri rpcinfo.RPCInfo, event rpcinfo.Event) {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	tracer.eventBuf = append(tracer.eventBuf, event.Event())
}

func (tracer *mockStreamTracer) Verify(expects ...stats.Event) {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	t := tracer.t
	test.Assert(t, len(tracer.eventBuf) == len(expects), tracer.eventBuf)
	for i, e := range expects {
		test.Assert(t, e.Index() == tracer.eventBuf[i].Index(), tracer.eventBuf)
	}
}

func (tracer *mockStreamTracer) Clean() {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	// wait for Finish invoked
	<-tracer.finishCh
	tracer.eventBuf = tracer.eventBuf[:0]
}

func TestTrace_trace(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		t.Run("clientStreaming", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)
			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)
			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "ClientStreaming"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})
			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			var wg sync.WaitGroup
			wg.Add(1)
			// client
			go func() {
				defer wg.Done()
				req := new(testRequest)
				req.B = "hello"
				for i := 0; i < 3; i++ {
					sErr := cs.SendMsg(context.Background(), req)
					test.Assert(t, sErr == nil, sErr)
				}
				sErr := cs.CloseSend(context.Background())
				test.Assert(t, sErr == nil, sErr)

				res := new(testResponse)
				rErr := cs.RecvMsg(context.Background(), res)
				test.Assert(t, rErr == nil, rErr)
			}()

			// server
			req := new(testRequest)
			for i := 0; i < 3; i++ {
				rErr := ss.RecvMsg(context.Background(), req)
				test.Assert(t, rErr == nil, rErr)
			}
			rErr := ss.RecvMsg(context.Background(), req)
			test.Assert(t, rErr == io.EOF, rErr)
			res := new(testResponse)
			res.B = req.B
			sErr := ss.SendMsg(context.Background(), res)
			test.Assert(t, sErr == nil, sErr)
			sErr = ss.CloseSend(nil)
			test.Assert(t, sErr == nil, sErr)

			wg.Wait()
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("serverStreaming", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "ServerStreaming"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			var wg sync.WaitGroup
			wg.Add(1)
			// client
			go func() {
				defer wg.Done()
				req := new(testRequest)
				req.B = "hello"
				err := cs.SendMsg(context.Background(), req)
				test.Assert(t, err == nil, err)

				err = cs.CloseSend(context.Background())
				test.Assert(t, err == nil, err)

				// Receive multiple responses
				res := new(testResponse)
				for {
					err = cs.RecvMsg(context.Background(), res)
					if err != nil {
						test.Assert(t, err == io.EOF, err)
						break
					}
				}
			}()

			// server
			req := new(testRequest)
			err = ss.RecvMsg(context.Background(), req)
			test.Assert(t, err == nil, err)
			err = ss.RecvMsg(context.Background(), req)
			test.Assert(t, err == io.EOF, err)

			// Send multiple responses
			res := new(testResponse)
			res.B = req.B
			for i := 0; i < 3; i++ {
				err = ss.SendMsg(context.Background(), res)
				test.Assert(t, err == nil, err)
			}
			err = ss.CloseSend(nil)
			test.Assert(t, err == nil, err)

			wg.Wait()

			// Verify trace events
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("bidiStreaming", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Bidi"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			var wg sync.WaitGroup
			wg.Add(2)
			// client
			go func() {
				defer wg.Done()
				req := new(testRequest)
				req.B = "hello"
				err := cs.SendMsg(context.Background(), req)
				test.Assert(t, err == nil, err)

				err = cs.CloseSend(context.Background())
				test.Assert(t, err == nil, err)
			}()

			// client receive
			go func() {
				defer wg.Done()
				res := new(testResponse)
				err := cs.RecvMsg(context.Background(), res)
				test.Assert(t, err == nil, err)
			}()

			// server
			req := new(testRequest)
			err = ss.RecvMsg(context.Background(), req)
			test.Assert(t, err == nil, err)
			err = ss.RecvMsg(context.Background(), req)
			test.Assert(t, err == io.EOF, err)
			res := new(testResponse)
			res.B = req.B
			err = ss.SendMsg(context.Background(), res)
			test.Assert(t, err == nil, err)
			err = ss.CloseSend(nil)
			test.Assert(t, err == nil, err)

			wg.Wait()

			// Verify trace events
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
	})
	t.Run("return handler", func(t *testing.T) {
		t.Run("server returns immediately with no data", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			// Server closes immediately with no data
			err = ss.CloseSend(nil)
			test.Assert(t, err == nil, err)

			// Client receives EOF
			res := new(testResponse)
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err == io.EOF, err)

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("server sends header then closes", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			// Server sends header
			header := streaming.Header{"key": "value"}
			err = ss.SetHeader(header)
			test.Assert(t, err == nil, err)

			// Server closes without sending data
			err = ss.CloseSend(nil)
			test.Assert(t, err == nil, err)

			// Client receives EOF
			res := new(testResponse)
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err == io.EOF, err)

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("server sends partial data then closes", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx := context.Background()
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			// Server sends one message
			res := new(testResponse)
			res.B = "partial response"
			err = ss.SendMsg(context.Background(), res)
			test.Assert(t, err == nil, err)

			// Client receives the message
			clientRes := new(testResponse)
			err = cs.RecvMsg(context.Background(), clientRes)
			test.Assert(t, err == nil, err)
			test.Assert(t, clientRes.B == "partial response", clientRes.B)

			// Server closes after sending partial data
			err = ss.CloseSend(nil)
			test.Assert(t, err == nil, err)

			// Client receives EOF on next recv
			err = cs.RecvMsg(context.Background(), clientRes)
			test.Assert(t, err == io.EOF, err)

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
	})
	t.Run("cancel", func(t *testing.T) {
		t.Run("cancel after stream creation", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx, cancel := context.WithCancel(context.Background())
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			// Initialize rpcInfo for cancel handling
			cs.rpcInfo = rpcinfo.NewRPCInfo(
				rpcinfo.NewEndpointInfo("client", "Test", nil, nil), nil, nil, nil, nil)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			// Cancel immediately after stream creation
			cancel()

			// Wait for cancellation to propagate
			res := new(testResponse)
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err != nil, err)

			// Server should receive RST frame
			req := new(testRequest)
			err = ss.RecvMsg(context.Background(), req)
			test.Assert(t, err != nil, err)

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("cancel after sending data", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx, cancel := context.WithCancel(context.Background())
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			// Initialize rpcInfo for cancel handling
			cs.rpcInfo = rpcinfo.NewRPCInfo(
				rpcinfo.NewEndpointInfo("client", "Test", nil, nil), nil, nil, nil, nil)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			// Send one message
			req := new(testRequest)
			req.B = "hello"
			err = cs.SendMsg(context.Background(), req)
			test.Assert(t, err == nil, err)

			// Server receives the message
			serverReq := new(testRequest)
			err = ss.RecvMsg(context.Background(), serverReq)
			test.Assert(t, err == nil, err)

			// Cancel after sending data
			cancel()

			// Wait for cancellation to propagate
			res := new(testResponse)
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err != nil, err)

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
		t.Run("cancel after receiving data", func(t *testing.T) {
			cliTraceCtl := &rpcinfo.TraceController{}
			cliTracer := newMockStreamTracer()
			cliTraceCtl.Append(cliTracer)
			cliTracer.SetT(t)
			defer cliTracer.Clean()

			cconn, sconn := newTestConnectionPipe(t)

			intHeader := make(IntHeader)
			strHeader := make(streaming.Header)

			ctrans := newClientTransport(cconn, nil)
			defer ctrans.Close(nil)
			ctx, cancel := context.WithCancel(context.Background())
			cs := newClientStream(ctx, ctrans, streamFrame{sid: genStreamID(), method: "Test"})
			cs.setTraceController(cliTraceCtl)
			// Initialize rpcInfo for cancel handling
			cs.rpcInfo = rpcinfo.NewRPCInfo(
				rpcinfo.NewEndpointInfo("client", "Test", nil, nil), nil, nil, nil, nil)
			cs.RegisterCloseCallback(func(err error) {
				cliTracer.Finish(cs.ctx)
			})

			err := ctrans.WriteStream(ctx, cs, intHeader, strHeader)
			test.Assert(t, err == nil, err)

			// Manually report StreamStart
			cliTraceCtl.ReportStreamEvent(ctx, stats.StreamStart, nil)

			strans := newServerTransport(sconn)
			ss, err := strans.ReadStream(context.Background())
			test.Assert(t, err == nil, err)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Server sends one response
				res := new(testResponse)
				res.B = "response"
				err := ss.SendMsg(context.Background(), res)
				test.Assert(t, err == nil, err)
			}()

			// Client receives one message
			res := new(testResponse)
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err == nil, err)

			// Cancel after receiving data
			cancel()

			// Try to receive again, should fail
			err = cs.RecvMsg(context.Background(), res)
			test.Assert(t, err != nil, err)

			wg.Wait()

			// Verify trace events: StreamStart -> StreamFinish
			cliTracer.Verify(stats.StreamStart, stats.StreamFinish)
		})
	})
}
