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

package client

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/profiler"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

var errUserProvidedContextNonCompliant = errors.New("user-provided context does not comply with context.Context conventions, ctx.Done() triggered but ctx.Err() returns nil. please check your code")

// timeoutPool is a worker pool for task with timeout
type timeoutPool struct {
	size  int32
	tasks chan *timeoutTask

	// maxIdle is the number of the max idle workers in the pool.
	// if maxIdle too small, the pool works like a native 'go func()'.
	maxIdle int32
	// maxIdleTime is the max idle time that the worker will wait for the new task.
	maxIdleTime time.Duration

	mu     sync.Mutex
	ticker chan struct{}
}

// newTimeoutPool ...
func newTimeoutPool(maxIdle int, maxIdleTime time.Duration) *timeoutPool {
	return &timeoutPool{
		tasks:       make(chan *timeoutTask),
		maxIdle:     int32(maxIdle),
		maxIdleTime: maxIdleTime,
	}
}

// Size returns the number of the running workers.
func (p *timeoutPool) Size() int32 {
	return atomic.LoadInt32(&p.size)
}

func (p *timeoutPool) createTicker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// make sure previous goroutine will be closed before creating a new one
	if p.ticker != nil {
		close(p.ticker)
	}
	ch := make(chan struct{})
	p.ticker = ch

	go func(done <-chan struct{}) {
		// if maxIdleTime=60s, maxIdle=100
		// it sends noop task every 60ms
		// but always d >= 10*time.Millisecond
		// this may cause goroutines take more time to exit which is acceptable.
		d := p.maxIdleTime / time.Duration(p.maxIdle) / 10
		if d < 10*time.Millisecond {
			d = 10 * time.Millisecond
		}
		tk := time.NewTicker(d)
		defer tk.Stop()
		for p.Size() > 0 {
			select {
			case <-tk.C:
			case <-done:
				return
			}
			select {
			case p.tasks <- nil: // noop task for checking idletime
			case <-tk.C:
			}
		}
	}(ch)
}

func (p *timeoutPool) createWorker(t *timeoutTask) bool {
	if n := atomic.AddInt32(&p.size, 1); n < p.maxIdle {
		if n == 1 {
			p.createTicker()
		}
		go func(t *timeoutTask) {
			defer atomic.AddInt32(&p.size, -1)

			t.Run()

			lastactive := time.Now()
			for t := range p.tasks {
				if t == nil { // from `createTicker` func
					if time.Since(lastactive) > p.maxIdleTime {
						break
					}
					continue
				}
				t.Run()
				lastactive = time.Now()
			}
		}(t)
		return true
	} else {
		atomic.AddInt32(&p.size, -1)
		return false
	}
}

// RunTask creates/reuses a worker to run task.
//
// It returns:
// - the underlying ctx used for calling endpoint.Endpoint
// - err returned by endpoint.Endpoint or ctx.Err()
//
// NOTE:
// Err() method of the ctx will always be set context.Canceled after the func returns,
// because ctx.Done() will be closed, and context package requires Err() must not nil.
// Caller should always check the err first,
// if it's nil, which means everything is ok, no need to do the further ctx.Err() check.
func (p *timeoutPool) RunTask(ctx context.Context, timeout time.Duration,
	req, resp any, ep endpoint.Endpoint,
) (context.Context, error) {
	t := newTimeoutTask(ctx, timeout, req, resp, ep)
	select {
	case p.tasks <- t:
		return t.Wait()
	default:
	}
	if !p.createWorker(t) {
		// if created worker, t.Run() will be called in worker goroutine
		// if NOT, we should go t.Run() here.
		go t.Run()
	}
	return t.Wait()
}

var poolTask = sync.Pool{
	New: func() any {
		return &timeoutTask{}
	},
}

// timeoutTask is the function that the worker will execute.
type timeoutTask struct {
	ctx *timeoutContext

	wg sync.WaitGroup

	req, resp any
	ep        endpoint.Endpoint

	err atomic.Value
}

func newTimeoutTask(ctx context.Context, timeout time.Duration,
	req, resp any, ep endpoint.Endpoint,
) *timeoutTask {
	t := poolTask.Get().(*timeoutTask)

	// timeoutContext must not be reused,
	// coz user may keep ref to it even though after calling endpoint.Endpoint
	t.ctx = newTimeoutContext(ctx, timeout)

	t.req, t.resp = req, resp
	t.ep = ep
	t.err = atomic.Value{}
	t.wg.Add(1) // for Wait, Wait must be called before Recycle()
	return t
}

func (t *timeoutTask) recycle() {
	// make sure Wait done before returning it to pool
	t.wg.Wait()

	t.ctx = nil
	t.req, t.resp = nil, nil
	t.ep = nil
	t.err = atomic.Value{}
	poolTask.Put(t)
}

func (t *timeoutTask) Cancel(err error) {
	t.ctx.Cancel(err)
}

