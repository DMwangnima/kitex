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

package streaming

import (
	"context"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

// 回调中可能需要考虑到一个状态，比如 req/resp 等等，如何拿到？
// 两个传输方向，第几个包，以及具体的包是什么

// 此处的回调不应该阻塞，可以考虑添加一个专门的 Trace Event 用于定位
// 但如果阻塞了，上报也无从谈起，只能说 Trace Event 可以用于定位回调执行时间长
type StreamCallbackConfig struct {
	// 此处的所有错误处理，其实都是用于自定义打点和日志处理，因此需要提供最准确的 Stream 会话信息

	// WireError 不允许修改，专门处理框架错误
	OnWireError func(ctx context.Context, wErr WireError)
	// OnBusinessError 不允许修改，专门处理业务异常，但这里需要考虑对端 handler 直接 return err 的场景
	OnBusinessError func(ctx context.Context, bizErr kerrors.BizStatusErrorIface)
	// 超时错误应该也属于 WireError，但单独拎出来处理
	// Stream 整体超时，Recv/Send 超时
	OnTimeoutError func(ctx context.Context, tErr TimeoutError)

	OnCancelError func(ctx context.Context, cErr CancelError)
}

type ServerStreamingCallback[Res any, Cli ServerStreamingClient[Res]] func(ctx context.Context, cli Cli) error

type ClientStreamingCallback[Req, Res any, Cli ClientStreamingClient[Req, Res]] func(ctx context.Context, cli Cli) error

type BidiStreamingCallback[Req any, Res any, Cli BidiStreamingClient[Req, Res]] func(ctx context.Context, cli Cli) error

func HandleServerStreaming[Res any, Cli ServerStreamingClient[Res]](ctx context.Context, cli Cli, callback ServerStreamingCallback[Res, Cli]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// todo: 记录堆栈
		}
		FinishClientStream(cli, err)
	}()

	return callback(ctx, cli)
}

func HandleClientStreaming[Req, Res any, Cli ClientStreamingClient[Req, Res]](ctx context.Context, cli Cli, callback ClientStreamingCallback[Req, Res, Cli]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// todo: 记录堆栈
		}
		FinishClientStream(cli, err)
	}()
	return callback(ctx, cli)
}

func HandleBidiStreaming[Req, Res any, Cli BidiStreamingClient[Req, Res]](ctx context.Context, cli Cli, callback BidiStreamingCallback[Req, Res, Cli]) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// todo: 记录堆栈
		}
		FinishClientStream(cli, err)
	}()

	return callback(ctx, cli)
}
