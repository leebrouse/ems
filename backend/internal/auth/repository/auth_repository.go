package repository

import (
	"context"

	"github.com/leebrouse/ems/backend/auth/model"
	"gorm.io/gorm"
)

// AuthRepository 定义刷新令牌的数据访问接口
type AuthRepository interface {
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error)
	// RevokeRefreshToken 将刷新令牌标记为已撤销
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
}

// authRepository 使用 GORM 实现 AuthRepository
type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository 创建 AuthRepository 实例
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

// CreateRefreshToken 保存刷新令牌
func (r *authRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetRefreshToken 获取未撤销的刷新令牌
func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ? AND revoked = ?", token, false).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

// RevokeRefreshToken 撤销刷新令牌
func (r *authRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

// DeleteExpiredTokens 清理过期的刷新令牌
func (r *authRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", gorm.Expr("NOW()")).Delete(&model.RefreshToken{}).Error
}