// Run must be called in a separated goroutine
func (t *timeoutTask) Run() {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			ri := rpcinfo.GetRPCInfo(t.ctx)
			if ri != nil {
				t.err.Store(rpcinfo.ClientPanicToErr(t.ctx, panicInfo, ri, true))
			} else {
				t.err.Store(fmt.Errorf("KITEX: panic without rpcinfo, error=%v\nstack=%s",
					panicInfo, debug.Stack()))
			}
		}
		t.Cancel(context.Canceled)
		t.recycle()
	}()
	var err error
	if profiler.IsEnabled(t.ctx) {
		profiler.Tag(t.ctx)
		err = t.ep(t.ctx, t.req, t.resp)
		profiler.Untag(t.ctx)
	} else {
		err = t.ep(t.ctx, t.req, t.resp)
	}
	if err != nil { // panic if store nil
		t.err.Store(err)
	}
}

// Wait waits Run finishes and returns result
func (t *timeoutTask) Wait() (context.Context, error) {
	defer t.wg.Done()
	dl, ok := t.ctx.Deadline()
	if !ok {
		return t.waitNoTimeout()
	}
	d := time.Until(dl)
	if d < 0 {
		t.Cancel(context.DeadlineExceeded)
		return t.ctx, t.ctx.Err()
	}
	tm := time.NewTimer(d)
	defer tm.Stop()
	select {
	case <-t.ctx.Done(): // finished before timeout
		v := t.err.Load()
		if v != nil {
			return t.ctx, v.(error)
		}
		return t.ctx, nil

	case <-t.ctx.Context.Done(): // parent done
		return t.handleParentDone()
	case <-tm.C: // timeout
		t.Cancel(context.DeadlineExceeded)
	}
	return t.ctx, t.ctx.Err()
}

func (t *timeoutTask) waitNoTimeout() (context.Context, error) {
	select {
	case <-t.ctx.Done(): // finished before parent done
		v := t.err.Load()
		if v != nil {
			return t.ctx, v.(error)
		}
		return t.ctx, nil

	case <-t.ctx.Context.Done(): // parent done
		return t.handleParentDone()
	}
}

// handleParentDone is called when t.ctx.Context.Done() fires (parent context cancelled or timed out).
// It propagates the parent's cancellation to the internal timeoutContext via Cancel.
//
// Defense against non-compliant context implementations:
//
// The context.Context contract requires that after Done() is closed, Err() must return non-nil.
// However, some custom implementations violate this — for example, a "WithoutCancel" wrapper
// that embeds the parent (inheriting its Done channel) but overrides Err() to always return nil.
// When such a context is used and Cancel receives nil, it becomes a no-op: the timeoutContext's
// ch is never closed and Err() remains nil. This causes the following panic chain:
//
//  1. Wait() returns (ctx, nil) to rpcTimeoutMW → rpcTimeoutMW returns nil to Call()
//  2. Call() sees err==nil → sets recycleRI=true → PutRPCInfo recycles the RPCInfo (zeroing ri.to)
//  3. Worker goroutine is still running → accesses ri.To().ServiceName() → nil pointer panic
//  4. The panic is caught by Run()'s recover → ClientPanicToErr also calls ri.To() → double panic
//
// To prevent this, we check parentErr before Cancel. If nil, we substitute a sentinel error.
// This guarantees Cancel always receives a non-nil error, so the timeoutContext's ch is always
// closed (unblocking downstream goroutines) and Err() always returns non-nil.
func (t *timeoutTask) handleParentDone() (context.Context, error) {
	parentErr := t.ctx.Context.Err()
	if parentErr == nil {
		parentErr = errUserProvidedContextNonCompliant
	}
	t.Cancel(parentErr)
	return t.ctx, t.ctx.Err()
}

type timeoutContext struct {
	context.Context

	dl time.Time
	ch chan struct{}

	mu  sync.Mutex
	err error
}

func newTimeoutContext(ctx context.Context, timeout time.Duration) *timeoutContext {
	ret := &timeoutContext{Context: ctx, ch: make(chan struct{})}
	deadline, ok := ctx.Deadline()
	if ok {
		ret.dl = deadline
	}
	if timeout != 0 {
		dl := time.Now().Add(timeout)
		if ret.dl.IsZero() || dl.Before(ret.dl) {
			// The new deadline is sooner than the ctx one
			ret.dl = dl
		}
	}
	return ret
}

func (p *timeoutContext) Deadline() (deadline time.Time, ok bool) {
	return p.dl, !p.dl.IsZero()
}

func (p *timeoutContext) Done() <-chan struct{} {
	return p.ch
}

func (p *timeoutContext) Err() error {
	p.mu.Lock()
	err := p.err
	p.mu.Unlock()
	if err != nil {
		return err
	}
	return p.Context.Err()
}

func (p *timeoutContext) Cancel(err error) {
	p.mu.Lock()
	if err != nil && p.err == nil {
		p.err = err
		close(p.ch)
	}
	p.mu.Unlock()
}
