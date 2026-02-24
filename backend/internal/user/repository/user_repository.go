package repository

import (
	"context"

	"github.com/leebrouse/ems/backend/user/model"
	"gorm.io/gorm"
)

// UserRepository 定义用户数据访问接口
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id int64) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, size int) ([]model.User, int64, error)
	ListRoles(ctx context.Context) ([]model.Role, error)

	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	CreateRole(ctx context.Context, role *model.Role) error
}

// userRepository 使用 GORM 实现 UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser 保存用户
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUser 根据 ID 获取用户并预加载角色
func (r *userRepository) GetUser(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户并预加载角色
func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户
func (r *userRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser 删除用户
func (r *userRepository) DeleteUser(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("user_roles").Where("user_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, id).Error
	})
}

// ListUsers 分页查询用户列表
func (r *userRepository) ListUsers(ctx context.Context, page, size int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	db := r.db.WithContext(ctx).Model(&model.User{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Roles").Offset((page - 1) * size).Limit(size).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// GetRoleByName 根据名称获取角色
func (r *userRepository) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles 获取角色列表
func (r *userRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// CreateRole 创建角色
func (r *userRepository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}
