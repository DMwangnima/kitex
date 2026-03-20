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
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex/internal/test"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

func newTestClientStream(ctx context.Context) *clientStream {
	return newClientStreamWithWriter(ctx, mockStreamWriter{})
}

func newClientStreamWithWriter(ctx context.Context, writer streamWriter) *clientStream {
	cs := newClientStream(ctx, writer, streamFrame{method: "testMethod"})
	cs.rpcInfo = rpcinfo.NewRPCInfo(
		rpcinfo.NewEndpointInfo("testService", "testMethod", nil, nil), nil, nil, nil, nil)
	return cs
}

func Test_clientStreamStateChange(t *testing.T) {
	t.Run("serverStream close, then RecvMsg/SendMsg returning exception", func(t *testing.T) {
		cliSt := newTestClientStream(context.Background())
		testException := errors.New("test")
		cliSt.close(testException, false, "", nil)
		rErr := cliSt.RecvMsg(cliSt.ctx, nil)
		test.Assert(t, rErr == testException, rErr)
		sErr := cliSt.SendMsg(cliSt.ctx, nil)
		test.Assert(t, sErr == testException, sErr)
	})
	t.Run("serverStream close twice with different exception, RecvMsg/SendMsg returning the first time exception", func(t *testing.T) {
		cliSt := newTestClientStream(context.Background())
		testException1 := errors.New("test1")
		cliSt.close(testException1, false, "", nil)
		testException2 := errors.New("test2")
		cliSt.close(testException2, false, "", nil)

		rErr := cliSt.RecvMsg(cliSt.ctx, nil)
		test.Assert(t, rErr == testException1, rErr)
		sErr := cliSt.SendMsg(cliSt.ctx, nil)
		test.Assert(t, sErr == testException1, sErr)
	})
	t.Run("serverStream close concurrently with RecvMsg/SendMsg", func(t *testing.T) {
		cliSt := newTestClientStream(context.Background())
		var wg sync.WaitGroup
		wg.Add(2)
		testException := errors.New("test")
		go func() {
			time.Sleep(100 * time.Millisecond)
			cliSt.close(testException, false, "", nil)
		}()
		go func() {
			defer wg.Done()
			rErr := cliSt.RecvMsg(cliSt.ctx, nil)
			test.Assert(t, rErr == testException, rErr)
		}()
		go func() {
			defer wg.Done()
			for {
				sErr := cliSt.SendMsg(cliSt.ctx, &testRequest{})
				if sErr == nil {
					continue
				}
				test.Assert(t, sErr == testException, sErr)
				break
			}
		}()
		wg.Wait()
	})
}

func Test_clientStream_parseStreamCtxErr(t *testing.T) {
	testcases := []struct {
		desc             string
		ctxFunc          func() (context.Context, context.CancelFunc)
		waitCtxDone      bool // if true, wait for ctx.Done() instead of calling cancel()
		expectEx         *Exception
		expectNoCancel   bool
		expectCancelPath string
	}{
		{
			desc: "invoke cancel() actively",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx := rpcinfo.NewCtxWithRPCInfo(context.Background(), rpcinfo.NewRPCInfo(
					rpcinfo.NewEndpointInfo("serviceA", "testMethod", nil, nil), nil, nil, rpcinfo.NewRPCConfig(), nil))
				return context.WithCancel(ctx)
			},
			expectEx:         errBizCancel.newBuilder().withSide(clientSide),
			expectCancelPath: "serviceA",
		},
		{
			desc: "cascading cancel",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx := rpcinfo.NewCtxWithRPCInfo(context.Background(), rpcinfo.NewRPCInfo(
					rpcinfo.NewEndpointInfo("serviceB", "testMethod", nil, nil), nil, nil, rpcinfo.NewRPCConfig(), nil))
				ctx, cancel := context.WithCancel(ctx)
				ctx, cancelFunc := newContextWithCancelReason(ctx, cancel)
				cause := defaultRstException
				return ctx, func() {
					cancelFunc(errUpstreamCancel.newBuilder().withSide(serverSide).setOrAppendCancelPath("serviceA").withCauseAndTypeId(cause, cause.TypeId()))
				}
			},
			expectEx:         errUpstreamCancel.newBuilder().withSide(clientSide).setOrAppendCancelPath("serviceA").withCauseAndTypeId(defaultRstException, defaultRstException.TypeId()),
			expectCancelPath: "serviceA,serviceB",
		},
		{
			desc: "stream ctx timeout",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx := rpcinfo.NewCtxWithRPCInfo(context.Background(), rpcinfo.NewRPCInfo(
					rpcinfo.NewEndpointInfo("serviceA", "testMethod", nil, nil), nil, nil, rpcinfo.NewRPCConfig(), nil))
				return context.WithTimeout(ctx, 100*time.Millisecond)
			},
			waitCtxDone:      true,
			expectEx:         newStreamTimeoutException(0),
			expectCancelPath: "serviceA",
		},
		{
			desc: "other customized ctx Err",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx := rpcinfo.NewCtxWithRPCInfo(context.Background(), rpcinfo.NewRPCInfo(
					rpcinfo.NewEndpointInfo("serviceA", "testMethod", nil, nil), nil, nil, rpcinfo.NewRPCConfig(), nil))
				ctx, cancel := context.WithCancel(ctx)
				ctx, cancelFunc := newContextWithCancelReason(ctx, cancel)
				return ctx, func() {
					cancelFunc(errors.New("test"))
				}
			},
			expectEx:         errInternalCancel.newBuilder().withSide(clientSide).withCause(errors.New("test")),
			expectCancelPath: "non-ttstream path,serviceA",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := tc.ctxFunc()
			cs := newClientStream(ctx, mockStreamWriter{}, streamFrame{})
			if tc.waitCtxDone {
				<-ctx.Done()
			} else {
				cancel()
			}
			finalEx, noCancel, cancelPath := cs.parseStreamCtxErr(ctx)
			test.DeepEqual(t, finalEx, tc.expectEx)
			test.Assert(t, tc.expectCancelPath == cancelPath, cancelPath)
			test.Assert(t, tc.expectNoCancel == noCancel, noCancel)
		})
	}
}

