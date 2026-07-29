package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"minipms/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SystemService struct {
	db *gorm.DB
}

func NewSystemService(db *gorm.DB) *SystemService {
	return &SystemService{db: db}
}

type UserVO struct {
	model.User
	Roles []model.Role `json:"roles"`
}

func (s *SystemService) ListUsers(page, pageSize int, status, keyword string) (*PageResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	countQ := s.db.Table("`user`").Where("deleted = 0")
	if status != "" {
		countQ = countQ.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		countQ = countQ.Where("account LIKE ? OR realname LIKE ?", like, like)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, err
	}

	// 一条 SQL 查出用户 + 角色（JSON 聚合）
	type row struct {
		model.User
		RolesJSON []byte `gorm:"column:roles_json"`
	}
	dataQ := s.db.Table("`user` u").
		Select(`u.id, u.account, u.password_hash, u.realname, u.email, u.status, u.created_at, u.updated_at, u.deleted,
			COALESCE((
				SELECT JSON_ARRAYAGG(JSON_OBJECT(
					'id', r.id,
					'code', r.code,
					'name', r.name,
					'builtin', r.builtin,
					'remark', r.remark
				))
				FROM user_role ur
				JOIN role r ON r.id = ur.role_id
				WHERE ur.user_id = u.id
			), JSON_ARRAY()) AS roles_json`).
		Where("u.deleted = 0")
	if status != "" {
		dataQ = dataQ.Where("u.status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		dataQ = dataQ.Where("u.account LIKE ? OR u.realname LIKE ?", like, like)
	}

	var rows []row
	if err := dataQ.Order("u.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error; err != nil {
		return nil, err
	}
	list := make([]UserVO, 0, len(rows))
	for _, r := range rows {
		roles := []model.Role{}
		if len(r.RolesJSON) > 0 && string(r.RolesJSON) != "null" {
			_ = json.Unmarshal(r.RolesJSON, &roles)
		}
		if roles == nil {
			roles = []model.Role{}
		}
		list = append(list, UserVO{User: r.User, Roles: roles})
	}
	return &PageResult{List: list, Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *SystemService) rolesOfUser(userID uint64) ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Table("role r").
		Joins("JOIN user_role ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (s *SystemService) GetUser(id uint64) (*UserVO, error) {
	var u model.User
	if err := s.db.Where("id = ? AND deleted = 0", id).First(&u).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	roles, err := s.rolesOfUser(id)
	if err != nil {
		return nil, err
	}
	return &UserVO{User: u, Roles: roles}, nil
}

const DefaultUserPassword = "123456"

type CreateUserInput struct {
	Account  string   `json:"account" binding:"required"`
	Password string   `json:"password"` // 空则使用 DefaultUserPassword
	Realname string   `json:"realname" binding:"required"`
	Email    *string  `json:"email"`
	RoleIDs  []uint64 `json:"roleIds"`
}

func (s *SystemService) CreateUser(in CreateUserInput) (*UserVO, error) {
	var cnt int64
	s.db.Model(&model.User{}).Where("account = ? AND deleted = 0", in.Account).Count(&cnt)
	if cnt > 0 {
		return nil, fmt.Errorf("账号已存在")
	}
	pwd := strings.TrimSpace(in.Password)
	if pwd == "" {
		pwd = DefaultUserPassword
	}
	if len(pwd) < 6 {
		return nil, fmt.Errorf("密码至少 6 位")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := model.User{
		Account:      in.Account,
		PasswordHash: string(hash),
		Realname:     in.Realname,
		Email:        in.Email,
		Status:       "active",
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		return s.replaceUserRolesTx(tx, u.ID, in.RoleIDs)
	})
	if err != nil {
		return nil, err
	}
	return s.GetUser(u.ID)
}

func (s *SystemService) UpdateUser(id uint64, realname *string, email *string, status *string) (*UserVO, error) {
	updates := map[string]interface{}{}
	if realname != nil {
		updates["realname"] = *realname
	}
	if email != nil {
		updates["email"] = *email
	}
	if status != nil {
		if *status != "active" && *status != "disabled" {
			return nil, fmt.Errorf("非法状态")
		}
		updates["status"] = *status
	}
	if len(updates) == 0 {
		return s.GetUser(id)
	}
	res := s.db.Model(&model.User{}).Where("id = ? AND deleted = 0", id).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("用户不存在")
	}
	return s.GetUser(id)
}

func (s *SystemService) SetUserStatus(id uint64, status string) error {
	if status != "active" && status != "disabled" {
		return fmt.Errorf("非法状态")
	}
	res := s.db.Model(&model.User{}).Where("id = ? AND deleted = 0", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

func (s *SystemService) AssignUserRoles(userID uint64, roleIDs []uint64) error {
	if _, err := s.GetUser(userID); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.replaceUserRolesTx(tx, userID, roleIDs)
	})
}

func (s *SystemService) replaceUserRolesTx(tx *gorm.DB, userID uint64, roleIDs []uint64) error {
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return err
	}
	seen := map[uint64]struct{}{}
	for _, rid := range roleIDs {
		if _, ok := seen[rid]; ok {
			continue
		}
		seen[rid] = struct{}{}
		var cnt int64
		tx.Model(&model.Role{}).Where("id = ?", rid).Count(&cnt)
		if cnt == 0 {
			return fmt.Errorf("角色不存在: %d", rid)
		}
		if err := tx.Create(&model.UserRole{UserID: userID, RoleID: rid}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ---------- Role ----------

func (s *SystemService) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := s.db.Order("id ASC").Find(&roles).Error
	return roles, err
}

func (s *SystemService) UpdateRole(id uint64, name, remark *string) (*model.Role, error) {
	var role model.Role
	if err := s.db.First(&role, id).Error; err != nil {
		return nil, fmt.Errorf("角色不存在")
	}
	updates := map[string]interface{}{}
	if name != nil {
		updates["name"] = *name
	}
	if remark != nil {
		updates["remark"] = *remark
	}
	if len(updates) > 0 {
		if err := s.db.Model(&role).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	_ = s.db.First(&role, id)
	return &role, nil
}

func (s *SystemService) GetRoleMenuIDs(roleID uint64) ([]uint64, error) {
	var ids []uint64
	err := s.db.Model(&model.RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}

func (s *SystemService) AssignRoleMenus(roleID uint64, menuIDs []uint64) error {
	var role model.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("角色不存在")
	}
	// F-SYS-08: manager 必须至少保留 system 相关菜单，避免锁死
	if role.Code == "manager" {
		if err := s.ensureManagerMenus(menuIDs); err != nil {
			return err
		}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		seen := map[uint64]struct{}{}
		for _, mid := range menuIDs {
			if _, ok := seen[mid]; ok {
				continue
			}
			seen[mid] = struct{}{}
			var cnt int64
			tx.Model(&model.Menu{}).Where("id = ?", mid).Count(&cnt)
			if cnt == 0 {
				return fmt.Errorf("菜单不存在: %d", mid)
			}
			if err := tx.Create(&model.RoleMenu{RoleID: roleID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SystemService) ensureManagerMenus(menuIDs []uint64) error {
	needCodes := []string{
		"system", "user.list", "role.list", "menu.list",
		"role.assignMenu", "user.assignRole",
	}
	var menus []model.Menu
	if err := s.db.Where("code IN ?", needCodes).Find(&menus).Error; err != nil {
		return err
	}
	have := map[uint64]struct{}{}
	for _, id := range menuIDs {
		have[id] = struct{}{}
	}
	for _, m := range menus {
		if _, ok := have[m.ID]; !ok {
			return fmt.Errorf("经理角色必须保留系统管理相关菜单权限（缺少 %s）", m.Code)
		}
	}
	return nil
}

// ---------- Menu ----------

func (s *SystemService) AllMenuTree() ([]*MenuNode, error) {
	var menus []model.Menu
	if err := s.db.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

func (s *SystemService) FlatMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := s.db.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

type MenuInput struct {
	ParentID *uint64 `json:"parentId"`
	Code     string  `json:"code" binding:"required"`
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required"`
	Path     *string `json:"path"`
	Icon     *string `json:"icon"`
	Sort     *int    `json:"sort"`
	Status   *string `json:"status"`
}

func (s *SystemService) CreateMenu(in MenuInput) (*model.Menu, error) {
	if in.Type != "dir" && in.Type != "menu" && in.Type != "button" {
		return nil, fmt.Errorf("type 必须是 dir/menu/button")
	}
	var cnt int64
	s.db.Model(&model.Menu{}).Where("code = ?", in.Code).Count(&cnt)
	if cnt > 0 {
		return nil, fmt.Errorf("菜单编码已存在")
	}
	if in.ParentID != nil {
		var p model.Menu
		if err := s.db.First(&p, *in.ParentID).Error; err != nil {
			return nil, fmt.Errorf("父菜单不存在")
		}
	}
	sort := 0
	if in.Sort != nil {
		sort = *in.Sort
	}
	status := "enabled"
	if in.Status != nil {
		status = *in.Status
	}
	m := model.Menu{
		ParentID: in.ParentID,
		Code:     in.Code,
		Name:     in.Name,
		Type:     in.Type,
		Path:     in.Path,
		Icon:     in.Icon,
		Sort:     sort,
		Status:   status,
	}
	if err := s.db.Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *SystemService) UpdateMenu(id uint64, in MenuInput) (*model.Menu, error) {
	var m model.Menu
	if err := s.db.First(&m, id).Error; err != nil {
		return nil, fmt.Errorf("菜单不存在")
	}
	updates := map[string]interface{}{}
	if in.Name != "" {
		updates["name"] = in.Name
	}
	if in.Type != "" {
		if in.Type != "dir" && in.Type != "menu" && in.Type != "button" {
			return nil, fmt.Errorf("type 必须是 dir/menu/button")
		}
		updates["type"] = in.Type
	}
	if in.Code != "" && in.Code != m.Code {
		var cnt int64
		s.db.Model(&model.Menu{}).Where("code = ? AND id <> ?", in.Code, id).Count(&cnt)
		if cnt > 0 {
			return nil, fmt.Errorf("菜单编码已存在")
		}
		updates["code"] = in.Code
	}
	if in.Path != nil {
		updates["path"] = *in.Path
	}
	if in.Icon != nil {
		updates["icon"] = *in.Icon
	}
	if in.Sort != nil {
		updates["sort"] = *in.Sort
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	// allow clearing/changing parent
	updates["parent_id"] = in.ParentID
	if in.ParentID != nil && *in.ParentID == id {
		return nil, fmt.Errorf("不能将自己设为父菜单")
	}
	if err := s.db.Model(&m).Updates(updates).Error; err != nil {
		return nil, err
	}
	_ = s.db.First(&m, id)
	return &m, nil
}

func (s *SystemService) DeleteMenu(id uint64) error {
	var child int64
	s.db.Model(&model.Menu{}).Where("parent_id = ?", id).Count(&child)
	if child > 0 {
		return errors.New("存在子菜单，无法删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.Menu{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("菜单不存在")
		}
		return nil
	})
}
