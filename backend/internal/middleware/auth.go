package middleware

import (
	"strings"

	"minipms/internal/pkg/auth"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

const CtxUserID = "userId"
const CtxAccount = "account"

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.Parse(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			response.Unauthorized(c, "登录已失效")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxAccount, claims.Account)
		c.Next()
	}
}

func RequirePerm(permService *service.PermService, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get(CtxUserID)
		if !ok {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		okPerm, err := permService.HasCode(uid.(uint64), code)
		if err != nil {
			response.Fail(c, 500, 50000, "权限校验失败")
			c.Abort()
			return
		}
		if !okPerm {
			response.Forbidden(c, "无权限: "+code)
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) uint64 {
	v, _ := c.Get(CtxUserID)
	id, _ := v.(uint64)
	return id
}
