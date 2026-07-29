package service

import (
	"errors"
	"fmt"

	"minipms/internal/model"

	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

type PageResult struct {
	List     interface{} `json:"list"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Total    int64       `json:"total"`
	Meta     interface{} `json:"meta,omitempty"`
}

type ProductVO struct {
	model.Product
	Creator *UserBrief `json:"creator,omitempty"`
}

func (s *ProductService) List(page, pageSize int, status, keyword string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	type row struct {
		model.Product
		CreatorAccount *string `gorm:"column:creator_account"`
		CreatorName    *string `gorm:"column:creator_name"`
		Total          int64   `gorm:"column:total_count"`
	}

	where := "p.deleted = 0"
	args := make([]interface{}, 0, 8)
	if status != "" {
		where += " AND p.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (p.name LIKE ? OR p.code LIKE ?)"
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	sql := `SELECT p.*,
			cu.account AS creator_account, cu.realname AS creator_name,
			COUNT(*) OVER() AS total_count
		FROM product p
		LEFT JOIN ` + "`user`" + ` cu ON cu.id = p.created_by AND cu.deleted = 0
		WHERE ` + where + `
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	var rows []row
	if err := s.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var total int64
	list := make([]ProductVO, 0, len(rows))
	for _, r := range rows {
		total = r.Total
		vo := ProductVO{Product: r.Product}
		vo.Creator = &UserBrief{ID: r.Product.CreatedBy}
		if r.CreatorAccount != nil {
			vo.Creator.Account = *r.CreatorAccount
		}
		if r.CreatorName != nil {
			vo.Creator.Realname = *r.CreatorName
		}
		list = append(list, vo)
	}
	return &PageResult{List: list, Page: page, PageSize: pageSize, Total: total}, nil
}

type CreateProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Code        *string `json:"code"`
	PO          *uint64 `json:"po"`
	Description *string `json:"description"`
}

type CreateProductResult struct {
	Product        model.Product `json:"product"`
	DefaultProject model.Project `json:"defaultProject"`
}

func (s *ProductService) Create(userID uint64, in CreateProductInput) (*CreateProductResult, error) {
	var out CreateProductResult
	err := s.db.Transaction(func(tx *gorm.DB) error {
		p := model.Product{
			Name:        in.Name,
			Code:        in.Code,
			PO:          in.PO,
			Description: in.Description,
			Status:      "normal",
			CreatedBy:   userID,
		}
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		projName := in.Name + "1.0"
		var projCode *string
		if in.Code != nil && *in.Code != "" {
			c := *in.Code + "-1.0"
			projCode = &c
		}
		proj := model.Project{
			ProductID: p.ID,
			Name:      projName,
			Code:      projCode,
			Status:    "wait",
			CreatedBy: userID,
		}
		if err := tx.Create(&proj).Error; err != nil {
			// code conflict: retry without code
			if projCode != nil {
				proj.Code = nil
				if err2 := tx.Create(&proj).Error; err2 != nil {
					return err2
				}
			} else {
				return err
			}
		}
		out.Product = p
		out.DefaultProject = proj
		return nil
	})
	return &out, err
}

func (s *ProductService) Get(id uint64) (*model.Product, error) {
	var p model.Product
	err := s.db.Where("id = ? AND deleted = 0", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("产品不存在")
	}
	return &p, err
}

func (s *ProductService) Update(id uint64, updates map[string]interface{}) error {
	return s.db.Model(&model.Product{}).Where("id = ? AND deleted = 0", id).Updates(updates).Error
}

func (s *ProductService) Delete(id uint64) error {
	var projCount, storyCount int64
	s.db.Model(&model.Project{}).Where("product_id = ? AND deleted = 0", id).Count(&projCount)
	s.db.Model(&model.Story{}).Where("product_id = ? AND deleted = 0", id).Count(&storyCount)
	if projCount > 0 || storyCount > 0 {
		return fmt.Errorf("存在未删除的项目或需求，无法删除产品")
	}
	return s.db.Model(&model.Product{}).Where("id = ?", id).Update("deleted", 1).Error
}
