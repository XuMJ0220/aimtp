package grpc

import (
	"context"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthnBypasswInterceptor 是一个 gRPC 拦截器，模拟所有请求都通过认证.
func AuthnBypasswInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 这里实际得通过 Token 来设置初始化 userID
		userID := "userID-00001"

		md, _ := metadata.FromIncomingContext(ctx)

		// 从请求中获取用户 ID
		if userIDs := md[known.XUserID]; len(userIDs) > 0 {
			userID = userIDs[0]
		}

		log.Debugw("Simulated authentication successful", "userID", userID)

		// 将默认的用户信息存入上下文
		ctx = context.WithValue(ctx, known.XUserID, userID)

		// 将用户 ID 添加到 ctx 中
		ctx = contextx.WithUserID(ctx, userID)

		// 继续处理请求
		return handler(ctx, req)
	}
}
