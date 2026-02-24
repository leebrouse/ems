package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/leebrouse/ems/backend/auth/model"
	"github.com/leebrouse/ems/backend/auth/repository"
	authjwt "github.com/leebrouse/ems/backend/common/auth"
	userpb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"
)

var (
	// ErrUnauthorized 表示认证失败
	ErrUnauthorized = errors.New("unauthorized")
)

// AuthService 定义认证相关业务能力
type AuthService interface {
	Login(ctx context.Context, username, password string) (string, string, *userpb.UserResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

// authService 是 AuthService 的默认实现
type authService struct {
	repo       repository.AuthRepository
	userClient userpb.UserServiceClient
	jwtHandler *authjwt.JWTHandler
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(repo repository.AuthRepository, userClient userpb.UserServiceClient, jwtHandler *authjwt.JWTHandler) AuthService {
	return &authService{
		repo:       repo,
		userClient: userClient,
		jwtHandler: jwtHandler,
	}
}

// Login 校验用户并签发访问/刷新令牌
func (s *authService) Login(ctx context.Context, username, password string) (string, string, *userpb.UserResponse, error) {
	// 1. Validate credentials via UserService
	log.Println("Start login,access to auth service.......")
	usr, err := s.userClient.ValidateCredentials(ctx, &userpb.ValidateCredentialsRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", nil, ErrUnauthorized
	}

	// 2. Generate Access Token
	accessToken, err := s.jwtHandler.GenerateToken(int64(usr.Id), usr.Username, usr.Roles)
	if err != nil {
		return "", "", nil, err
	}

	// 3. Generate Refresh Token
	refreshTokenStr := uuid.New().String()
	refreshToken := &model.RefreshToken{
		UserID:    int64(usr.Id),
		Token:     refreshTokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 days
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshTokenStr, usr, nil
}

// Logout 注销并撤销刷新令牌
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshToken)
}

// RefreshToken 校验并轮换刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	// 1. Get refresh token from DB
	rt, err := s.repo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil {
		return "", "", ErrUnauthorized
	}

	if rt.ExpiresAt.Before(time.Now()) {
		s.repo.RevokeRefreshToken(ctx, refreshTokenStr)
		return "", "", ErrUnauthorized
	}

	// 2. Get user info (to refresh roles/username)
	usr, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Id: int32(rt.UserID)})
	if err != nil {
		return "", "", ErrUnauthorized
	}

	// 3. Generate new Access Token
	accessToken, err := s.jwtHandler.GenerateToken(int64(usr.Id), usr.Username, usr.Roles)
	if err != nil {
		return "", "", err
	}

	// Optional: Rotate refresh token
	newRefreshTokenStr := uuid.New().String()
	newRefreshToken := &model.RefreshToken{
		UserID:    int64(usr.Id),
		Token:     newRefreshTokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7),
	}

	if err := s.repo.CreateRefreshToken(ctx, newRefreshToken); err != nil {
		return "", "", err
	}

	s.repo.RevokeRefreshToken(ctx, refreshTokenStr)

	return accessToken, newRefreshTokenStr, nil
}
