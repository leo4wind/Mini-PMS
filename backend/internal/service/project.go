package service

import (
	"fmt"
	"time"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

type ProjectVO struct {
	model.Project
	ProductName string     `json:"productName"`
	PMUser      *UserBrief `json:"pmUser,omitempty"`
	Creator     *UserBrief `json:"creator,omitempty"`
	SprintCount int64      `json:"sprintCount"`
}

type UserBrief struct {
	ID       uint64 `json:"id"`
	Account  string `json:"account"`
	Realname string `json:"realname"`
}

func (s *ProjectService) List(page, pageSize int, productID uint64, status, keyword string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	type row struct {
		ID          uint64     `gorm:"column:id"`
		ProductID   uint64     `gorm:"column:product_id"`
		Name        string     `gorm:"column:name"`
		Code        *string    `gorm:"column:code"`
		Status      string     `gorm:"column:status"`
		Begin       *time.Time `gorm:"column:begin"`
		End         *time.Time `gorm:"column:end"`
		PM          *uint64    `gorm:"column:pm"`
		Description *string    `gorm:"column:description"`
		CreatedBy   uint64     `gorm:"column:created_by"`
		CreatedAt   time.Time  `gorm:"column:created_at"`
		UpdatedAt   time.Time  `gorm:"column:updated_at"`
		Deleted     uint8      `gorm:"column:deleted"`
		ProductName string     `gorm:"column:product_name"`
		PMAccount   *string    `gorm:"column:pm_account"`
		PMRealname  *string    `gorm:"column:pm_realname"`
		CreatorAccount *string `gorm:"column:creator_account"`
		CreatorName    *string `gorm:"column:creator_name"`
		SprintCount int64      `gorm:"column:sprint_count"`
		Total       int64      `gorm:"column:total_count"`
	}

	where := "p.deleted = 0"
	args := make([]interface{}, 0, 8)
	if productID > 0 {
		where += " AND p.product_id = ?"
		args = append(args, productID)
	}
	if status != "" {
		where += " AND p.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (p.name LIKE ? OR p.code LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	sql := `SELECT p.*, prod.name AS product_name,
			pm.account AS pm_account, pm.realname AS pm_realname,
			cu.account AS creator_account, cu.realname AS creator_name,
			COALESCE(sc.sprint_count, 0) AS sprint_count,
			COUNT(*) OVER() AS total_count
		FROM project p
		LEFT JOIN product prod ON prod.id = p.product_id
		LEFT JOIN ` + "`user`" + ` pm ON pm.id = p.pm AND pm.deleted = 0
		LEFT JOIN ` + "`user`" + ` cu ON cu.id = p.created_by AND cu.deleted = 0
		LEFT JOIN (
			SELECT project_id, COUNT(*) AS sprint_count
			FROM sprint WHERE deleted = 0
			GROUP BY project_id
		) sc ON sc.project_id = p.id
		WHERE ` + where + `
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	var rows []row
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var total int64
	vos := make([]ProjectVO, 0, len(rows))
	for _, r := range rows {
		total = r.Total
		p := model.Project{
			ID: r.ID, ProductID: r.ProductID, Name: r.Name, Code: r.Code, Status: r.Status,
			Begin: r.Begin, End: r.End, PM: r.PM, Description: r.Description,
			CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Deleted: r.Deleted,
		}
		vo := ProjectVO{Project: p, ProductName: r.ProductName, SprintCount: r.SprintCount}
		vo.Creator = &UserBrief{ID: r.CreatedBy}
		if r.CreatorAccount != nil {
			vo.Creator.Account = *r.CreatorAccount
		}
		if r.CreatorName != nil {
			vo.Creator.Realname = *r.CreatorName
		}
		if r.PM != nil {
			vo.PMUser = &UserBrief{ID: *r.PM}
			if r.PMAccount != nil {
				vo.PMUser.Account = *r.PMAccount
			}
			if r.PMRealname != nil {
				vo.PMUser.Realname = *r.PMRealname
			}
		}
		vos = append(vos, vo)
	}
	return &PageResult{List: vos, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *ProjectService) toVO(p model.Project) ProjectVO {
	vo := ProjectVO{Project: p}
	var prod model.Product
	if err := s.db.Select("id", "name").Where("id = ?", p.ProductID).First(&prod).Error; err == nil {
		vo.ProductName = prod.Name
	}
	if p.PM != nil {
		var u model.User
		if err := s.db.Select("id", "account", "realname").Where("id = ? AND deleted = 0", *p.PM).First(&u).Error; err == nil {
			vo.PMUser = &UserBrief{ID: u.ID, Account: u.Account, Realname: u.Realname}
		}
	}
	s.db.Model(&model.Sprint{}).Where("project_id = ? AND deleted = 0", p.ID).Count(&vo.SprintCount)
	return vo
}

func (s *ProjectService) Get(id uint64) (*ProjectVO, error) {
	var p model.Project
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&p).Error; err != nil {
		return nil, fmt.Errorf("项目不存在")
	}
	vo := s.toVO(p)
	return &vo, nil
}

type CreateProjectInput struct {
	ProductID   uint64  `json:"productId" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Code        *string `json:"code"`
	Begin       *string `json:"begin"` // YYYY-MM-DD
	End         *string `json:"end"`
	PM          *uint64 `json:"pm"`
	Description *string `json:"description"`
}

func parseDatePtr(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", *s, time.Local)
	if err != nil {
		return nil, fmt.Errorf("日期格式应为 YYYY-MM-DD")
	}
	return &t, nil
}

func (s *ProjectService) Create(userID uint64, in CreateProjectInput) (*ProjectVO, error) {
	var prod model.Product
	if err := s.db.Where("id = ? AND deleted = 0", in.ProductID).First(&prod).Error; err != nil {
		return nil, fmt.Errorf("产品不存在")
	}
	if prod.Status == "closed" {
		return nil, fmt.Errorf("产品已关闭，禁止新建项目")
	}
	begin, err := parseDatePtr(in.Begin)
	if err != nil {
		return nil, err
	}
	end, err := parseDatePtr(in.End)
	if err != nil {
		return nil, err
	}
	p := model.Project{
		ProductID:   in.ProductID,
		Name:        in.Name,
		Code:        in.Code,
		Status:      "wait",
		Begin:       begin,
		End:         end,
		PM:          in.PM,
		Description: in.Description,
		CreatedBy:   userID,
	}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	return s.Get(p.ID)
}

type UpdateProjectInput struct {
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Begin       *string `json:"begin"`
	End         *string `json:"end"`
	PM          *uint64 `json:"pm"`
	ClearPM     bool    `json:"clearPm"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

var projectStatusNext = map[string]map[string]bool{
	"wait":      {"doing": true, "closed": true},
	"doing":     {"suspended": true, "closed": true},
	"suspended": {"doing": true, "closed": true},
	"closed":    {},
}

func (s *ProjectService) Update(id uint64, in UpdateProjectInput) (*ProjectVO, error) {
	var p model.Project
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&p).Error; err != nil {
		return nil, fmt.Errorf("项目不存在")
	}
	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Code != nil {
		updates["code"] = *in.Code
	}
	if in.Description != nil {
		updates["description"] = *in.Description
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
	if in.ClearPM {
		updates["pm"] = nil
	} else if in.PM != nil {
		updates["pm"] = *in.PM
	}
	if in.Status != nil {
		next := *in.Status
		if next != p.Status {
			allow := projectStatusNext[p.Status]
			if allow == nil || !allow[next] {
				return nil, fmt.Errorf("状态不允许从 %s 变为 %s", p.Status, next)
			}
			updates["status"] = next
		}
	}
	if len(updates) > 0 {
		if err := s.db.Model(&p).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

func (s *ProjectService) Delete(id uint64) error {
	var cnt int64
	s.db.Model(&model.Sprint{}).Where("project_id = ? AND deleted = 0", id).Count(&cnt)
	if cnt > 0 {
		return fmt.Errorf("存在未删除的迭代，无法删除项目")
	}
	res := s.db.Model(&model.Project{}).Where("id = ? AND deleted = 0", id).Update("deleted", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("项目不存在")
	}
	return nil
}

func (s *ProjectService) ListByProduct(productID uint64) ([]ProjectVO, error) {
	var list []model.Project
	if err := s.db.Where("product_id = ? AND deleted = 0", productID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	vos := make([]ProjectVO, 0, len(list))
	for _, p := range list {
		vos = append(vos, s.toVO(p))
	}
	return vos, nil
}
