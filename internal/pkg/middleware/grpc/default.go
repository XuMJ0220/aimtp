package grpc

import (
	"context"

	"google.golang.org/grpc"
)

func DefaulterInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, rq any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 调用 Default() 方法（如果存在）
		if defaulter, ok := rq.(interface{ Default() }); ok {
			defaulter.Default()
		}

		// 继续处理请求
		return handler(ctx, rq)
	}
}
