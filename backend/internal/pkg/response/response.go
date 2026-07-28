package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 40100, message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 40300, message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 40400, message)
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusUnprocessableEntity, 42200, message)
}

func FailCode(c *gin.Context, code int, message string) {
	Fail(c, http.StatusUnprocessableEntity, code, message)
}
