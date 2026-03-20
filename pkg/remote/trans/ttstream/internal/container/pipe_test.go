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

package container

import (
	"context"
	"errors"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/kitex/internal/test"
)

func TestPipeline(t *testing.T) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)
	var recv int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		items := make([]int, 10)
		for {
			n, err := pipe.Read(items)
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				return
			}
			for i := 0; i < n; i++ {
				recv += items[i]
			}
		}
	}()
	round := 10000
	itemsPerRound := []int{1, 1, 1, 1, 1}
	for i := 0; i < round; i++ {
		err := pipe.Write(itemsPerRound...)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("Pipe closing")
	pipe.Close()
	t.Logf("Pipe closed")
	wg.Wait()
	if recv != len(itemsPerRound)*round {
		t.Fatalf("Pipe expect %d items, got %d", len(itemsPerRound)*round, recv)
	}
}

func TestPipelineWriteCloseAndRead(t *testing.T) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)
	pipeSize := 10
	var pipeRead int32
	var readWG sync.WaitGroup
	readWG.Add(1)
	go func() {
		defer readWG.Done()
		readBuf := make([]int, 1)
		for {
			n, err := pipe.Read(readBuf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					t.Errorf("un except err: %v", err)
				}
				break
			}
			atomic.AddInt32(&pipeRead, int32(n))
		}
	}()
	time.Sleep(time.Millisecond * 10) // let read goroutine start first
	for i := 0; i < pipeSize; i++ {
		err := pipe.Write(i)
		if err != nil {
			t.Error(err)
		}
	}
	for atomic.LoadInt32(&pipeRead) != int32(pipeSize) {
		runtime.Gosched()
	}
	pipe.Close()
	readWG.Wait()
}

func TestPipelineCancelCallback(t *testing.T) {
	pipeWrapper := struct{ pipe *Pipe[int] }{}
	finalNum := 100
	firstNum := 1
	rCtx, cancel := context.WithCancel(context.Background())
	pipe := NewPipe[int](rCtx, func(ctx context.Context) error {
		_ = pipeWrapper.pipe.Write(finalNum)
		pipeWrapper.pipe.Close()
		return ctx.Err()
	})
	pipeWrapper.pipe = pipe
	var readWG sync.WaitGroup
	readWG.Add(1)
	go func() {
		defer readWG.Done()
		var canceled bool
		readBuf := make([]int, 1)
		for {
			n, rErr := pipe.Read(readBuf)
			if canceled {
				if rErr != nil {
					test.Assert(t, rErr == context.Canceled, rErr)
					return
				}
				test.Assert(t, n == 1, n)
				test.Assert(t, readBuf[0] == finalNum, readBuf[0])
				continue
			}
			test.Assert(t, n == 1, n)
			test.Assert(t, readBuf[0] == firstNum, readBuf[0])
			cancel()
			canceled = true
		}
	}()
	wErr := pipe.Write(firstNum)
	test.Assert(t, wErr == nil, wErr)
	readWG.Wait()
}

func BenchmarkPipeline(b *testing.B) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)
	readCache := make([]int, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(readCache); j++ {
			go pipe.Write(1)
		}
		total := 0
		for total < len(readCache) {
			n, _ := pipe.Read(readCache)
			total += n
		}
	}
}

// TestPipeReadCtx tests basic ReadCtx functionality
func TestPipeReadCtx(t *testing.T) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)

	// Write some data
	err := pipe.Write(1, 2, 3, 4, 5)
	test.Assert(t, err == nil, err)

	// Read with a separate context
	readCtx := context.Background()
	items := make([]int, 10)
	n, err := pipe.ReadCtx(readCtx, nil, items)
	test.Assert(t, err == nil, err)
	test.Assert(t, n == 5, n)
	test.Assert(t, items[0] == 1, items[0])
	test.Assert(t, items[4] == 5, items[4])

	pipe.Close()
}

// TestPipeReadCtxCancel tests ReadCtx behavior when perReadCtx is canceled
func TestPipeReadCtxCancel(t *testing.T) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)

	readCtx, cancel := context.WithCancel(context.Background())
	var readErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		items := make([]int, 10)
		// This should block and then return error when readCtx is canceled
		_, readErr = pipe.ReadCtx(readCtx, nil, items)
	}()

	// Give goroutine time to start reading
	time.Sleep(time.Millisecond * 50)

	// Cancel the read context
	cancel()

	wg.Wait()
	test.Assert(t, readErr != nil, "expected error when context canceled")
	test.Assert(t, errors.Is(readErr, context.Canceled), readErr)

	pipe.Close()
}

// TestPipeReadCtxWithCallback tests ReadCtx with a custom callback
func TestPipeReadCtxWithCallback(t *testing.T) {
	ctx := context.Background()
	pipe := NewPipe[int](ctx, nil)

	readCtx, cancel := context.WithCancel(context.Background())
	var callbackCalled bool
	var callbackCtx context.Context
	callback := func(ctx context.Context) error {
		callbackCalled = true
		callbackCtx = ctx
		return errors.New("custom callback error")
	}

	var readErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		items := make([]int, 10)
		_, readErr = pipe.ReadCtx(readCtx, callback, items)
	}()

	// Give goroutine time to start reading
	time.Sleep(time.Millisecond * 50)

	// Cancel the read context
	cancel()

	wg.Wait()
	test.Assert(t, callbackCalled, "callback should be called")
	test.Assert(t, callbackCtx == readCtx, "callback should receive readCtx")
	test.Assert(t, readErr != nil, "expected error from callback")
	test.Assert(t, readErr.Error() == "custom callback error", readErr)

	pipe.Close()
}

// TestPipeReadCtxPipeContextCancel tests ReadCtx when pipe's main context is canceled
func TestPipeReadCtxPipeContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var pipeCallbackCalled bool
	pipeCallback := func(ctx context.Context) error {
		pipeCallbackCalled = true
		return errors.New("pipe context canceled")
	}
	pipe := NewPipe[int](ctx, pipeCallback)

	readCtx := context.Background()
	var readErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		items := make([]int, 10)
		// Use ReadCtx with a different context
		_, readErr = pipe.ReadCtx(readCtx, nil, items)
	}()

	// Give goroutine time to start reading
	time.Sleep(time.Millisecond * 50)

	// Cancel the pipe's main context
	cancel()

	wg.Wait()
	test.Assert(t, pipeCallbackCalled, "pipe callback should be called")
	test.Assert(t, readErr != nil, "expected error from pipe callback")
	test.Assert(t, readErr.Error() == "pipe context canceled", readErr)
}

// TestPipeReadCtxBothContextsCancel tests ReadCtx when both contexts can be canceled
func TestPipeReadCtxBothContextsCancel(t *testing.T) {
	pipeCtx, _ := context.WithCancel(context.Background())
	pipe := NewPipe[int](pipeCtx, nil)

	readCtx, cancelRead := context.WithCancel(context.Background())
	var readErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		items := make([]int, 10)
		_, readErr = pipe.ReadCtx(readCtx, nil, items)
	}()

	// Give goroutine time to start reading
	time.Sleep(time.Millisecond * 50)

	// Cancel the read context (should take precedence if both are canceled)
	cancelRead()

	wg.Wait()
	test.Assert(t, readErr != nil, "expected error")
	test.Assert(t, errors.Is(readErr, context.Canceled), readErr)

	pipe.Close()
}
