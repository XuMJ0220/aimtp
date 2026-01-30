package grpc

import (
	"context"
	"aimtp/internal/apiserver/model"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/token"

	"google.golang.org/grpc"
)

// UserRetriever 用于根据用户 ID 获取用户信息的接口.
type UserRetriever interface {
	GetUser(ctx context.Context, userID string) (*model.UserM, error)
}

func AuthnInterceptor(retriever UserRetriever) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		userID, err := token.ParseRequest(ctx)
		if err != nil {
			log.Errorw("Failed to parse request", "err", err)
			return nil, errno.ErrTokenInvalid.WithMessage(err.Error(), "")
		}

		log.Debugw("Token parsing successful", "userID", userID)

		// 获取用户信息
		userM, err := retriever.GetUser(ctx, userID)
		if err != nil {
			log.Errorw("Failed to get user", "err", err)
			return nil, errno.ErrUnauthenticated.WithMessage(err.Error(), "")
		}

		// 往 ctx 中注入 userIDKey{} 和 userNameKey{}
		// 具体对应的是请求用户自己本身的 userID 和 userName
		ctx = contextx.WithUserID(ctx, userM.UserID)
		ctx = contextx.WithUsername(ctx, userM.Username)

		// 将用户信息存入上下文
		ctx = context.WithValue(ctx, known.XUserID, userM.UserID)
		ctx = context.WithValue(ctx, known.XUsername, userM.Username)

		// 继续处理请求
		return handler(ctx, req)
	}
}
