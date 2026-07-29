package service

import (
	"fmt"
	"slices"
	"strings"

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
	Creator     *UserBrief         `json:"creator,omitempty"`
	AttachCount int64              `json:"attachCount,omitempty"`
	Attachments []AttachmentBrief  `json:"attachments,omitempty"`
}

var bugListSortCols = map[string]string{
	"severity": "b.severity",
	"pri":      "b.pri",
	"status":   "b.status",
}

func (s *BugService) List(page, pageSize int, productID, projectID, sprintID, storyID uint64, status, severity, pri, assignedTo, keyword, sortBy, sortOrder string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	type row struct {
		model.Bug
		ProductName     string  `gorm:"column:product_name"`
		AssigneeAccount *string `gorm:"column:assignee_account"`
		AssigneeName    *string `gorm:"column:assignee_name"`
		CreatorAccount  *string `gorm:"column:creator_account"`
		CreatorName     *string `gorm:"column:creator_name"`
		AttachCount     int64   `gorm:"column:attach_count"`
		Total           int64   `gorm:"column:total_count"`
	}

	where := "b.deleted = 0"
	args := make([]interface{}, 0, 16)
	if productID > 0 {
		where += " AND b.product_id = ?"
		args = append(args, productID)
	}
	if projectID > 0 {
		where += " AND b.project_id = ?"
		args = append(args, projectID)
	}
	if sprintID > 0 {
		where += " AND b.sprint_id = ?"
		args = append(args, sprintID)
	}
	if storyID > 0 {
		where += " AND b.story_id = ?"
		args = append(args, storyID)
	}
	if status != "" {
		where += " AND b.status = ?"
		args = append(args, status)
	}
	if severity != "" {
		where += " AND b.severity = ?"
		args = append(args, severity)
	}
	if pri != "" {
		where += " AND b.pri = ?"
		args = append(args, pri)
	}
	if assignedTo != "" {
		where += " AND b.assigned_to = ?"
		args = append(args, assignedTo)
	}
	if keyword != "" {
		where += " AND (b.title LIKE ? OR b.steps LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	orderSQL := "b.id DESC"
	if col, ok := bugListSortCols[sortBy]; ok {
		switch strings.ToLower(sortOrder) {
		case "asc":
			orderSQL = col + " ASC, b.id DESC"
		case "desc":
			orderSQL = col + " DESC, b.id DESC"
		}
	}

	sql := `SELECT b.*, prod.name AS product_name,
			au.account AS assignee_account, au.realname AS assignee_name,
			cu.account AS creator_account, cu.realname AS creator_name,
			COALESCE(ac.attach_count, 0) AS attach_count,
			COUNT(*) OVER() AS total_count
		FROM bug b
		LEFT JOIN product prod ON prod.id = b.product_id
		LEFT JOIN ` + "`user`" + ` au ON au.id = b.assigned_to AND au.deleted = 0
		LEFT JOIN ` + "`user`" + ` cu ON cu.id = b.opened_by AND cu.deleted = 0
		LEFT JOIN (
			SELECT object_id, COUNT(*) AS attach_count
			FROM attachment
			WHERE object_type = 'bug' AND deleted = 0
			GROUP BY object_id
		) ac ON ac.object_id = b.id
		WHERE ` + where + `
		ORDER BY ` + orderSQL + `
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	var rows []row
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var total int64
	vos := make([]BugVO, 0, len(rows))
	for _, r := range rows {
		total = r.Total
		vo := BugVO{Bug: r.Bug, ProductName: r.ProductName, AttachCount: r.AttachCount}
		if r.Bug.AssignedTo != nil {
			vo.Assignee = &UserBrief{ID: *r.Bug.AssignedTo}
			if r.AssigneeAccount != nil {
				vo.Assignee.Account = *r.AssigneeAccount
			}
			if r.AssigneeName != nil {
				vo.Assignee.Realname = *r.AssigneeName
			}
		}
		vo.Creator = &UserBrief{ID: r.Bug.OpenedBy}
		if r.CreatorAccount != nil {
			vo.Creator.Account = *r.CreatorAccount
		}
		if r.CreatorName != nil {
			vo.Creator.Realname = *r.CreatorName
		}
		vos = append(vos, vo)
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
	var cu model.User
	if err := s.db.Select("id", "account", "realname").Where("id = ? AND deleted = 0", b.OpenedBy).First(&cu).Error; err == nil {
		vo.Creator = &UserBrief{ID: cu.ID, Account: cu.Account, Realname: cu.Realname}
	} else {
		vo.Creator = &UserBrief{ID: b.OpenedBy}
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
