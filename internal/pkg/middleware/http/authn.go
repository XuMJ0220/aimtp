package http

import (
	"context"
	"aimtp/internal/apiserver/model"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/core"
	"aimtp/pkg/token"

	"github.com/gin-gonic/gin"
)

// UserRetriever 用于根据用户 ID 获取用户信息的接口.
type UserRetriever interface {
	GetUser(ctx context.Context, userID string) (*model.UserM, error)
}

// AuthnMiddleware 是一个认证中间件，用于从 gin.Context 中提取 token 并验证 token 是否合法.
func AuthnMiddleware(retriever UserRetriever) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := token.ParseRequest(c)
		if err != nil {
			core.WriteResponse(c, nil, errno.ErrTokenInvalid.WithMessage(err.Error(), ""))
			c.Abort()
			return
		}

		log.Debugw("Token parsing successful", "userID", userID)

		userM, err := retriever.GetUser(c, userID)
		if err != nil {
			core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage(err.Error(), ""))
			c.Abort()
			return
		}

		ctx := contextx.WithUserID(c.Request.Context(), userM.UserID)
		ctx = contextx.WithUsername(ctx, userM.Username)
		c.Request = c.Request.WithContext(ctx)

		// 继续后续的操作
		c.Next()
	}
}
