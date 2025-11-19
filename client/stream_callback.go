package client

import (
	"context"
	"errors"
	"io"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/codes"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/remote/trans/ttstream"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/streaming"
)

// todo: gRPC 部分需要允许内部版注册 Mesh 处理相关逻辑
func processGRPC(ctx context.Context, ri rpcinfo.RPCInfo, err error, cfg *streaming.StreamCallbackConfig) error {
	if err == nil || err == io.EOF {
		// 帮助结束生命周期
		return err
	}
	// gRPC 专属库解决错误
	// TODO，考虑到内部有专属的错误，需要区分
	if st, ok := status.FromError(err); ok {
		// TODO 由于对端可以非常自由地构建和返回 err，所以协议内部必须提供专用的判断函数
		switch st.Code() {
		// Canceled 还有一种场景是退出 handler，也需要特别注明
		case codes.Canceled:
			// 级联 cancel
			if cfg.OnCancelError != nil {
				// 由于 gRPC 不支持更详细的 Cancel Error，所以只能构造一个基础的错误
				// Todo: gRPC Status 支持更为详细的信息
				cErr := streaming.NewCancelError(err, nil)
				cfg.OnCancelError(ctx, cErr)
			}
			return err
		case codes.DeadlineExceeded:
			// Stream 级别超时
			// Todo: 为 Recv/Send 添加特别的超时对应错误
			if cfg.OnTimeoutError != nil {
				tErr := streaming.NewTimeoutError(err, streaming.StreamTimeoutType)
				cfg.OnTimeoutError(ctx, tErr)
			}
			return err
		case codes.Internal:
			// 需要区分对端 handler return 的错误，和 gRPC 本身的错误
			// 最好是在 gRPC 内部先给上标识符，然后将其归类为用户异常
			return err
		default:
			return err
		}
	}
	if bizErr, ok := kerrors.FromBizStatusError(err); ok {
		if cfg.OnBusinessError != nil {
			cfg.OnBusinessError(ctx, bizErr)
		}
		return err
	}
	// 包装一个错误，用于表明这是用户退出 scope 之后，需要集体处理
	// 开源侧使用 code 13，保证兼容性，允许扩展，在内部更换错误码
	err = status.Err(codes.Internal, err.Error())
	return err
	// 完全劫持错误处理
	// 如果用户没有指定，要么是用户自行返回了错误，要么是自行修改了错误，是否需要统一收敛到 biz status err?
	// 还是给一个独立的错误码？
}

func processTTStream(ctx context.Context, err error, cfg *streaming.StreamCallbackConfig) error {
	if err == nil || err == io.EOF {
		// 帮助结束生命周期
		return err
	}
	if ex, ok := err.(*ttstream.Exception); ok {
		if errors.Is(ex, kerrors.ErrStreamingProtocol) && cfg.OnWireError != nil {
			wErr := streaming.WireError{
				//ex: ex,
			}
			cfg.OnWireError(ctx, wErr)
			return err
		}
		if errors.Is(ex, kerrors.ErrStreamingCanceled) && cfg.OnCancelError != nil {
			// todo: return related information
			cErr := streaming.CancelError{
				//path: nil,
				//err:  nil,
			}
			cfg.OnCancelError(ctx, cErr)
			return err
		}
		// todo: timeout
		// todo: ApplicationException 归一化为 BizStatusError
	}
	if bizErr, ok := kerrors.FromBizStatusError(err); ok {
		if cfg.OnBusinessError != nil {
			cfg.OnBusinessError(ctx, bizErr)
			return err
		}
	}

	// todo: 暴露构建 Exception 的函数
	return &ttstream.Exception{}
}