func Test_clientStream_SendMsg(t *testing.T) {
	ctx := context.Background()
	cs := newClientStream(ctx, &mockStreamWriter{}, streamFrame{})
	req := &testRequest{B: "SendMsgTest"}

	// Send successfully
	err := cs.SendMsg(ctx, req)
	test.Assert(t, err == nil, err)
	// CloseSend
	err = cs.CloseSend(ctx)
	test.Assert(t, err == nil, err)
	err = cs.SendMsg(ctx, req)
	test.Assert(t, err != nil)
	test.Assert(t, errors.Is(err, errIllegalOperation), err)
	test.Assert(t, strings.Contains(err.Error(), "stream is closed send"))

	cs = newClientStream(ctx, &mockStreamWriter{}, streamFrame{})
	// Send retrieves the close stream exception
	ex := errDownstreamCancel.newBuilder().withSide(clientSide)
	cs.close(ex, false, "", nil)
	err = cs.SendMsg(ctx, req)
	test.Assert(t, errors.Is(err, errDownstreamCancel), err)
	// subsequent close will not affect the error returned by Send
	ex1 := errApplicationException.newBuilder().withSide(clientSide)
	cs.close(ex1, false, "", nil)
	err = cs.SendMsg(ctx, req)
	test.Assert(t, errors.Is(err, errDownstreamCancel), err)
}

// Test_CancelWithErr_concurrent tests that CancelWithErr (outer layer) racing with
// ctxDoneCallback (pipe stream ctx) is safe. This simulates the dual-timer scenario:
// outer callWithTimeout fires at the same moment as stream ctx deadline.
func Test_CancelWithErr_concurrent(t *testing.T) {
	t.Run("CancelWithErr races with ctxDoneCallback", func(t *testing.T) {
		// Run multiple iterations to increase chance of hitting the race window
		for i := 0; i < 100; i++ {
			var rstCount int32
			writer := mockStreamWriter{
				writeFrameFunc: func(f *Frame) error {
					if f.typ == rstFrameType {
						atomic.AddInt32(&rstCount, 1)
					}
					return nil
				},
			}
			// stream ctx with a very short deadline
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			cs := newClientStreamWithWriter(ctx, writer)
			cs.setStreamTimeout(10 * time.Millisecond)

			outerErr := errBizCancel.newBuilder().withSide(clientSide)

			var wg sync.WaitGroup
			wg.Add(2)

			// goroutine 1: simulate outer layer calling CancelWithErr on timeout
			go func() {
				defer wg.Done()
				<-ctx.Done() // wait for ctx to expire, then race
				cs.CancelWithErr(outerErr)
			}()

			// goroutine 2: RecvMsg blocks in pipe, ctxDoneCallback fires on ctx expiry
			go func() {
				defer wg.Done()
				_ = cs.RecvMsg(cs.ctx, nil)
			}()

			wg.Wait()
			cancel()

			test.Assert(t, atomic.LoadInt32(&cs.state) == streamStateInactive)
			ex := cs.closeStreamException.Load()
			test.Assert(t, ex != nil)
			test.Assert(t, atomic.LoadInt32(&rstCount) == 1, atomic.LoadInt32(&rstCount))
		}
	})
	t.Run("concurrent CancelWithErr calls", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			writer := mockStreamWriter{
				writeFrameFunc: func(f *Frame) error { return nil },
			}
			cs := newClientStreamWithWriter(context.Background(), writer)

			ex1 := errBizCancel.newBuilder().withSide(clientSide)
			ex2 := errDownstreamCancel.newBuilder().withSide(clientSide)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				cs.CancelWithErr(ex1)
			}()
			go func() {
				defer wg.Done()
				cs.CancelWithErr(ex2)
			}()
			wg.Wait()

			test.Assert(t, atomic.LoadInt32(&cs.state) == streamStateInactive)
			stored := cs.closeStreamException.Load()
			test.Assert(t, stored != nil)
			storedErr := stored.(error)
			test.Assert(t, errors.Is(storedErr, errBizCancel) || errors.Is(storedErr, errDownstreamCancel), storedErr)
		}
	})
}

