package handler

import (
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var productID uint64
	if v := c.Query("productId"); v != "" {
		productID, _ = strconv.ParseUint(v, 10, 64)
	}
	res, err := h.svc.List(page, pageSize, productID, c.Query("status"), c.Query("keyword"))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var in service.CreateProjectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	res, err := h.svc.Create(middleware.UserID(c), in)
	if err != nil {
		if err.Error() == "产品已关闭，禁止新建项目" {
			response.FailCode(c, 42202, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	res, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.UpdateProjectInput
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

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		response.FailCode(c, 42205, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *ProjectHandler) ListByProduct(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.svc.ListByProduct(productID)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, list)
}
