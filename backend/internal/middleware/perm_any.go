package middleware

import (
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

func RequireAnyPerm(permService *service.PermService, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get(CtxUserID)
		if !ok {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		for _, code := range codes {
			okPerm, err := permService.HasCode(uid.(uint64), code)
			if err != nil {
				response.Fail(c, 500, 50000, "权限校验失败")
				c.Abort()
				return
			}
			if okPerm {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "无权限")
		c.Abort()
	}
}
