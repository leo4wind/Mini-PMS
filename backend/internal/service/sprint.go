package service

import (
	"fmt"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type SprintService struct {
	db *gorm.DB
}

func NewSprintService(db *gorm.DB) *SprintService {
	return &SprintService{db: db}
}

type SprintVO struct {
	model.Sprint
	ProjectName string `json:"projectName,omitempty"`
	ProductID   uint64 `json:"productId,omitempty"`
	ProductName string `json:"productName,omitempty"`
	StoryCount  int64  `json:"storyCount"`
}

func (s *SprintService) List(page, pageSize int, projectID, productID uint64, status string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	countQ := s.db.Table("sprint sp").
		Joins("JOIN project proj ON proj.id = sp.project_id AND proj.deleted = 0").
		Where("sp.deleted = 0")
	if projectID > 0 {
		countQ = countQ.Where("sp.project_id = ?", projectID)
	}
	if productID > 0 {
		countQ = countQ.Where("proj.product_id = ?", productID)
	}
	if status != "" {
		countQ = countQ.Where("sp.status = ?", status)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, err
	}

	type row struct {
		model.Sprint
		ProjectName string `gorm:"column:project_name"`
		ProductID   uint64 `gorm:"column:product_id"`
		ProductName string `gorm:"column:product_name"`
		StoryCount  int64  `gorm:"column:story_count"`
	}
	dataQ := s.db.Table("sprint sp").
		Select(`sp.*, proj.name AS project_name, proj.product_id AS product_id, prod.name AS product_name,
			(SELECT COUNT(*) FROM sprint_story ss WHERE ss.sprint_id = sp.id) AS story_count`).
		Joins("JOIN project proj ON proj.id = sp.project_id AND proj.deleted = 0").
		Joins("LEFT JOIN product prod ON prod.id = proj.product_id").
		Where("sp.deleted = 0")
	if projectID > 0 {
		dataQ = dataQ.Where("sp.project_id = ?", projectID)
	}
	if productID > 0 {
		dataQ = dataQ.Where("proj.product_id = ?", productID)
	}
	if status != "" {
		dataQ = dataQ.Where("sp.status = ?", status)
	}

	var rows []row
	if err := dataQ.Order("sp.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	vos := make([]SprintVO, 0, len(rows))
	for _, r := range rows {
		vos = append(vos, SprintVO{
			Sprint: r.Sprint, ProjectName: r.ProjectName,
			ProductID: r.ProductID, ProductName: r.ProductName, StoryCount: r.StoryCount,
		})
	}
	return &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *SprintService) toVO(sp model.Sprint) SprintVO {
	vo := SprintVO{Sprint: sp}
	var proj model.Project
	if err := s.db.Select("id", "name", "product_id").Where("id = ? AND deleted = 0", sp.ProjectID).First(&proj).Error; err == nil {
		vo.ProjectName = proj.Name
		vo.ProductID = proj.ProductID
		var prod model.Product
		if err := s.db.Select("name").Where("id = ?", proj.ProductID).First(&prod).Error; err == nil {
			vo.ProductName = prod.Name
		}
	}
	s.db.Model(&model.SprintStory{}).Where("sprint_id = ?", sp.ID).Count(&vo.StoryCount)
	return vo
}

func (s *SprintService) Get(id uint64) (*SprintVO, error) {
	var sp model.Sprint
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("迭代不存在")
	}
	vo := s.toVO(sp)
	return &vo, nil
}

type CreateSprintInput struct {
	ProjectID uint64  `json:"projectId" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Begin     *string `json:"begin"`
	End       *string `json:"end"`
	Goal      *string `json:"goal"`
}

func (s *SprintService) Create(in CreateSprintInput) (*SprintVO, error) {
	var proj model.Project
	if err := s.db.Where("id = ? AND deleted = 0", in.ProjectID).First(&proj).Error; err != nil {
		return nil, fmt.Errorf("项目不存在")
	}
	begin, err := parseDatePtr(in.Begin)
	if err != nil {
		return nil, err
	}
	end, err := parseDatePtr(in.End)
	if err != nil {
		return nil, err
	}
	sp := model.Sprint{
		ProjectID: in.ProjectID,
		Name:      in.Name,
		Status:    "wait",
		Begin:     begin,
		End:       end,
		Goal:      in.Goal,
	}
	if err := s.db.Create(&sp).Error; err != nil {
		return nil, err
	}
	return s.Get(sp.ID)
}

type UpdateSprintInput struct {
	Name   *string `json:"name"`
	Begin  *string `json:"begin"`
	End    *string `json:"end"`
	Goal   *string `json:"goal"`
	Status *string `json:"status"`
}

var sprintStatusNext = map[string]map[string]bool{
	"wait":   {"doing": true, "closed": true},
	"doing":  {"done": true, "closed": true},
	"done":   {"doing": true, "closed": true},
	"closed": {},
}

func (s *SprintService) Update(id uint64, in UpdateSprintInput) (*SprintVO, error) {
	var sp model.Sprint
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("迭代不存在")
	}
	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Goal != nil {
		updates["goal"] = *in.Goal
	}
	if in.Begin != nil {
		t, err := parseDatePtr(in.Begin)
		if err != nil {
			return nil, err
		}
		updates["begin"] = t
	}
	if in.End != nil {
		t, err := parseDatePtr(in.End)
		if err != nil {
			return nil, err
		}
		updates["end"] = t
	}
	if in.Status != nil {
		next := *in.Status
		if next != sp.Status {
			allow := sprintStatusNext[sp.Status]
			if allow == nil || !allow[next] {
				return nil, fmt.Errorf("状态不允许从 %s 变为 %s", sp.Status, next)
			}
			updates["status"] = next
		}
	}
	if len(updates) > 0 {
		if err := s.db.Model(&sp).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

func (s *SprintService) Delete(id uint64) error {
	var cnt int64
	s.db.Model(&model.SprintStory{}).Where("sprint_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("迭代仍有关联需求，无法删除")
	}
	res := s.db.Model(&model.Sprint{}).Where("id = ? AND deleted = 0", id).Update("deleted", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("迭代不存在")
	}
	return nil
}

func (s *SprintService) ListByProject(projectID uint64) ([]SprintVO, error) {
	var list []model.Sprint
	if err := s.db.Where("project_id = ? AND deleted = 0", projectID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	vos := make([]SprintVO, 0, len(list))
	for _, sp := range list {
		vos = append(vos, s.toVO(sp))
	}
	return vos, nil
}

func (s *SprintService) ListStories(sprintID uint64) ([]StoryVO, error) {
	var storyIDs []uint64
	s.db.Model(&model.SprintStory{}).Where("sprint_id = ?", sprintID).Pluck("story_id", &storyIDs)
	if len(storyIDs) == 0 {
		return []StoryVO{}, nil
	}
	var list []model.Story
	if err := s.db.Where("id IN ? AND deleted = 0", storyIDs).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	storySvc := NewStoryService(s.db)
	vos := make([]StoryVO, 0, len(list))
	for _, st := range list {
		vos = append(vos, storySvc.toListVO(st))
	}
	return vos, nil
}

type LinkStoriesInput struct {
	StoryIDs []uint64 `json:"storyIds" binding:"required"`
}

func (s *SprintService) LinkStories(sprintID uint64, storyIDs []uint64) error {
	var sp model.Sprint
	if err := s.db.Where("id = ? AND deleted = 0", sprintID).First(&sp).Error; err != nil {
		return fmt.Errorf("迭代不存在")
	}
	if sp.Status != "doing" {
		return fmt.Errorf("迭代非进行中，不可关联需求")
	}
	var proj model.Project
	if err := s.db.Where("id = ? AND deleted = 0", sp.ProjectID).First(&proj).Error; err != nil {
		return fmt.Errorf("项目不存在")
	}
	for _, storyID := range storyIDs {
		var st model.Story
		if err := s.db.Where("id = ? AND deleted = 0", storyID).First(&st).Error; err != nil {
			return fmt.Errorf("需求 %d 不存在", storyID)
		}
		if st.ProductID != proj.ProductID {
			return fmt.Errorf("需求 %d 与迭代产品不一致", storyID)
		}
		if st.Type != "story" || st.Status != "active" {
			return fmt.Errorf("需求 %d 类型或状态不可拉入迭代", storyID)
		}
		var exist int64
		s.db.Model(&model.SprintStory{}).Where("sprint_id = ? AND story_id = ?", sprintID, storyID).Count(&exist)
		if exist > 0 {
			return fmt.Errorf("需求 %d 已在本迭代中", storyID)
		}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, storyID := range storyIDs {
			ss := model.SprintStory{
				ProjectID: sp.ProjectID,
				SprintID:  sprintID,
				ProductID: proj.ProductID,
				StoryID:   storyID,
			}
			if err := tx.Create(&ss).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SprintService) UnlinkStory(sprintID, storyID uint64) error {
	var sp model.Sprint
	if err := s.db.Where("id = ? AND deleted = 0", sprintID).First(&sp).Error; err != nil {
		return fmt.Errorf("迭代不存在")
	}
	if sp.Status != "doing" {
		return fmt.Errorf("迭代非进行中，不可移除需求")
	}
	res := s.db.Where("sprint_id = ? AND story_id = ?", sprintID, storyID).Delete(&model.SprintStory{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("关联不存在")
	}
	return nil
}

func (s *SprintService) StoryCandidates(sprintID uint64, page, pageSize int, keyword string) (*PageResult, error) {
	var sp model.Sprint
	if err := s.db.Where("id = ? AND deleted = 0", sprintID).First(&sp).Error; err != nil {
		return nil, fmt.Errorf("迭代不存在")
	}
	var proj model.Project
	if err := s.db.Where("id = ? AND deleted = 0", sp.ProjectID).First(&proj).Error; err != nil {
		return nil, fmt.Errorf("项目不存在")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var linkedIDs []uint64
	s.db.Model(&model.SprintStory{}).Where("sprint_id = ?", sprintID).Pluck("story_id", &linkedIDs)

	q := s.db.Model(&model.Story{}).Where("deleted = 0 AND product_id = ? AND type = ? AND status = ?",
		proj.ProductID, "story", "active")
	if len(linkedIDs) > 0 {
		q = q.Where("id NOT IN ?", linkedIDs)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var list []model.Story
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	storySvc := NewStoryService(s.db)
	vos := make([]StoryVO, 0, len(list))
	for _, st := range list {
		vos = append(vos, storySvc.toListVO(st))
	}
	return &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}, nil
}
