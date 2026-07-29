package service

import (
	"fmt"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type BugRemarkService struct {
	db *gorm.DB
}

func NewBugRemarkService(db *gorm.DB) *BugRemarkService {
	return &BugRemarkService{db: db}
}

type BugRemarkVO struct {
	ID          uint64            `json:"id"`
	BugID       uint64            `json:"bugId"`
	Content     string            `json:"content"`
	Finalized   bool              `json:"finalized"`
	CreatedBy   uint64            `json:"createdBy"`
	Creator     *UserBrief        `json:"creator,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	Attachments []AttachmentBrief `json:"attachments,omitempty"`
}

func (s *BugRemarkService) ensureBug(bugID uint64) (*model.Bug, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", bugID).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	return &b, nil
}

func (s *BugRemarkService) List(bugID uint64) ([]BugRemarkVO, error) {
	if _, err := s.ensureBug(bugID); err != nil {
		return nil, err
	}
	var rows []model.BugRemark
	if err := s.db.Where("bug_id = ? AND deleted = 0 AND finalized = 1", bugID).
		Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	vos := make([]BugRemarkVO, 0, len(rows))
	for _, r := range rows {
		vos = append(vos, s.toVO(r))
	}
	return vos, nil
}

func (s *BugRemarkService) toVO(r model.BugRemark) BugRemarkVO {
	vo := BugRemarkVO{
		ID:        r.ID,
		BugID:     r.BugID,
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
	s.db.Where("object_type = ? AND object_id = ? AND deleted = 0", "bug_remark", r.ID).
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

func (s *BugRemarkService) CreateDraft(bugID, userID uint64) (*BugRemarkVO, error) {
	if _, err := s.ensureBug(bugID); err != nil {
		return nil, err
	}
	r := model.BugRemark{
		BugID:     bugID,
		Finalized: 0,
		CreatedBy: userID,
	}
	if err := s.db.Create(&r).Error; err != nil {
		return nil, err
	}
	return &BugRemarkVO{
		ID: r.ID, BugID: r.BugID, Finalized: false, CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

type FinalizeBugRemarkInput struct {
	Content string `json:"content"`
}

func (s *BugRemarkService) Finalize(bugID, remarkID uint64, in FinalizeBugRemarkInput) (*BugRemarkVO, error) {
	if _, err := s.ensureBug(bugID); err != nil {
		return nil, err
	}
	var r model.BugRemark
	if err := s.db.Where("id = ? AND bug_id = ? AND deleted = 0", remarkID, bugID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("备注不存在")
	}
	if r.Finalized == 1 {
		return nil, fmt.Errorf("备注已定稿，不可修改")
	}
	res := s.db.Model(&model.BugRemark{}).
		Where("id = ? AND bug_id = ? AND deleted = 0 AND finalized = 0", remarkID, bugID).
		Updates(map[string]interface{}{
			"content":   in.Content,
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

func (s *BugRemarkService) SoftDeleteDraft(bugID, remarkID uint64) error {
	var r model.BugRemark
	if err := s.db.Where("id = ? AND bug_id = ? AND deleted = 0", remarkID, bugID).First(&r).Error; err != nil {
		return fmt.Errorf("备注不存在")
	}
	if r.Finalized == 1 {
		return fmt.Errorf("已定稿备注不可删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Attachment{}).
			Where("object_type = ? AND object_id = ? AND deleted = 0", "bug_remark", remarkID).
			Update("deleted", 1).Error; err != nil {
			return err
		}
		return tx.Model(&r).Update("deleted", 1).Error
	})
}

func (s *BugRemarkService) GetOwned(bugID, remarkID uint64) (*model.BugRemark, error) {
	var r model.BugRemark
	if err := s.db.Where("id = ? AND bug_id = ? AND deleted = 0", remarkID, bugID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("备注不存在")
	}
	return &r, nil
}
