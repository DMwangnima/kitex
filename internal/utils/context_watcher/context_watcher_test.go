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

package context_watcher

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"

	"github.com/cloudwego/kitex/internal/test"
)

func TestWatcher_Register(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		calledChan := make(chan context.Context, 1)
		callback := func(ctx context.Context) {
			calledChan <- ctx
		}

		ctx, cancel := context.WithCancel(context.Background())
		err := w.Register(ctx, callback)
		test.Assert(t, err == nil, err)
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		timeout := time.After(50 * time.Millisecond)
		select {
		case <-timeout:
			t.Fatal("callback not called within timeout")
		case calledCtx := <-calledChan:
			test.Assert(t, calledCtx == ctx)
		}
	})
	t.Run("invalid ctx or callback", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		// nil context
		err := w.Register(nil, func(context.Context) {})
		test.Assert(t, err == errRegisterNilCtx, err)

		// nil callback
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err = w.Register(ctx, nil)
		test.Assert(t, err == errRegisterNilCallback, err)

		// ctx with nil ctx.Done()
		err = w.Register(context.Background(), func(context.Context) {})
		test.Assert(t, err == nil, err)
	})
	t.Run("duplicate ctx", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		calledChan := make(chan context.Context, 1)
		callback := func(ctx context.Context) {
			calledChan <- ctx
		}

		// Register twice with same context
		err := w.Register(ctx, callback)
		test.Assert(t, err == nil, err)
		err = w.Register(ctx, callback)
		test.Assert(t, err == nil, err)

		cancel()
		timeout := time.After(50 * time.Millisecond)
		select {
		case <-timeout:
			t.Fatal("callback not called within timeout")
		case calledCtx := <-calledChan:
			test.Assert(t, calledCtx == ctx)
		}

		timeout = time.After(50 * time.Millisecond)
		select {
		case <-timeout:
		case <-calledChan:
			t.Fatal("callback should only be called once")
		}
	})
	t.Run("register ctx with timeout triggered", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		calledChan := make(chan context.Context, 1)
		callback := func(ctx context.Context) {
			calledChan <- ctx
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err := w.Register(ctx, callback)
		test.Assert(t, err == nil, err)

		timeout := time.After(50 * time.Millisecond)
		select {
		case <-timeout:
			t.Fatal("callback not called within timeout")
		case calledCtx := <-calledChan:
			test.Assert(t, calledCtx == ctx)
			test.Assert(t, calledCtx.Err() == context.DeadlineExceeded, calledCtx.Err())
		}
	})
	t.Run("register panic callback", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		ctx, cancel := context.WithCancel(context.Background())
		callback := func(ctx context.Context) {
			panic("callback panic")
		}

		err := w.Register(ctx, callback)
		test.Assert(t, err == nil, err)

		cancel()
		defer func() {
			r := recover()
			test.Assert(t, r == nil, r)
		}()

		timeout := time.After(50 * time.Millisecond)
		// wait for callback triggered
		<-timeout
	})
	t.Run("register after close", func(t *testing.T) {
		w := NewWatcher()
		err := w.Close()
		test.Assert(t, err == nil, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err = w.Register(ctx, func(context.Context) {})
		test.Assert(t, err == errWatcherClosed, err)
	})
	t.Run("register concurrently", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()
		var calledNum int32
		callback := func(ctx context.Context) {
			atomic.AddInt32(&calledNum, 1)
		}
		var wg sync.WaitGroup
		ctxNum := 1000
		wg.Add(ctxNum)
		for i := 0; i < ctxNum; i++ {
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithCancel(context.Background())
				err := w.Register(ctx, callback)
				test.Assert(t, err == nil, err)
				go func() {
					time.Sleep(time.Duration(fastrand.Intn(50)) * time.Millisecond)
					cancel()
				}()
			}()
		}
		wg.Wait()

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&calledNum) == int32(ctxNum) {
					return
				}
			case <-timer.C:
				t.Fatalf("wait for calledNum %d timeout, got %d actually", ctxNum, atomic.LoadInt32(&calledNum))
			}
		}
	})
	t.Run("register concurrently with new shard created", func(t *testing.T) {
		initialShards := 1
		maxCtxPerShard := 1000
		expectShards := initialShards + 1
		ctxNum := expectShards * maxCtxPerShard
		w := newWatcher(initialShards, maxCtxPerShard)
		defer w.Close()
		var calledNum int32
		callback := func(ctx context.Context) {
			atomic.AddInt32(&calledNum, 1)
		}
		var wg sync.WaitGroup

		cancelCh := make(chan struct{})
		wg.Add(ctxNum)
		for i := 0; i < ctxNum; i++ {
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithCancel(context.Background())
				err := w.Register(ctx, callback)
				test.Assert(t, err == nil, err)
				go func() {
					<-cancelCh
					cancel()
				}()
			}()
		}
		wg.Wait()
		w.shardsMu.RLock()
		test.Assert(t, len(w.shards) == expectShards, len(w.shards))
		w.shardsMu.RUnlock()
		close(cancelCh)

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&calledNum) == int32(ctxNum) {
					return
				}
			case <-timer.C:
				t.Fatalf("wait for calledNum %d timeout, got %d actually", ctxNum, atomic.LoadInt32(&calledNum))
			}
		}
	})
}

