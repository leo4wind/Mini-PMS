package handler

import (
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type BugRemarkHandler struct {
	svc *service.BugRemarkService
}

func NewBugRemarkHandler(svc *service.BugRemarkService) *BugRemarkHandler {
	return &BugRemarkHandler{svc: svc}
}

func (h *BugRemarkHandler) List(c *gin.Context) {
	bugID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.List(bugID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *BugRemarkHandler) CreateDraft(c *gin.Context) {
	bugID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.CreateDraft(bugID, middleware.UserID(c))
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugRemarkHandler) Finalize(c *gin.Context) {
	bugID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	remarkID, _ := strconv.ParseUint(c.Param("remarkId"), 10, 64)
	var in service.FinalizeBugRemarkInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Finalize(bugID, remarkID, in)
	if err != nil {
		if strings.Contains(err.Error(), "已定稿") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugRemarkHandler) DiscardDraft(c *gin.Context) {
	bugID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	remarkID, _ := strconv.ParseUint(c.Param("remarkId"), 10, 64)
	if err := h.svc.SoftDeleteDraft(bugID, remarkID); err != nil {
		if strings.Contains(err.Error(), "已定稿") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, nil)
}
