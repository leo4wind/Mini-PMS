package service

import (
	"minipms/internal/model"

	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

type DashboardStoryItem struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	ProductID uint64 `json:"productId"`
}

type DashboardBugItem struct {
	ID        uint64 `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	ProductID uint64 `json:"productId"`
}

type DashboardSprintItem struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	ProjectID uint64 `json:"projectId"`
	Status    string `json:"status"`
}

type DashboardSummary struct {
	MyActiveStoryCount  int64                 `json:"myActiveStoryCount"`
	MyActiveBugCount    int64                 `json:"myActiveBugCount"`
	ProductCount        int64                 `json:"productCount"`
	ProjectDoingCount   int64                 `json:"projectDoingCount"`
	SprintDoingCount    int64                 `json:"sprintDoingCount"`
	OpenBugCount        int64                 `json:"openBugCount"`
	MyStories           []DashboardStoryItem  `json:"myStories"`
	MyBugs              []DashboardBugItem    `json:"myBugs"`
	DoingSprints        []DashboardSprintItem `json:"doingSprints"`
}

func (s *DashboardService) Summary(userID uint64) (*DashboardSummary, error) {
	out := &DashboardSummary{
		MyStories:    []DashboardStoryItem{},
		MyBugs:       []DashboardBugItem{},
		DoingSprints: []DashboardSprintItem{},
	}

	s.db.Model(&model.Story{}).Where("deleted = 0 AND assigned_to = ? AND status = ?", userID, "active").Count(&out.MyActiveStoryCount)
	s.db.Model(&model.Bug{}).Where("deleted = 0 AND assigned_to = ? AND status = ?", userID, "active").Count(&out.MyActiveBugCount)
	s.db.Model(&model.Product{}).Where("deleted = 0 AND status = ?", "normal").Count(&out.ProductCount)
	s.db.Model(&model.Project{}).Where("deleted = 0 AND status = ?", "doing").Count(&out.ProjectDoingCount)
	s.db.Model(&model.Sprint{}).Where("deleted = 0 AND status = ?", "doing").Count(&out.SprintDoingCount)
	s.db.Model(&model.Bug{}).Where("deleted = 0 AND status = ?", "active").Count(&out.OpenBugCount)

	var stories []model.Story
	s.db.Where("deleted = 0 AND assigned_to = ? AND status = ?", userID, "active").
		Order("updated_at DESC").Limit(5).Find(&stories)
	for _, st := range stories {
		out.MyStories = append(out.MyStories, DashboardStoryItem{
			ID: st.ID, Title: st.Title, Status: st.Status, ProductID: st.ProductID,
		})
	}

	var bugs []model.Bug
	s.db.Where("deleted = 0 AND assigned_to = ? AND status = ?", userID, "active").
		Order("updated_at DESC").Limit(5).Find(&bugs)
	for _, b := range bugs {
		out.MyBugs = append(out.MyBugs, DashboardBugItem{
			ID: b.ID, Title: b.Title, Status: b.Status, ProductID: b.ProductID,
		})
	}

	var sprints []model.Sprint
	s.db.Where("deleted = 0 AND status = ?", "doing").Order("updated_at DESC").Limit(5).Find(&sprints)
	for _, sp := range sprints {
		out.DoingSprints = append(out.DoingSprints, DashboardSprintItem{
			ID: sp.ID, Name: sp.Name, ProjectID: sp.ProjectID, Status: sp.Status,
		})
	}

	return out, nil
}
