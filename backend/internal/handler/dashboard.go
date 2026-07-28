package handler

import (
	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	res, err := h.svc.Summary(middleware.UserID(c))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}
