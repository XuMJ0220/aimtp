package grpc

import (
	"context"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/known"

	"aimtp/pkg/errorsx"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RequestIDInterceptor 是一个 gRPC 拦截器，用于设置请求 ID
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		var requestID string
		md, _ := metadata.FromIncomingContext(ctx)

		// 从请求中获取请求 ID
		if requestIDs := md[known.XRequestID]; len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}

		// 如果没有请求 ID，则生成一个新的 UUID
		if requestID == "" {
			requestID = uuid.New().String()
			md.Append(known.XRequestID, requestID)
		}

		// 将元数据设置为新的 incoming context
		// 把 md 注入到 ctx中，以便后续的处理器可以访问它
		ctx = metadata.NewIncomingContext(ctx, md)

		// 将包含请求 ID 的 md 设置到响应的 Header Metadata 中，最终返回给客户端
		// grpc.SetHeader 会在 gRPC 方法响应中添加元数据（Metadata），
		// 此处将包含请求 ID 的 Metadata 设置到 Header 中。
		// 注意：grpc.SetHeader 仅设置数据，它不会立即发送给客户端。
		// Header Metadata 会在 RPC 响应返回时一并发送。
		_ = grpc.SetHeader(ctx, md)

		// 将请求 ID 添加到 ctx 中
		ctx = contextx.WithRequestID(ctx, requestID)

		// 继续处理请求（处理下一个拦截器或最终的处理器）
		res, err := handler(ctx, req)
		// 错误处理，附加请求 ID
		if err != nil {
			return res, errorsx.FromError(err).WithRequestID(requestID)
		}

		return res, nil
	}
}
