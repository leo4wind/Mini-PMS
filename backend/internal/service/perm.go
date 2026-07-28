package service

import (
	"minipms/internal/model"

	"gorm.io/gorm"
)

type PermService struct {
	db *gorm.DB
}

func NewPermService(db *gorm.DB) *PermService {
	return &PermService{db: db}
}

func (s *PermService) HasCode(userID uint64, code string) (bool, error) {
	var count int64
	err := s.db.Table("role_menu rm").
		Joins("JOIN user_role ur ON ur.role_id = rm.role_id").
		Joins("JOIN menu m ON m.id = rm.menu_id").
		Where("ur.user_id = ? AND m.code = ? AND m.status = ?", userID, code, "enabled").
		Count(&count).Error
	return count > 0, err
}

type MenuNode struct {
	ID       uint64      `json:"id"`
	ParentID *uint64     `json:"parentId"`
	Code     string      `json:"code"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Path     *string     `json:"path"`
	Icon     *string     `json:"icon"`
	Sort     int         `json:"sort"`
	Children []*MenuNode `json:"children,omitempty"`
}

func (s *PermService) MenuTreeForUser(userID uint64) ([]*MenuNode, error) {
	var menus []model.Menu
	err := s.db.Table("menu m").
		Select("DISTINCT m.*").
		Joins("JOIN role_menu rm ON rm.menu_id = m.id").
		Joins("JOIN user_role ur ON ur.role_id = rm.role_id").
		Where("ur.user_id = ? AND m.status = ?", userID, "enabled").
		Order("m.sort ASC, m.id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

func buildMenuTree(menus []model.Menu) []*MenuNode {
	byID := map[uint64]*MenuNode{}
	for _, m := range menus {
		node := &MenuNode{
			ID: m.ID, ParentID: m.ParentID, Code: m.Code, Name: m.Name, Type: m.Type,
			Path: m.Path, Icon: m.Icon, Sort: m.Sort,
		}
		byID[m.ID] = node
	}
	var roots []*MenuNode
	for _, m := range menus {
		node := byID[m.ID]
		if m.ParentID != nil {
			if p, ok := byID[*m.ParentID]; ok {
				p.Children = append(p.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}
	return roots
}

func (s *PermService) RolesOfUser(userID uint64) ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Table("role r").
		Joins("JOIN user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}
