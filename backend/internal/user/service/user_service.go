package service

import (
	"context"
	"errors"

	"github.com/leebrouse/ems/backend/user/model"
	"github.com/leebrouse/ems/backend/user/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUserNotFound 表示用户不存在
	ErrUserNotFound   = errors.New("user not found")
	// ErrUsernameExists 表示用户名已存在
	ErrUsernameExists = errors.New("username already exists")
)

// UserService 定义用户管理业务能力
type UserService interface {
	CreateUser(ctx context.Context, username, password string, roles []string) (*model.User, error)
	GetUser(ctx context.Context, id int64) (*model.User, error)
	UpdateUser(ctx context.Context, id int64, password *string, roles *[]string) (*model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, size int) ([]model.User, int64, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	ValidatePassword(ctx context.Context, username, password string) (*model.User, error)
}

// userService 是 UserService 的默认实现
type userService struct {
	repo repository.UserRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateUser 创建新用户并初始化角色
func (s *userService) CreateUser(ctx context.Context, username, password string, roles []string) (*model.User, error) {
	// 1. Check if user exists
	_, err := s.repo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, ErrUsernameExists
	}

	// 2. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Prepare roles
	var userRoles []model.Role
	for _, rName := range roles {
		role, err := s.repo.GetRoleByName(ctx, rName)
		if err != nil {
			// Create role if not exists? For simple demo, let's create it.
			role = &model.Role{Name: rName}
			s.repo.CreateRole(ctx, role)
		}
		userRoles = append(userRoles, *role)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Roles:        userRoles,
		IsActive:     true,
	}

	// save into db
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser 根据 ID 获取用户
func (s *userService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetUser(ctx, id)
}

// UpdateUser 更新用户密码或角色
func (s *userService) UpdateUser(ctx context.Context, id int64, password *string, roles *[]string) (*model.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hash)
	}

	if roles != nil {
		var userRoles []model.Role
		for _, rName := range *roles {
			role, err := s.repo.GetRoleByName(ctx, rName)
			if err != nil {
				role = &model.Role{Name: rName}
				s.repo.CreateRole(ctx, role)
			}
			userRoles = append(userRoles, *role)
		}
		user.Roles = userRoles
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.DeleteUser(ctx, id)
}

// ListUsers 分页查询用户列表
func (s *userService) ListUsers(ctx context.Context, page, size int) ([]model.User, int64, error) {
	return s.repo.ListUsers(ctx, page, size)
}

// ListRoles 查询角色列表
func (s *userService) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.repo.ListRoles(ctx)
}

// ValidatePassword 校验用户名与密码
func (s *userService) ValidatePassword(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return user, nil
}
