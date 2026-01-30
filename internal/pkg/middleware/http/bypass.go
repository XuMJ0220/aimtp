package http

import (
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"

	"github.com/gin-gonic/gin"
)

func AuthnBypasswInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中提取用户ID，假设请求头名称为 "X-User-ID"
		userID := "user-000001" // 默认用户ID
		if val := c.GetHeader(known.XUserID); val != "" {
			userID = val
		}

		log.Debugw("Simulated authentication successful", "userID", userID)

		// 将用户ID和用户名注入到上下文中
		ctx := contextx.WithUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)

		// 继续后续的操作
		c.Next()
	}
}