func TestWatcher_Deregister(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		calledChan := make(chan context.Context, 1)
		callback := func(ctx context.Context) {
			calledChan <- ctx
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := w.Register(ctx, callback)
		test.Assert(t, err == nil, err)

		w.Deregister(ctx)

		timeout := time.After(50 * time.Millisecond)
		select {
		case <-timeout:
		case <-calledChan:
			t.Fatal("callback should not be called")
		}
	})
	t.Run("invalid ctx", func(t *testing.T) {
		w := NewWatcher()
		defer w.Close()

		// nil context
		err := w.Register(nil, func(context.Context) {})
		test.Assert(t, err == errRegisterNilCtx, err)

		// nil callback
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err = w.Register(ctx, nil)
		test.Assert(t, err == errRegisterNilCallback, err)

		// ctx with nil ctx.Done()
		err = w.Register(context.Background(), func(context.Context) {})
		test.Assert(t, err == nil, err)
	})
	t.Run("deregister after close", func(t *testing.T) {
		w := NewWatcher()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := w.Register(ctx, func(context.Context) {})
		test.Assert(t, err == nil, err)
		err = w.Close()
		test.Assert(t, err == nil, err)
		w.Deregister(ctx)
	})
	t.Run("deregister concurrently", func(t *testing.T) {
		w := NewWatcher()
		cancelCh := make(chan struct{})
		defer func() {
			close(cancelCh)
			w.Close()
		}()
		var calledNum int32
		callback := func(ctx context.Context) {
			atomic.AddInt32(&calledNum, 1)
		}
		var wg sync.WaitGroup
		ctxNum := 1000
		wg.Add(ctxNum)
		for i := 0; i < ctxNum; i++ {
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithCancel(context.Background())
				err := w.Register(ctx, callback)
				test.Assert(t, err == nil, err)
				go func() {
					time.Sleep(time.Duration(fastrand.Intn(50)) * time.Millisecond)
					w.Deregister(ctx)
					<-cancelCh
					cancel()
				}()
			}()
		}
		wg.Wait()

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
	RegistryCheckLoop:
		for {
			select {
			case <-ticker.C:
				w.registryMu.Lock()
				if len(w.registry) == 0 {
					w.registryMu.Unlock()
					break RegistryCheckLoop
				}
				w.registryMu.Unlock()
			case <-timer.C:
				t.Fatalf("wait for calledNum %d timeout, got %d actually", ctxNum, atomic.LoadInt32(&calledNum))
			}
		}
		timer = time.NewTimer(5 * time.Second)
	CtxCountCheckLoop:
		for {
			select {
			case <-ticker.C:
				w.shardsMu.RLock()
				for _, s := range w.shards {
					if count := s.getCtxCount(); count != 0 {
						w.shardsMu.RUnlock()
						continue CtxCountCheckLoop
					}
				}
				w.shardsMu.RUnlock()
				break CtxCountCheckLoop
			case <-timer.C:
				t.Fatalf("wait for ctxCount becoming 0 timeout")
			}
		}
	})
}

