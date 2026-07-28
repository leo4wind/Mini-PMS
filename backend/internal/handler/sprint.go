package handler

import (
	"strconv"
	"strings"

	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type SprintHandler struct {
	svc *service.SprintService
}

func NewSprintHandler(svc *service.SprintService) *SprintHandler {
	return &SprintHandler{svc: svc}
}

func (h *SprintHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var projectID, productID uint64
	if v := c.Query("projectId"); v != "" {
		projectID, _ = strconv.ParseUint(v, 10, 64)
	}
	if v := c.Query("productId"); v != "" {
		productID, _ = strconv.ParseUint(v, 10, 64)
	}
	res, err := h.svc.List(page, pageSize, projectID, productID, c.Query("status"))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *SprintHandler) Create(c *gin.Context) {
	var in service.CreateSprintInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Create(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *SprintHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *SprintHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.UpdateSprintInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Update(id, in)
	if err != nil {
		if strings.HasPrefix(err.Error(), "状态") {
			response.FailCode(c, 42201, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *SprintHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		if strings.Contains(err.Error(), "关联") {
			response.FailCode(c, 42205, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SprintHandler) ListByProject(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.ListByProject(projectID)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *SprintHandler) ListStories(c *gin.Context) {
	sprintID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.ListStories(sprintID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *SprintHandler) LinkStories(c *gin.Context) {
	sprintID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.LinkStoriesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.svc.LinkStories(sprintID, in.StoryIDs); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "非进行中") {
			response.FailCode(c, 42203, msg)
			return
		}
		if strings.Contains(msg, "类型或状态") || strings.Contains(msg, "产品不一致") || strings.Contains(msg, "已在本迭代") {
			response.FailCode(c, 42204, msg)
			return
		}
		response.BadRequest(c, msg)
		return
	}
	response.OK(c, nil)
}

func (h *SprintHandler) UnlinkStory(c *gin.Context) {
	sprintID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	storyID, _ := strconv.ParseUint(c.Param("storyId"), 10, 64)
	if err := h.svc.UnlinkStory(sprintID, storyID); err != nil {
		if strings.Contains(err.Error(), "非进行中") {
			response.FailCode(c, 42203, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SprintHandler) StoryCandidates(c *gin.Context) {
	sprintID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	res, err := h.svc.StoryCandidates(sprintID, page, pageSize, c.Query("keyword"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}
