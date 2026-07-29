package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"minipms/internal/config"
	"minipms/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var allowedExts = []string{"doc", "docx", "txt", "md", "jpg", "jpeg", "png", "gif", "webp", "mp4"}

var imageExts = []string{"jpg", "jpeg", "png", "gif", "webp"}

var previewableExts = []string{"jpg", "jpeg", "png", "gif", "webp", "mp4"}

type AttachmentService struct {
	db        *gorm.DB
	uploadCfg config.UploadConfig
}

func NewAttachmentService(db *gorm.DB, upload config.UploadConfig) *AttachmentService {
	return &AttachmentService{db: db, uploadCfg: upload}
}

type AttachmentVO struct {
	ID           uint64 `json:"id"`
	ObjectType   string `json:"objectType"`
	ObjectID     uint64 `json:"objectId"`
	OriginalName string `json:"originalName"`
	Ext          string `json:"ext"`
	MimeType     string `json:"mimeType,omitempty"`
	SizeBytes    uint64 `json:"sizeBytes"`
	UploadedBy   uint64 `json:"uploadedBy"`
	CreatedAt    string `json:"createdAt"`
}

func (s *AttachmentService) UploadForStory(storyID, userID uint64, fileHeader *multipart.FileHeader) (*AttachmentVO, error) {
	st, err := NewStoryService(s.db).Exists(storyID)
	if err != nil {
		return nil, err
	}
	if st.Type == "story" {
		return nil, fmt.Errorf("可交付需求不可再上传需求级附件，请追加备注")
	}
	return s.doUpload("story", storyID, userID, fileHeader)
}

func (s *AttachmentService) UploadForBug(bugID, userID uint64, fileHeader *multipart.FileHeader) (*AttachmentVO, error) {
	b, err := NewBugService(s.db).Exists(bugID)
	if err != nil {
		return nil, err
	}
	// 正文已写入后不可再挂缺陷级附件，请走备注
	if b.Steps != nil && strings.TrimSpace(*b.Steps) != "" {
		return nil, fmt.Errorf("缺陷正文已锁定，请追加备注")
	}
	return s.doUpload("bug", bugID, userID, fileHeader)
}

func (s *AttachmentService) UploadForStoryRemark(storyID, remarkID, userID uint64, fileHeader *multipart.FileHeader) (*AttachmentVO, error) {
	r, err := NewStoryRemarkService(s.db).GetOwned(storyID, remarkID)
	if err != nil {
		return nil, err
	}
	if r.Finalized == 1 {
		return nil, fmt.Errorf("备注已定稿，不可再上传附件")
	}
	return s.doUpload("story_remark", remarkID, userID, fileHeader)
}

func (s *AttachmentService) UploadForBugRemark(bugID, remarkID, userID uint64, fileHeader *multipart.FileHeader) (*AttachmentVO, error) {
	r, err := NewBugRemarkService(s.db).GetOwned(bugID, remarkID)
	if err != nil {
		return nil, err
	}
	if r.Finalized == 1 {
		return nil, fmt.Errorf("备注已定稿，不可再上传附件")
	}
	return s.doUpload("bug_remark", remarkID, userID, fileHeader)
}

func (s *AttachmentService) Delete(id uint64) error {
	att, err := s.Get(id)
	if err != nil {
		return err
	}
	if att.ObjectType == "story_remark" || att.ObjectType == "bug_remark" {
		return fmt.Errorf("备注附件不可删除")
	}
	// 可交付需求的需求级附件也不允许删
	if att.ObjectType == "story" {
		st, err := NewStoryService(s.db).Exists(att.ObjectID)
		if err == nil && st.Type == "story" {
			return fmt.Errorf("可交付需求附件不可删除")
		}
	}
	if att.ObjectType == "bug" {
		b, err := NewBugService(s.db).Exists(att.ObjectID)
		if err == nil && b.Steps != nil && strings.TrimSpace(*b.Steps) != "" {
			return fmt.Errorf("缺陷附件不可删除")
		}
	}
	res := s.db.Model(&model.Attachment{}).Where("id = ? AND deleted = 0", id).Update("deleted", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("附件不存在")
	}
	return nil
}

func (s *AttachmentService) doUpload(objectType string, objectID, userID uint64, fileHeader *multipart.FileHeader) (*AttachmentVO, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("未选择文件")
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileHeader.Filename), "."))
	if ext == "" || !slices.Contains(allowedExts, ext) {
		return nil, fmt.Errorf("不支持的文件类型")
	}
	maxBytes := s.uploadCfg.MaxSizeMB * 1024 * 1024
	if fileHeader.Size > maxBytes {
		return nil, fmt.Errorf("文件大小超过限制 %dMB", s.uploadCfg.MaxSizeMB)
	}

	storedName := uuid.New().String() + "." + ext
	subDir := time.Now().Format("2006/01")
	storagePath := filepath.Join(subDir, storedName)
	absDir := filepath.Join(s.uploadCfg.Dir, subDir)
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败")
	}
	absPath := filepath.Join(s.uploadCfg.Dir, storagePath)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("读取文件失败")
	}
	defer src.Close()

	dst, err := os.Create(absPath)
	if err != nil {
		return nil, fmt.Errorf("保存文件失败")
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		os.Remove(absPath)
		return nil, fmt.Errorf("保存文件失败")
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	var mimePtr *string
	if mimeType != "" {
		mimePtr = &mimeType
	}

	att := model.Attachment{
		ObjectType:   objectType,
		ObjectID:     objectID,
		OriginalName: fileHeader.Filename,
		StoredName:   storedName,
		StoragePath:  filepath.ToSlash(storagePath),
		Ext:          ext,
		MimeType:     mimePtr,
		SizeBytes:    uint64(written),
		UploadedBy:   userID,
	}
	if err := s.db.Create(&att).Error; err != nil {
		os.Remove(absPath)
		return nil, err
	}
	return s.toVO(att), nil
}

func (s *AttachmentService) toVO(a model.Attachment) *AttachmentVO {
	vo := &AttachmentVO{
		ID: a.ID, ObjectType: a.ObjectType, ObjectID: a.ObjectID,
		OriginalName: a.OriginalName, Ext: a.Ext, SizeBytes: a.SizeBytes,
		UploadedBy: a.UploadedBy, CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.MimeType != nil {
		vo.MimeType = *a.MimeType
	}
	return vo
}

func (s *AttachmentService) Get(id uint64) (*model.Attachment, error) {
	var att model.Attachment
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&att).Error; err != nil {
		return nil, fmt.Errorf("附件不存在")
	}
	return &att, nil
}

func (s *AttachmentService) AbsPath(att *model.Attachment) string {
	return filepath.Join(s.uploadCfg.Dir, filepath.FromSlash(att.StoragePath))
}

func (s *AttachmentService) IsImage(ext string) bool {
	return slices.Contains(imageExts, strings.ToLower(ext))
}

func (s *AttachmentService) IsPreviewable(ext string) bool {
	return slices.Contains(previewableExts, strings.ToLower(ext))
}