func Test_recvTimeout(t *testing.T) {
	t.Run("recv timeout closes stream and sends RST", func(t *testing.T) {
		var rstFrameCount int32
		closeStreamCh := make(chan struct{})
		writer := mockStreamWriter{
			writeFrameFunc: func(f *Frame) error {
				if f.typ == rstFrameType {
					atomic.AddInt32(&rstFrameCount, 1)
				}
				return nil
			},
			closeStreamFunc: func(sid int32) error {
				close(closeStreamCh)
				return nil
			},
		}
		cs := newClientStreamWithWriter(context.Background(), writer)
		cs.setRecvTimeout(50 * time.Millisecond)

		rErr := cs.RecvMsg(cs.ctx, nil)
		test.Assert(t, rErr != nil, rErr)
		ex, ok := rErr.(*Exception)
		test.Assert(t, ok, rErr)
		test.Assert(t, ex.TypeId() == errRecvTimeoutTypeId, ex)
		test.Assert(t, errors.Is(rErr, kerrors.ErrStreamingTimeout), rErr)
		t.Log(rErr)

		select {
		case <-closeStreamCh:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CloseStream")
		}
		test.Assert(t, atomic.LoadInt32(&rstFrameCount) == 1, atomic.LoadInt32(&rstFrameCount))
	})
	t.Run("stream ctx expires before recv timeout, returns stream timeout error", func(t *testing.T) {
		var rstFrameCount int32
		closeStreamCh := make(chan struct{})
		writer := mockStreamWriter{
			writeFrameFunc: func(f *Frame) error {
				if f.typ == rstFrameType {
					atomic.AddInt32(&rstFrameCount, 1)
				}
				return nil
			},
			closeStreamFunc: func(sid int32) error {
				close(closeStreamCh)
				return nil
			},
		}
		// stream ctx expires in 50ms, recv timeout is 300ms
		streamTimeout := 50 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), streamTimeout)
		defer cancel()
		cs := newClientStreamWithWriter(ctx, writer)
		cs.setStreamTimeout(streamTimeout)
		cs.setRecvTimeout(300 * time.Millisecond)

		// RecvMsg should block until stream ctx expires (50ms), NOT until recv timeout (300ms)
		start := time.Now()
		rErr := cs.RecvMsg(cs.ctx, nil)
		elapsed := time.Since(start)
		test.Assert(t, rErr != nil, rErr)
		ex, ok := rErr.(*Exception)
		test.Assert(t, ok, rErr)
		// must be stream timeout (errStreamTimeoutTypeId), NOT recv timeout (errRecvTimeoutTypeId)
		test.Assert(t, ex.TypeId() == errStreamTimeoutTypeId, ex.TypeId())
		t.Log(rErr)

		// should return around 50ms, not 300ms
		test.Assert(t, elapsed < 200*time.Millisecond, elapsed)

		select {
		case <-closeStreamCh:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CloseStream")
		}
		test.Assert(t, atomic.LoadInt32(&rstFrameCount) > 0, atomic.LoadInt32(&rstFrameCount))
	})
	t.Run("recv timeout with stream ctx, recv timeout fires first", func(t *testing.T) {
		var rstFrameCount int32
		closeStreamCh := make(chan struct{})
		writer := mockStreamWriter{
			writeFrameFunc: func(f *Frame) error {
				if f.typ == rstFrameType {
					atomic.AddInt32(&rstFrameCount, 1)
				}
				return nil
			},
			closeStreamFunc: func(sid int32) error {
				close(closeStreamCh)
				return nil
			},
		}
		// stream ctx expires in 300ms, recv timeout is 50ms
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		cs := newClientStreamWithWriter(ctx, writer)
		cs.setRecvTimeout(50 * time.Millisecond)

		rErr := cs.RecvMsg(cs.ctx, nil)
		test.Assert(t, rErr != nil, rErr)
		ex, ok := rErr.(*Exception)
		test.Assert(t, ok, rErr)
		test.Assert(t, ex.TypeId() == errRecvTimeoutTypeId, ex)
		test.Assert(t, errors.Is(rErr, kerrors.ErrStreamingTimeout), rErr)
		t.Log(rErr)

		select {
		case <-closeStreamCh:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for CloseStream")
		}
		test.Assert(t, atomic.LoadInt32(&rstFrameCount) > 0, atomic.LoadInt32(&rstFrameCount))
	})
}
