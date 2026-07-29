package handler

import (
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type BugHandler struct {
	svc *service.BugService
}

func NewBugHandler(svc *service.BugService) *BugHandler {
	return &BugHandler{svc: svc}
}

func parseUintQuery(c *gin.Context, key string) uint64 {
	if v := c.Query(key); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		return id
	}
	return 0
}

func (h *BugHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	assignedTo := c.Query("assignedTo")
	if assignedTo == "me" {
		assignedTo = strconv.FormatUint(middleware.UserID(c), 10)
	}
	res, err := h.svc.List(page, pageSize,
		parseUintQuery(c, "productId"),
		parseUintQuery(c, "projectId"),
		parseUintQuery(c, "sprintId"),
		parseUintQuery(c, "storyId"),
		c.Query("status"), c.Query("severity"), c.Query("pri"),
		assignedTo, c.Query("keyword"), c.Query("sortBy"), c.Query("sortOrder"))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Create(c *gin.Context) {
	var in service.CreateBugInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Create(middleware.UserID(c), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.UpdateBugInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Update(id, in)
	if err != nil {
		if strings.Contains(err.Error(), "不可编辑") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Resolve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.ResolveBugInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Resolve(id, middleware.UserID(c), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Close(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Close(id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "状态") {
			response.FailCode(c, 42201, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Activate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Activate(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *BugHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		if strings.Contains(err.Error(), "仅 active") {
			response.FailCode(c, 42207, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, nil)
}
