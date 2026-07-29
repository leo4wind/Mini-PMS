package service

import (
	"encoding/json"
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

func (s *StoryService) List(page, pageSize int, productID uint64, storyType, status, assignedToFilter, keyword string, userID uint64, withMeta bool) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	type row struct {
		model.Story
		ProductName     string  `gorm:"column:product_name"`
		AssigneeAccount *string `gorm:"column:assignee_account"`
		AssigneeName    *string `gorm:"column:assignee_name"`
		AttachCount     int64   `gorm:"column:attach_count"`
		Total           int64   `gorm:"column:total_count"`
		ProductsJSON    []byte  `gorm:"column:products_json"`
		AssigneesJSON   []byte  `gorm:"column:assignees_json"`
	}

	selectCols := `st.*, prod.name AS product_name,
			au.account AS assignee_account, au.realname AS assignee_name,
			(SELECT COUNT(*) FROM attachment a WHERE a.object_type = 'story' AND a.object_id = st.id AND a.deleted = 0) AS attach_count,
			COUNT(*) OVER() AS total_count`
	if withMeta {
		// 筛选项与列表同一条 SQL 带回（每行重复，读第一行即可）
		selectCols += `,
			meta.products_json AS products_json,
			meta.assignees_json AS assignees_json`
	}

	dataQ := s.db.Table("story st").Select(selectCols).
		Joins("LEFT JOIN product prod ON prod.id = st.product_id").
		Joins("LEFT JOIN `user` au ON au.id = st.assigned_to AND au.deleted = 0")
	if withMeta {
		dataQ = dataQ.Joins(`CROSS JOIN (
			SELECT
				COALESCE((
					SELECT JSON_ARRAYAGG(JSON_OBJECT('id', p.id, 'name', p.name))
					FROM product p WHERE p.deleted = 0 AND p.status = 'normal'
				), JSON_ARRAY()) AS products_json,
				COALESCE((
					SELECT JSON_ARRAYAGG(JSON_OBJECT('id', u.id, 'account', u.account, 'realname', u.realname))
					FROM ` + "`user`" + ` u WHERE u.deleted = 0 AND u.status = 'active'
				), JSON_ARRAY()) AS assignees_json
		) meta`)
	}
	dataQ = dataQ.Where("st.deleted = 0")
	if productID > 0 {
		dataQ = dataQ.Where("st.product_id = ?", productID)
	}
	if storyType != "" {
		dataQ = dataQ.Where("st.type = ?", storyType)
	}
	if status != "" {
		dataQ = dataQ.Where("st.status = ?", status)
	}
	if assignedToFilter == "me" && userID > 0 {
		dataQ = dataQ.Where("st.assigned_to = ?", userID)
	} else if assignedToFilter != "" && assignedToFilter != "me" {
		dataQ = dataQ.Where("st.assigned_to = ?", assignedToFilter)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		dataQ = dataQ.Where("st.title LIKE ? OR st.description LIKE ?", like, like)
	}

	var rows []row
	if err := dataQ.Order("st.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var total int64
	if len(rows) > 0 {
		total = rows[0].Total
	} else {
		// 当前页无数据时窗口函数拿不到 total，补一次 count
		countQ := s.db.Table("story st").Where("st.deleted = 0")
		if productID > 0 {
			countQ = countQ.Where("st.product_id = ?", productID)
		}
		if storyType != "" {
			countQ = countQ.Where("st.type = ?", storyType)
		}
		if status != "" {
			countQ = countQ.Where("st.status = ?", status)
		}
		if assignedToFilter == "me" && userID > 0 {
			countQ = countQ.Where("st.assigned_to = ?", userID)
		} else if assignedToFilter != "" && assignedToFilter != "me" {
			countQ = countQ.Where("st.assigned_to = ?", assignedToFilter)
		}
		if keyword != "" {
			like := "%" + keyword + "%"
			countQ = countQ.Where("st.title LIKE ? OR st.description LIKE ?", like, like)
		}
		_ = countQ.Count(&total)
	}

	vos := make([]StoryVO, 0, len(rows))
	for _, r := range rows {
		vo := StoryVO{Story: r.Story, ProductName: r.ProductName, AttachCount: r.AttachCount}
		if r.Story.AssignedTo != nil {
			vo.Assignee = &UserBrief{ID: *r.Story.AssignedTo}
			if r.AssigneeAccount != nil {
				vo.Assignee.Account = *r.AssigneeAccount
			}
			if r.AssigneeName != nil {
				vo.Assignee.Realname = *r.AssigneeName
			}
		}
		vos = append(vos, vo)
	}

	res := &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}
	if withMeta {
		type productOpt struct {
			ID   uint64 `json:"id"`
			Name string `json:"name"`
		}
		type userOpt struct {
			ID       uint64 `json:"id"`
			Account  string `json:"account"`
			Realname string `json:"realname"`
		}
		products := []productOpt{}
		assignees := []userOpt{}
		if len(rows) > 0 {
			_ = json.Unmarshal(rows[0].ProductsJSON, &products)
			_ = json.Unmarshal(rows[0].AssigneesJSON, &assignees)
		} else {
			type metaRow struct {
				ProductsJSON  []byte `gorm:"column:products_json"`
				AssigneesJSON []byte `gorm:"column:assignees_json"`
			}
			var mr metaRow
			_ = s.db.Raw(`SELECT
				COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id', p.id, 'name', p.name)) FROM product p WHERE p.deleted = 0 AND p.status = 'normal'), JSON_ARRAY()) AS products_json,
				COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT('id', u.id, 'account', u.account, 'realname', u.realname)) FROM ` + "`user`" + ` u WHERE u.deleted = 0 AND u.status = 'active'), JSON_ARRAY()) AS assignees_json`).
				Scan(&mr)
			_ = json.Unmarshal(mr.ProductsJSON, &products)
			_ = json.Unmarshal(mr.AssigneesJSON, &assignees)
		}
		if products == nil {
			products = []productOpt{}
		}
		if assignees == nil {
			assignees = []userOpt{}
		}
		res.Meta = map[string]interface{}{
			"products":  products,
			"assignees": assignees,
		}
	}
	return res, nil
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
