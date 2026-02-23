package repository

import (
	"context"

	"github.com/leebrouse/ems/backend/auth/model"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ? AND revoked = ?", token, false).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

func (r *authRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", gorm.Expr("NOW()")).Delete(&model.RefreshToken{}).Error
}
