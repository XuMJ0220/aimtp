package http

import (
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/log"
	"aimtp/pkg/core"

	"github.com/gin-gonic/gin"
)

// Authorizer 用于定义授权接口的实现.
type Authorizer interface {
	Authorize(sub, obj, act string) (bool, error)
}

// AuthzInterceptor 是一个 gin 中间件，用于进行请求授权.
func AuthzMiddleware(authorizer Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject := contextx.UserID(c.Request.Context())
		object := c.Request.URL.Path
		action := c.Request.Method

		// 记录授权上下文信息
		log.Debugw("Build authorize context", "subject", subject, "object", object, "action", action)

		// 调用授权接口进行验证
		if allowed, err := authorizer.Authorize(subject, object, action); err != nil || !allowed {
			core.WriteResponse(c, nil, errno.ErrPermissionDenied.WithMessage(
				"access denied: subject=%s, object=%s, action=%s, reason=%v",
				subject,
				object,
				action,
				err,
			))
			c.Abort()
			return
		}

		// 继续后续的操作
		c.Next()
	}
}
