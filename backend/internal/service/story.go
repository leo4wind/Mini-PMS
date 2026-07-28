package service

import (
	"fmt"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type StoryService struct {
	db *gorm.DB
}

func NewStoryService(db *gorm.DB) *StoryService {
	return &StoryService{db: db}
}

type StoryVO struct {
	model.Story
	ProductName string             `json:"productName,omitempty"`
	Assignee    *UserBrief         `json:"assignee,omitempty"`
	AttachCount int64              `json:"attachCount,omitempty"`
	Attachments []AttachmentBrief  `json:"attachments,omitempty"`
	Sprints     []SprintBrief        `json:"sprints,omitempty"`
}

type AttachmentBrief struct {
	ID           uint64 `json:"id"`
	OriginalName string `json:"originalName"`
	Ext          string `json:"ext"`
	SizeBytes    uint64 `json:"sizeBytes"`
	CreatedAt    string `json:"createdAt"`
}

type SprintBrief struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	ProjectID uint64 `json:"projectId"`
	Status    string `json:"status"`
}

func (s *StoryService) List(page, pageSize int, productID uint64, storyType, status, assignedToFilter, keyword string, userID uint64) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := s.db.Model(&model.Story{}).Where("deleted = 0")
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	if storyType != "" {
		q = q.Where("type = ?", storyType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if assignedToFilter == "me" && userID > 0 {
		q = q.Where("assigned_to = ?", userID)
	} else if assignedToFilter != "" && assignedToFilter != "me" {
		q = q.Where("assigned_to = ?", assignedToFilter)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR description LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Story
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	vos := make([]StoryVO, 0, len(list))
	for _, st := range list {
		vos = append(vos, s.toListVO(st))
	}
	return &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *StoryService) toListVO(st model.Story) StoryVO {
	vo := StoryVO{Story: st}
	var prod model.Product
	if err := s.db.Select("name").Where("id = ?", st.ProductID).First(&prod).Error; err == nil {
		vo.ProductName = prod.Name
	}
	if st.AssignedTo != nil {
		var u model.User
		if err := s.db.Select("id", "account", "realname").Where("id = ? AND deleted = 0", *st.AssignedTo).First(&u).Error; err == nil {
			vo.Assignee = &UserBrief{ID: u.ID, Account: u.Account, Realname: u.Realname}
		}
	}
	s.db.Model(&model.Attachment{}).Where("object_type = ? AND object_id = ? AND deleted = 0", "story", st.ID).Count(&vo.AttachCount)
	return vo
}

func (s *StoryService) toDetailVO(st model.Story) StoryVO {
	vo := s.toListVO(st)
	var prod model.Product
	if err := s.db.Select("name").Where("id = ?", st.ProductID).First(&prod).Error; err == nil {
		vo.ProductName = prod.Name
	}
	s.db.Model(&model.Attachment{}).Where("object_type = ? AND object_id = ? AND deleted = 0", "story", st.ID).Count(&vo.AttachCount)

	var atts []model.Attachment
	s.db.Where("object_type = ? AND object_id = ? AND deleted = 0", "story", st.ID).Order("id DESC").Find(&atts)
	vo.Attachments = make([]AttachmentBrief, 0, len(atts))
	for _, a := range atts {
		vo.Attachments = append(vo.Attachments, AttachmentBrief{
			ID: a.ID, OriginalName: a.OriginalName, Ext: a.Ext,
			SizeBytes: a.SizeBytes, CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	var sprintIDs []uint64
	s.db.Model(&model.SprintStory{}).Where("story_id = ?", st.ID).Distinct("sprint_id").Pluck("sprint_id", &sprintIDs)
	if len(sprintIDs) > 0 {
		var sprints []model.Sprint
		s.db.Where("id IN ? AND deleted = 0", sprintIDs).Find(&sprints)
		vo.Sprints = make([]SprintBrief, 0, len(sprints))
		for _, sp := range sprints {
			vo.Sprints = append(vo.Sprints, SprintBrief{ID: sp.ID, Name: sp.Name, ProjectID: sp.ProjectID, Status: sp.Status})
		}
	}
	return vo
}

func (s *StoryService) Get(id uint64) (*StoryVO, error) {
	var st model.Story
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&st).Error; err != nil {
		return nil, fmt.Errorf("需求不存在")
	}
	vo := s.toDetailVO(st)
	return &vo, nil
}

type CreateStoryInput struct {
	ProductID   uint64   `json:"productId" binding:"required"`
	Type        *string  `json:"type"`
	Title       string   `json:"title" binding:"required"`
	Description *string  `json:"description"`
	Pri         *uint8   `json:"pri"`
	Estimate    *float64 `json:"estimate"`
	AssignedTo  *uint64  `json:"assignedTo"`
}

func (s *StoryService) Create(userID uint64, in CreateStoryInput) (*StoryVO, error) {
	var prod model.Product
	if err := s.db.Where("id = ? AND deleted = 0", in.ProductID).First(&prod).Error; err != nil {
		return nil, fmt.Errorf("产品不存在")
	}
	if prod.Status == "closed" {
		return nil, fmt.Errorf("产品已关闭，禁止新建需求")
	}
	storyType := "planning"
	if in.Type != nil && *in.Type != "" {
		storyType = *in.Type
	}
	pri := uint8(3)
	if in.Pri != nil {
		pri = *in.Pri
	}
	st := model.Story{
		ProductID:   in.ProductID,
		Type:        storyType,
		Title:       in.Title,
		Description: in.Description,
		Pri:         pri,
		Status:      "draft",
		Estimate:    in.Estimate,
		AssignedTo:  in.AssignedTo,
		OpenedBy:    userID,
	}
	if err := s.db.Create(&st).Error; err != nil {
		return nil, err
	}
	return s.Get(st.ID)
}

type UpdateStoryInput struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Pri         *uint8   `json:"pri"`
	Estimate    *float64 `json:"estimate"`
	AssignedTo  *uint64  `json:"assignedTo"`
	ClearAssign bool     `json:"clearAssign"`
	Type        *string  `json:"type"`
	Status      *string  `json:"status"`
}

var storyStatusNext = map[string]map[string]bool{
	"draft":  {"active": true},
	"active": {"closed": true},
	"closed": {"active": true},
}

func (s *StoryService) Update(id uint64, in UpdateStoryInput) (*StoryVO, error) {
	var st model.Story
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&st).Error; err != nil {
		return nil, fmt.Errorf("需求不存在")
	}
	updates := map[string]interface{}{}
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Pri != nil {
		updates["pri"] = *in.Pri
	}
	if in.Estimate != nil {
		updates["estimate"] = *in.Estimate
	}
	if in.ClearAssign {
		updates["assigned_to"] = nil
	} else if in.AssignedTo != nil {
		updates["assigned_to"] = *in.AssignedTo
	}
	if in.Type != nil {
		updates["type"] = *in.Type
	}
	if in.Status != nil {
		next := *in.Status
		if next != st.Status {
			allow := storyStatusNext[st.Status]
			if allow == nil || !allow[next] {
				return nil, fmt.Errorf("状态不允许从 %s 变为 %s", st.Status, next)
			}
			updates["status"] = next
		}
	}
	if len(updates) > 0 {
		if err := s.db.Model(&st).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

func (s *StoryService) Delete(id uint64) error {
	var st model.Story
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&st).Error; err != nil {
		return fmt.Errorf("需求不存在")
	}
	var cnt int64
	s.db.Model(&model.SprintStory{}).Where("story_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("需求仍被迭代关联，无法删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Attachment{}).Where("object_type = ? AND object_id = ? AND deleted = 0", "story", id).
			Update("deleted", 1).Error; err != nil {
			return err
		}
		res := tx.Model(&model.Story{}).Where("id = ?", id).Update("deleted", 1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("需求不存在")
		}
		return nil
	})
}

func (s *StoryService) ListByProject(projectID uint64) ([]StoryVO, error) {
	var storyIDs []uint64
	s.db.Model(&model.SprintStory{}).Where("project_id = ?", projectID).Distinct("story_id").Pluck("story_id", &storyIDs)
	if len(storyIDs) == 0 {
		return []StoryVO{}, nil
	}
	var list []model.Story
	if err := s.db.Where("id IN ? AND deleted = 0", storyIDs).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	vos := make([]StoryVO, 0, len(list))
	for _, st := range list {
		vos = append(vos, s.toListVO(st))
	}
	return vos, nil
}

func (s *StoryService) Exists(id uint64) (*model.Story, error) {
	var st model.Story
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&st).Error; err != nil {
		return nil, fmt.Errorf("需求不存在")
	}
	return &st, nil
}
