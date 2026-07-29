package middleware

import (
	"time"

	"minipms/internal/pkg/applog"

	"github.com/gin-gonic/gin"
)

// RequestLog 以 [INFO] 打印 HTTP 请求
func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		if raw != "" {
			path = path + "?" + raw
		}
		applog.Info("HTTP %s %s | status=%d | %s | ip=%s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			applog.FormatDuration(time.Since(start)),
			c.ClientIP(),
		)
	}
}
