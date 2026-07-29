package service

import (
	"fmt"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type StoryRemarkService struct {
	db *gorm.DB
}

func NewStoryRemarkService(db *gorm.DB) *StoryRemarkService {
	return &StoryRemarkService{db: db}
}

type StoryRemarkVO struct {
	ID          uint64            `json:"id"`
	StoryID     uint64            `json:"storyId"`
	Content     string            `json:"content"`
	Finalized   bool              `json:"finalized"`
	CreatedBy   uint64            `json:"createdBy"`
	Creator     *UserBrief        `json:"creator,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	Attachments []AttachmentBrief `json:"attachments,omitempty"`
}

func (s *StoryRemarkService) ensureStory(storyID uint64) (*model.Story, error) {
	var st model.Story
	if err := s.db.Where("id = ? AND deleted = 0", storyID).First(&st).Error; err != nil {
		return nil, fmt.Errorf("需求不存在")
	}
	return &st, nil
}

func (s *StoryRemarkService) List(storyID uint64) ([]StoryRemarkVO, error) {
	if _, err := s.ensureStory(storyID); err != nil {
		return nil, err
	}
	var rows []model.StoryRemark
	if err := s.db.Where("story_id = ? AND deleted = 0 AND finalized = 1", storyID).
		Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	vos := make([]StoryRemarkVO, 0, len(rows))
	for _, r := range rows {
		vos = append(vos, s.toVO(r))
	}
	return vos, nil
}

func (s *StoryRemarkService) toVO(r model.StoryRemark) StoryRemarkVO {
	vo := StoryRemarkVO{
		ID:        r.ID,
		StoryID:   r.StoryID,
		Finalized: r.Finalized == 1,
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if r.Content != nil {
		vo.Content = *r.Content
	}
	var u model.User
	if err := s.db.Select("id", "account", "realname").Where("id = ? AND deleted = 0", r.CreatedBy).First(&u).Error; err == nil {
		vo.Creator = &UserBrief{ID: u.ID, Account: u.Account, Realname: u.Realname}
	} else {
		vo.Creator = &UserBrief{ID: r.CreatedBy}
	}
	var atts []model.Attachment
	s.db.Where("object_type = ? AND object_id = ? AND deleted = 0", "story_remark", r.ID).
		Order("id ASC").Find(&atts)
	vo.Attachments = make([]AttachmentBrief, 0, len(atts))
	for _, a := range atts {
		vo.Attachments = append(vo.Attachments, AttachmentBrief{
			ID: a.ID, OriginalName: a.OriginalName, Ext: a.Ext, SizeBytes: a.SizeBytes,
			CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return vo
}

// CreateDraft 创建未定稿备注，供随后上传附件/图片后 finalize
func (s *StoryRemarkService) CreateDraft(storyID, userID uint64) (*StoryRemarkVO, error) {
	st, err := s.ensureStory(storyID)
	if err != nil {
		return nil, err
	}
	if st.Type != "story" {
		return nil, fmt.Errorf("仅可交付需求可追加备注")
	}
	r := model.StoryRemark{
		StoryID:   storyID,
		Finalized: 0,
		CreatedBy: userID,
	}
	if err := s.db.Create(&r).Error; err != nil {
		return nil, err
	}
	return &StoryRemarkVO{
		ID: r.ID, StoryID: r.StoryID, Finalized: false, CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type FinalizeRemarkInput struct {
	Content string `json:"content"`
}

func (s *StoryRemarkService) Finalize(storyID, remarkID uint64, in FinalizeRemarkInput) (*StoryRemarkVO, error) {
	if _, err := s.ensureStory(storyID); err != nil {
		return nil, err
	}
	var r model.StoryRemark
	if err := s.db.Where("id = ? AND story_id = ? AND deleted = 0", remarkID, storyID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("备注不存在")
	}
	if r.Finalized == 1 {
		return nil, fmt.Errorf("备注已定稿，不可修改")
	}
	content := in.Content
	res := s.db.Model(&model.StoryRemark{}).
		Where("id = ? AND story_id = ? AND deleted = 0 AND finalized = 0", remarkID, storyID).
		Updates(map[string]interface{}{
			"content":   content,
			"finalized": 1,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("备注已定稿，不可修改")
	}
	_ = s.db.Where("id = ?", remarkID).First(&r)
	vo := s.toVO(r)
	return &vo, nil
}

// SoftDeleteDraft 清理未定稿备注（仅系统/失败回滚用）
func (s *StoryRemarkService) SoftDeleteDraft(storyID, remarkID uint64) error {
	var r model.StoryRemark
	if err := s.db.Where("id = ? AND story_id = ? AND deleted = 0", remarkID, storyID).First(&r).Error; err != nil {
		return fmt.Errorf("备注不存在")
	}
	if r.Finalized == 1 {
		return fmt.Errorf("已定稿备注不可删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Attachment{}).
			Where("object_type = ? AND object_id = ? AND deleted = 0", "story_remark", remarkID).
			Update("deleted", 1).Error; err != nil {
			return err
		}
		return tx.Model(&r).Update("deleted", 1).Error
	})
}

func (s *StoryRemarkService) GetOwned(storyID, remarkID uint64) (*model.StoryRemark, error) {
	var r model.StoryRemark
	if err := s.db.Where("id = ? AND story_id = ? AND deleted = 0", remarkID, storyID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("备注不存在")
	}
	return &r, nil
}