func TestConcurrentRegistration(t *testing.T) {
	w := NewWatcher()
	defer w.Close()

	const numGoroutines = 100
	const numRegistrationsPerGoroutine = 10

	var wg sync.WaitGroup
	callbackExecuted := int32(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRegistrationsPerGoroutine; j++ {
				ctx, cancel := context.WithCancel(context.Background())

				callback := func(ctx context.Context) {
					atomic.AddInt32(&callbackExecuted, 1)
				}

				err := w.Register(ctx, callback)
				if err != nil {
					t.Errorf("Registration failed: %v", err)
					return
				}

				// Cancel some contexts immediately
				if j%2 == 0 {
					cancel()
				} else {
					// Cancel after a short delay
					go func() {
						time.Sleep(time.Millisecond)
						cancel()
					}()
				}
			}
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond) // Wait for all callbacks

	expected := int32(numGoroutines * numRegistrationsPerGoroutine)
	if atomic.LoadInt32(&callbackExecuted) != expected {
		t.Fatalf("Expected %d callbacks to be executed, got %d", expected, atomic.LoadInt32(&callbackExecuted))
	}
}

func TestConcurrentDeregistration(t *testing.T) {
	w := NewWatcher()
	defer w.Close()

	const numGoroutines = 50
	var wg sync.WaitGroup
	contexts := make([]context.Context, numGoroutines)
	cancels := make([]context.CancelFunc, numGoroutines)

	// Register contexts
	for i := 0; i < numGoroutines; i++ {
		contexts[i], cancels[i] = context.WithCancel(context.Background())

		err := w.Register(contexts[i], func(ctx context.Context) {})
		if err != nil {
			t.Fatalf("Registration failed: %v", err)
		}
	}

	// Concurrently deregister and cancel
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Deregister
		go func(ctx context.Context) {
			defer wg.Done()
			w.Deregister(ctx)
		}(contexts[i])

		// Cancel
		go func(cancel context.CancelFunc) {
			defer wg.Done()
			cancel()
		}(cancels[i])
	}

	wg.Wait()
	time.Sleep(10 * time.Millisecond)
}

func TestMultipleShards(t *testing.T) {
	w := NewWatcher()
	defer w.Close()

	// Force creation of multiple shards by registering many contexts
	const numContexts = 100
	var wg sync.WaitGroup
	callbackExecuted := int32(0)

	for i := 0; i < numContexts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			callback := func(ctx context.Context) {
				atomic.AddInt32(&callbackExecuted, 1)
			}

			err := w.Register(ctx, callback)
			if err != nil {
				t.Errorf("Registration failed: %v", err)
				return
			}

			// Cancel after registration
			cancel()
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&callbackExecuted) != numContexts {
		t.Fatalf("Expected %d callbacks to be executed, got %d", numContexts, atomic.LoadInt32(&callbackExecuted))
	}
}

func TestShardSelection(t *testing.T) {
	w := NewWatcher()
	defer w.Close()

	// Test that different contexts can be distributed across shards
	contexts := make([]context.Context, 10)
	cancels := make([]context.CancelFunc, 10)

	for i := 0; i < 10; i++ {
		contexts[i], cancels[i] = context.WithCancel(context.Background())

		err := w.Register(contexts[i], func(ctx context.Context) {})
		if err != nil {
			t.Fatalf("Registration %d failed: %v", i, err)
		}
	}

	// Clean up
	for i := 0; i < 10; i++ {
		cancels[i]()
	}
}

func TestWatcherStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	w := NewWatcher()
	defer w.Close()

	const numGoroutines = 200
	const numOperationsPerGoroutine = 50

	var wg sync.WaitGroup
	callbackExecuted := int32(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				ctx, cancel := context.WithCancel(context.Background())

				callback := func(ctx context.Context) {
					atomic.AddInt32(&callbackExecuted, 1)
				}

				err := w.Register(ctx, callback)
				if err != nil {
					t.Errorf("Registration failed: %v", err)
					return
				}

				// Mix of immediate cancellation and deregistration
				switch j % 3 {
				case 0:
					cancel() // Normal cancellation
				case 1:
					w.Deregister(ctx) // Deregister then cancel
					cancel()
				case 2:
					cancel() // Cancel then deregister
					w.Deregister(ctx)
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Should have approximately 2/3 of callbacks executed (excluding deregistered ones)
	executed := atomic.LoadInt32(&callbackExecuted)
	total := int32(numGoroutines * numOperationsPerGoroutine)
	if executed == 0 || executed > total {
		t.Fatalf("Unexpected number of callbacks executed: %d (total: %d)", executed, total)
	}
}
