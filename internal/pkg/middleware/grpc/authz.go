package grpc

import (
	"context"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"

	"google.golang.org/grpc"
)

// Authorizer 用于定义授权接口的实现.
type Authorizer interface {
	Authorize(sub, obj, act string) (bool, error)
}

// AuthzInterceptor 是一个 gRPC 拦截器，用于进行请求授权.
func AuthzInterceptor(authorizer Authorizer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		subject := contextx.UserID(ctx) // 获取用户 ID
		object := info.FullMethod       // 获取请求资源
		action := "CALL"                // 默认操作

		// 记录授权上下文信息
		log.Debugw("Build authorize context", "subject", subject, "object", object, "action", action)

		// 调用授权接口进行验证
		allowed, err := authorizer.Authorize(subject, object, action)
		if err != nil || !allowed {
			return nil, errno.ErrPermissionDenied.WithMessage("access denied: subject=%s, object=%s, action=%s, reason=%v",
				subject,
				object,
				action,
				err,
			)
		}

		// 继续处理请求
		return handler(ctx, req)
	}
}
