package handler

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"minipms/internal/middleware"
	"minipms/internal/pkg/response"
	"minipms/internal/service"

	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	svc   *service.AttachmentService
	perms *service.PermService
}

func NewAttachmentHandler(svc *service.AttachmentService, perms *service.PermService) *AttachmentHandler {
	return &AttachmentHandler{svc: svc, perms: perms}
}

func (h *AttachmentHandler) UploadStory(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	res, err := h.svc.UploadForStory(storyID, middleware.UserID(c), file)
	if err != nil {
		if strings.Contains(err.Error(), "不支持") || strings.Contains(err.Error(), "超过限制") {
			response.FailCode(c, 42206, err.Error())
			return
		}
		if strings.Contains(err.Error(), "可交付") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *AttachmentHandler) UploadBug(c *gin.Context) {
	bugID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	res, err := h.svc.UploadForBug(bugID, middleware.UserID(c), file)
	if err != nil {
		if strings.Contains(err.Error(), "不支持") || strings.Contains(err.Error(), "超过限制") {
			response.FailCode(c, 42206, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *AttachmentHandler) UploadStoryRemark(c *gin.Context) {
	storyID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	remarkID, _ := strconv.ParseUint(c.Param("remarkId"), 10, 64)
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}
	res, err := h.svc.UploadForStoryRemark(storyID, remarkID, middleware.UserID(c), file)
	if err != nil {
		if strings.Contains(err.Error(), "不支持") || strings.Contains(err.Error(), "超过限制") {
			response.FailCode(c, 42206, err.Error())
			return
		}
		if strings.Contains(err.Error(), "已定稿") || strings.Contains(err.Error(), "可交付") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *AttachmentHandler) checkListPerm(c *gin.Context, objectType string) bool {
	code := "story.list"
	if objectType == "bug" {
		code = "bug.list"
	}
	// story_remark 与 story 同权
	ok, err := h.perms.HasCode(middleware.UserID(c), code)
	if err != nil || !ok {
		response.Forbidden(c, "无权限")
		return false
	}
	return true
}

func (h *AttachmentHandler) Download(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	att, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if !h.checkListPerm(c, att.ObjectType) {
		return
	}
	absPath := h.svc.AbsPath(att)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		response.NotFound(c, "文件不存在")
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", att.OriginalName))
	c.File(absPath)
}

func (h *AttachmentHandler) Preview(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	att, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if !h.checkListPerm(c, att.ObjectType) {
		return
	}
	if !h.svc.IsImage(att.Ext) {
		response.BadRequest(c, "仅图片可预览")
		return
	}
	absPath := h.svc.AbsPath(att)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		response.NotFound(c, "文件不存在")
		return
	}
	mime := "image/" + att.Ext
	if att.Ext == "jpg" {
		mime = "image/jpeg"
	}
	if att.MimeType != nil && *att.MimeType != "" {
		mime = *att.MimeType
	}
	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", mime)
	c.File(absPath)
}

func (h *AttachmentHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	att, err := h.svc.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	code := "story.attach"
	if att.ObjectType == "bug" {
		code = "bug.attach"
	}
	ok, err := h.perms.HasCode(middleware.UserID(c), code)
	if err != nil || !ok {
		response.Forbidden(c, "无权限")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		if strings.Contains(err.Error(), "不可删除") {
			response.FailCode(c, 42208, err.Error())
			return
		}
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, nil)
}
