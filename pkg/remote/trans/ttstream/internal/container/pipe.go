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
	"io"
	"sync/atomic"
)

const (
	pipeStateActive int32 = 0
	pipeStateClosed int32 = 1
)

type CtxDoneCallback func(ctx context.Context) error

// Pipe implement a queue that never block on Write but block on Read if there is nothing to read
type Pipe[Item any] struct {
	ctx      context.Context
	callback CtxDoneCallback

	queue   *Queue[Item]
	state   int32
	trigger chan struct{}
}

func NewPipe[Item any](ctx context.Context, callback CtxDoneCallback) *Pipe[Item] {
	p := new(Pipe[Item])
	p.ctx = ctx
	p.callback = callback
	p.queue = NewQueue[Item]()
	p.trigger = make(chan struct{}, 1)
	p.state = pipeStateActive
	return p
}

// Read will block if there is nothing to read
func (p *Pipe[Item]) Read(items []Item) (n int, err error) {
	return p.ReadCtx(nil, nil, items)
}

// ReadCtx will check another ctx
func (p *Pipe[Item]) ReadCtx(perReadCtx context.Context, perReadCallback CtxDoneCallback, items []Item) (n int, err error) {
	for {
		// check readable items
		n = p.read(items)
		if n > 0 {
			return n, nil
		}

		// check state
		state := atomic.LoadInt32(&p.state)
		if state == pipeStateClosed {
			return 0, io.EOF
		}

		// no data to read, waiting writes or ctx done
		var readMore bool
		var wErr error
		if perReadCtx != nil {
			readMore, wErr = p.waitCtx(perReadCtx, perReadCallback)
		} else {
			readMore, wErr = p.wait()
		}

		if !readMore {
			return 0, wErr
		}
	}
}

func (p *Pipe[Item]) read(items []Item) int {
	var n int
	for i := 0; i < len(items); i++ {
		val, ok := p.queue.Get()
		if !ok {
			break
		}
		items[i] = val
		n++
	}
	return n
}

func (p *Pipe[Item]) wait() (readMore bool, err error) {
	select {
	case <-p.ctx.Done():
		if p.callback != nil {
			return false, p.callback(p.ctx)
		}
		return false, p.ctx.Err()
	case <-p.trigger:
		return true, nil
	}
}

func (p *Pipe[Item]) waitCtx(ctx context.Context, callback CtxDoneCallback) (readMore bool, err error) {
	select {
	case <-p.ctx.Done():
		if p.callback != nil {
			return false, p.callback(p.ctx)
		}
		return false, p.ctx.Err()
	case <-ctx.Done():
		if callback != nil {
			return false, callback(ctx)
		}
		return false, ctx.Err()
	case <-p.trigger:
		return true, nil
	}
}

func (p *Pipe[Item]) Write(items ...Item) (err error) {
	for _, item := range items {
		p.queue.Add(item)
	}
	// wake up
	select {
	case p.trigger <- struct{}{}:
	default:
	}
	return nil
}

func (p *Pipe[Item]) Close() {
	if !atomic.CompareAndSwapInt32(&p.state, pipeStateActive, pipeStateClosed) {
		return
	}
	select {
	case p.trigger <- struct{}{}:
	default:
	}
}
