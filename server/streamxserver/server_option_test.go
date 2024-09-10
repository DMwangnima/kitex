package streamxserver

import (
	"context"
	"github.com/cloudwego/kitex/streamx"
	"testing"
)

func TestServerOption(t *testing.T) {
	WithStreamMiddleware(func(next streamx.StreamEndpoint) streamx.StreamEndpoint {
		return func(ctx context.Context, req, res any, stream streamx.StreamArgs) (err error) {
			return next(ctx, req, res, stream)
		}
	})
	WithStreamMiddleware[streamx.StreamEndpoint](
		func(next streamx.StreamEndpoint) streamx.StreamEndpoint {
			return func(ctx context.Context, req, res any, stream streamx.StreamArgs) (err error) {
				return next(ctx, req, res, stream)
			}
		},
	)
	WithStreamMiddleware[streamx.ClientStreamEndpoint](
		func(next streamx.ClientStreamEndpoint) streamx.ClientStreamEndpoint {
			return func(ctx context.Context, stream streamx.StreamArgs) error {
				return next(ctx, stream)
			}
		},
	)
	WithStreamMiddleware[streamx.ServerStreamEndpoint](
		func(next streamx.ServerStreamEndpoint) streamx.ServerStreamEndpoint {
			return func(ctx context.Context, req any, stream streamx.StreamArgs) error {
				return next(ctx, req, stream)
			}
		},
	)
}