package handler

import (
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type StoryHandler struct {
	svc *service.StoryService
}

func NewStoryHandler(svc *service.StoryService) *StoryHandler {
	return &StoryHandler{svc: svc}
}

func (h *StoryHandler) List(c *gin.Context) {
	var productID uint64
	if v := c.Query("productId"); v != "" {
		productID, _ = strconv.ParseUint(v, 10, 64)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	assignedTo := c.Query("assignedTo")
	userID := middleware.UserID(c)
	res, err := h.svc.List(page, pageSize, productID, c.Query("type"), c.Query("status"), assignedTo, c.Query("keyword"), userID)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *StoryHandler) Create(c *gin.Context) {
	var in service.CreateStoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Create(middleware.UserID(c), in)
	if err != nil {
		if err.Error() == "产品已关闭，禁止新建需求" {
			response.FailCode(c, 42202, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *StoryHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *StoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.UpdateStoryInput
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

func (h *StoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		if strings.Contains(err.Error(), "关联") {
			response.FailCode(c, 42205, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *StoryHandler) ListByProject(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.ListByProject(projectID)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, list)
}
