package service

import (
	"fmt"
	"slices"

	"minipms/internal/model"

	"gorm.io/gorm"
)

var validBugResolutions = []string{"fixed", "duplicate", "willnotfix", "external", "bydesign", "notrepro"}

type BugService struct {
	db *gorm.DB
}

func NewBugService(db *gorm.DB) *BugService {
	return &BugService{db: db}
}

type BugVO struct {
	model.Bug
	ProductName string             `json:"productName,omitempty"`
	Assignee    *UserBrief         `json:"assignee,omitempty"`
	AttachCount int64              `json:"attachCount,omitempty"`
	Attachments []AttachmentBrief  `json:"attachments,omitempty"`
}

func (s *BugService) List(page, pageSize int, productID, projectID, sprintID, storyID uint64, status, severity, pri, assignedTo, keyword string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := s.db.Model(&model.Bug{}).Where("deleted = 0")
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	if sprintID > 0 {
		q = q.Where("sprint_id = ?", sprintID)
	}
	if storyID > 0 {
		q = q.Where("story_id = ?", storyID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if severity != "" {
		q = q.Where("severity = ?", severity)
	}
	if pri != "" {
		q = q.Where("pri = ?", pri)
	}
	if assignedTo != "" {
		q = q.Where("assigned_to = ?", assignedTo)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR steps LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Bug
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	vos := make([]BugVO, 0, len(list))
	for _, b := range list {
		vos = append(vos, s.toListVO(b))
	}
	return &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *BugService) toListVO(b model.Bug) BugVO {
	vo := BugVO{Bug: b}
	var prod model.Product
	if err := s.db.Select("name").Where("id = ?", b.ProductID).First(&prod).Error; err == nil {
		vo.ProductName = prod.Name
	}
	if b.AssignedTo != nil {
		var u model.User
		if err := s.db.Select("id", "account", "realname").Where("id = ? AND deleted = 0", *b.AssignedTo).First(&u).Error; err == nil {
			vo.Assignee = &UserBrief{ID: u.ID, Account: u.Account, Realname: u.Realname}
		}
	}
	s.db.Model(&model.Attachment{}).Where("object_type = ? AND object_id = ? AND deleted = 0", "bug", b.ID).Count(&vo.AttachCount)
	return vo
}

func (s *BugService) toDetailVO(b model.Bug) BugVO {
	vo := s.toListVO(b)
	var atts []model.Attachment
	s.db.Where("object_type = ? AND object_id = ? AND deleted = 0", "bug", b.ID).Order("id DESC").Find(&atts)
	vo.Attachments = make([]AttachmentBrief, 0, len(atts))
	for _, a := range atts {
		vo.Attachments = append(vo.Attachments, AttachmentBrief{
			ID: a.ID, OriginalName: a.OriginalName, Ext: a.Ext,
			SizeBytes: a.SizeBytes, CreatedAt: a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return vo
}

func (s *BugService) Get(id uint64) (*BugVO, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	vo := s.toDetailVO(b)
	return &vo, nil
}

type CreateBugInput struct {
	ProductID  uint64  `json:"productId" binding:"required"`
	Title      string  `json:"title" binding:"required"`
	Steps      *string `json:"steps"`
	Severity   *uint8  `json:"severity"`
	Pri        *uint8  `json:"pri"`
	ProjectID  *uint64 `json:"projectId"`
	SprintID   *uint64 `json:"sprintId"`
	StoryID    *uint64 `json:"storyId"`
	AssignedTo *uint64 `json:"assignedTo"`
}

func (s *BugService) validateCascade(productID uint64, projectID, sprintID, storyID *uint64) error {
	var prod model.Product
	if err := s.db.Where("id = ? AND deleted = 0", productID).First(&prod).Error; err != nil {
		return fmt.Errorf("产品不存在")
	}
	if projectID != nil {
		var proj model.Project
		if err := s.db.Where("id = ? AND deleted = 0", *projectID).First(&proj).Error; err != nil {
			return fmt.Errorf("项目不存在")
		}
		if proj.ProductID != productID {
			return fmt.Errorf("项目不属于该产品")
		}
	}
	if sprintID != nil {
		var sp model.Sprint
		if err := s.db.Where("id = ? AND deleted = 0", *sprintID).First(&sp).Error; err != nil {
			return fmt.Errorf("迭代不存在")
		}
		if projectID == nil || sp.ProjectID != *projectID {
			return fmt.Errorf("迭代不属于该项目")
		}
	}
	if storyID != nil {
		var st model.Story
		if err := s.db.Where("id = ? AND deleted = 0", *storyID).First(&st).Error; err != nil {
			return fmt.Errorf("需求不存在")
		}
		if st.ProductID != productID {
			return fmt.Errorf("需求不属于该产品")
		}
	}
	return nil
}

func (s *BugService) Create(userID uint64, in CreateBugInput) (*BugVO, error) {
	if err := s.validateCascade(in.ProductID, in.ProjectID, in.SprintID, in.StoryID); err != nil {
		return nil, err
	}
	severity := uint8(3)
	if in.Severity != nil {
		severity = *in.Severity
	}
	pri := uint8(3)
	if in.Pri != nil {
		pri = *in.Pri
	}
	b := model.Bug{
		ProductID:  in.ProductID,
		ProjectID:  in.ProjectID,
		SprintID:   in.SprintID,
		StoryID:    in.StoryID,
		Title:      in.Title,
		Steps:      in.Steps,
		Severity:   severity,
		Pri:        pri,
		Status:     "active",
		AssignedTo: in.AssignedTo,
		OpenedBy:   userID,
	}
	if err := s.db.Create(&b).Error; err != nil {
		return nil, err
	}
	return s.Get(b.ID)
}

type UpdateBugInput struct {
	Title      *string `json:"title"`
	Steps      *string `json:"steps"`
	Severity   *uint8  `json:"severity"`
	Pri        *uint8  `json:"pri"`
	ProjectID  *uint64 `json:"projectId"`
	SprintID   *uint64 `json:"sprintId"`
	StoryID    *uint64 `json:"storyId"`
	AssignedTo *uint64 `json:"assignedTo"`
	ClearAssign bool   `json:"clearAssign"`
	Status     *string `json:"status"`
}

func (s *BugService) Update(id uint64, in UpdateBugInput) (*BugVO, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	updates := map[string]interface{}{}
	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Steps != nil {
		updates["steps"] = *in.Steps
	}
	if in.Severity != nil {
		updates["severity"] = *in.Severity
	}
	if in.Pri != nil {
		updates["pri"] = *in.Pri
	}
	if in.ClearAssign {
		updates["assigned_to"] = nil
	} else if in.AssignedTo != nil {
		updates["assigned_to"] = *in.AssignedTo
	}

	newProjectID := b.ProjectID
	newSprintID := b.SprintID
	newStoryID := b.StoryID
	if in.ProjectID != nil {
		newProjectID = in.ProjectID
	}
	if in.SprintID != nil {
		newSprintID = in.SprintID
	}
	if in.StoryID != nil {
		newStoryID = in.StoryID
	}
	if in.ProjectID != nil || in.SprintID != nil || in.StoryID != nil {
		if err := s.validateCascade(b.ProductID, newProjectID, newSprintID, newStoryID); err != nil {
			return nil, err
		}
		if in.ProjectID != nil {
			updates["project_id"] = *in.ProjectID
		}
		if in.SprintID != nil {
			updates["sprint_id"] = *in.SprintID
		}
		if in.StoryID != nil {
			updates["story_id"] = *in.StoryID
		}
	}
	if in.Status != nil && *in.Status == "active" {
		updates["status"] = "active"
		updates["resolution"] = nil
		updates["resolved_by"] = nil
	}
	if len(updates) > 0 {
		if err := s.db.Model(&b).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

type ResolveBugInput struct {
	Resolution string `json:"resolution" binding:"required"`
}

func (s *BugService) Resolve(id, userID uint64, resolution string) (*BugVO, error) {
	if !slices.Contains(validBugResolutions, resolution) {
		return nil, fmt.Errorf("无效的解决方案: %s", resolution)
	}
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	if b.Status != "active" {
		return nil, fmt.Errorf("仅 active 缺陷可解决")
	}
	updates := map[string]interface{}{
		"status":     "resolved",
		"resolution": resolution,
		"resolved_by": userID,
	}
	if err := s.db.Model(&b).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *BugService) Close(id uint64) (*BugVO, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	if b.Status != "resolved" {
		return nil, fmt.Errorf("状态不允许从 %s 变为 closed", b.Status)
	}
	if err := s.db.Model(&b).Update("status", "closed").Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *BugService) Activate(id uint64) (*BugVO, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	updates := map[string]interface{}{
		"status":      "active",
		"resolution":  nil,
		"resolved_by": nil,
	}
	if err := s.db.Model(&b).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *BugService) Delete(id uint64) error {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return fmt.Errorf("缺陷不存在")
	}
	if b.Status != "active" {
		return fmt.Errorf("仅 active 缺陷可删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Attachment{}).Where("object_type = ? AND object_id = ? AND deleted = 0", "bug", id).
			Update("deleted", 1).Error; err != nil {
			return err
		}
		return tx.Model(&model.Bug{}).Where("id = ?", id).Update("deleted", 1).Error
	})
}

func (s *BugService) Exists(id uint64) (*model.Bug, error) {
	var b model.Bug
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("缺陷不存在")
	}
	return &b, nil
}
