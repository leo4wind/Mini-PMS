package handler

import (
	"strconv"

	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	svc *service.SystemService
}

func NewSystemHandler(svc *service.SystemService) *SystemHandler {
	return &SystemHandler{svc: svc}
}

func (h *SystemHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	res, err := h.svc.ListUsers(page, pageSize, c.Query("status"), c.Query("keyword"))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *SystemHandler) GetUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	u, err := h.svc.GetUser(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *SystemHandler) CreateUser(c *gin.Context) {
	var in service.CreateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	u, err := h.svc.CreateUser(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *SystemHandler) UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Realname *string `json:"realname"`
		Email    *string `json:"email"`
		Status   *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	u, err := h.svc.UpdateUser(id, body.Realname, body.Email, body.Status)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *SystemHandler) DisableUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.SetUserStatus(id, "disabled"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SystemHandler) EnableUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.SetUserStatus(id, "active"); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SystemHandler) AssignUserRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		RoleIDs []uint64 `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.RoleIDs == nil {
		body.RoleIDs = []uint64{}
	}
	if err := h.svc.AssignUserRoles(id, body.RoleIDs); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, _ := h.svc.GetUser(id)
	response.OK(c, u)
}

func (h *SystemHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, roles)
}

func (h *SystemHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Name   *string `json:"name"`
		Remark *string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	role, err := h.svc.UpdateRole(id, body.Name, body.Remark)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, role)
}

func (h *SystemHandler) GetRoleMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ids, err := h.svc.GetRoleMenuIDs(id)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, gin.H{"menuIds": ids})
}

func (h *SystemHandler) AssignRoleMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		MenuIDs []uint64 `json:"menuIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.MenuIDs == nil {
		body.MenuIDs = []uint64{}
	}
	if err := h.svc.AssignRoleMenus(id, body.MenuIDs); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SystemHandler) MenuTree(c *gin.Context) {
	tree, err := h.svc.AllMenuTree()
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, tree)
}

func (h *SystemHandler) CreateMenu(c *gin.Context) {
	var in service.MenuInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	m, err := h.svc.CreateMenu(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, m)
}

func (h *SystemHandler) UpdateMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		ParentID *uint64 `json:"parentId"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Type     string  `json:"type"`
		Path     *string `json:"path"`
		Icon     *string `json:"icon"`
		Sort     *int    `json:"sort"`
		Status   *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	in := service.MenuInput{
		ParentID: body.ParentID,
		Code:     body.Code,
		Name:     body.Name,
		Type:     body.Type,
		Path:     body.Path,
		Icon:     body.Icon,
		Sort:     body.Sort,
		Status:   body.Status,
	}
	m, err := h.svc.UpdateMenu(id, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, m)
}

func (h *SystemHandler) DeleteMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteMenu(id); err != nil {
		response.FailCode(c, 42205, err.Error())
		return
	}
	response.OK(c, nil)
}
