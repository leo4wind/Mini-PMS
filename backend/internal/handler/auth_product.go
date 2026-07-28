package handler

import (
	"errors"
	"strconv"

	"minipms/internal/config"
	"minipms/internal/middleware"
	"minipms/internal/pkg/auth"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg   *config.Config
	auth  *service.AuthService
	perms *service.PermService
}

func NewAuthHandler(cfg *config.Config, authSvc *service.AuthService, perms *service.PermService) *AuthHandler {
	return &AuthHandler{cfg: cfg, auth: authSvc, perms: perms}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Account  string `json:"account" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	user, err := h.auth.Login(req.Account, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrBadCredential) || errors.Is(err, service.ErrUserDisabled) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	token, err := auth.Sign(h.cfg.JWT.Secret, h.cfg.JWT.ExpireHours, user.ID, user.Account)
	if err != nil {
		response.Fail(c, 500, 50000, "签发令牌失败")
		return
	}
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID, "account": user.Account, "realname": user.Realname, "email": user.Email,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.OK(c, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	uid := middleware.UserID(c)
	user, err := h.auth.GetUser(uid)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}
	roles, err := h.perms.RolesOfUser(uid)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	menus, err := h.perms.MenuTreeForUser(uid)
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, gin.H{
		"user": gin.H{
			"id": user.ID, "account": user.Account, "realname": user.Realname, "email": user.Email,
		},
		"roles": roles,
		"menus": menus,
	})
}

func (h *AuthHandler) Password(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.auth.ChangePassword(middleware.UserID(c), req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, nil)
}

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	res, err := h.svc.List(page, pageSize, c.Query("status"), c.Query("keyword"))
	if err != nil {
		response.Fail(c, 500, 50000, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var in service.CreateProductInput
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

func (h *ProductHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	allowed := map[string]bool{"name": true, "code": true, "po": true, "description": true, "status": true}
	updates := map[string]interface{}{}
	for k, v := range body {
		if allowed[k] {
			updates[k] = v
		}
	}
	if err := h.svc.Update(id, updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, _ := h.svc.Get(id)
	response.OK(c, p)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		response.FailCode(c, 42205, err.Error())
		return
	}
	response.OK(c, nil)
}
