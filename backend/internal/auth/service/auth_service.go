package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leebrouse/ems/backend/auth/model"
	"github.com/leebrouse/ems/backend/auth/repository"
	authjwt "github.com/leebrouse/ems/backend/common/auth"
	userpb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, string, *userpb.UserResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

type authService struct {
	repo       repository.AuthRepository
	userClient userpb.UserServiceClient
	jwtHandler *authjwt.JWTHandler
}

func NewAuthService(repo repository.AuthRepository, userClient userpb.UserServiceClient, jwtHandler *authjwt.JWTHandler) AuthService {
	return &authService{
		repo:       repo,
		userClient: userClient,
		jwtHandler: jwtHandler,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, *userpb.UserResponse, error) {
	// 1. Validate credentials via UserService
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

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshToken)
}

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
