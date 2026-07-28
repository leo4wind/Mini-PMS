package service

import (
	"errors"
	"fmt"

	"minipms/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

var (
	ErrBadCredential = errors.New("账号或密码错误")
	ErrUserDisabled  = errors.New("账号已停用")
)

func (s *AuthService) Login(account, password string) (*model.User, error) {
	var user model.User
	err := s.db.Where("account = ? AND deleted = 0", account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBadCredential
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, ErrUserDisabled
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrBadCredential
	}
	return &user, nil
}

func (s *AuthService) GetUser(id uint64) (*model.User, error) {
	var user model.User
	err := s.db.Where("id = ? AND deleted = 0", id).First(&user).Error
	return &user, err
}

func (s *AuthService) ChangePassword(userID uint64, oldPwd, newPwd string) error {
	user, err := s.GetUser(userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)) != nil {
		return fmt.Errorf("旧密码不正确")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", string(hash)).Error
}

func HashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(b), err
}
