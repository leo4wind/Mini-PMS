package handler

import (
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type StoryRemarkHandler struct {
	svc *service.StoryRemarkService
}

func NewStoryRemarkHandler(svc *service.StoryRemarkService) *StoryRemarkHandler {
	return &StoryRemarkHandler{svc: svc}
}

func (h *StoryRemarkHandler) List(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.List(storyID)
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

func (h *StoryRemarkHandler) CreateDraft(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.CreateDraft(storyID, middleware.UserID(c))
	if err != nil {
		if strings.Contains(err.Error(), "仅可交付") {
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

func (h *StoryRemarkHandler) Finalize(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	remarkID, _ := strconv.ParseUint(c.Param("remarkId"), 10, 64)
	var in service.FinalizeRemarkInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Finalize(storyID, remarkID, in)
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

func (h *StoryRemarkHandler) DiscardDraft(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	remarkID, _ := strconv.ParseUint(c.Param("remarkId"), 10, 64)
	if err := h.svc.SoftDeleteDraft(storyID, remarkID); err != nil {
		if strings.Contains(err.Error(), "已定稿") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, nil)
}
