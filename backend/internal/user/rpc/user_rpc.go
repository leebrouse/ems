package rpc

import (
	"context"
	"log"

	pb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"
	"github.com/leebrouse/ems/backend/user/model"
	"github.com/leebrouse/ems/backend/user/service"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserRPCServer 提供用户相关的 gRPC 接口实现
type UserRPCServer struct {
	pb.UnimplementedUserServiceServer
	svc service.UserService
}

// NewUserRPCServer 创建 UserRPCServer 实例
func NewUserRPCServer(svc service.UserService) *UserRPCServer {
	return &UserRPCServer{svc: svc}
}

// CreateUser 创建用户
func (s *UserRPCServer) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.UserResponse, error) {
	u, err := s.svc.CreateUser(ctx, in.Username, in.Password, in.Roles)
	if err != nil {
		if err == service.ErrUsernameExists {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toUserResponse(u), nil
}

// GetUser 根据 ID 获取用户
func (s *UserRPCServer) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.UserResponse, error) {
	u, err := s.svc.GetUser(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return s.toUserResponse(u), nil
}

// ListUsers 分页查询用户列表
func (s *UserRPCServer) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := s.svc.ListUsers(ctx, int(in.Page), int(in.Size))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.UserResponse
	for _, u := range users {
		res = append(res, s.toUserResponse(&u))
	}
	return &pb.ListUsersResponse{
		Total: int32(total),
		Users: res,
	}, nil
}

// UpdateUser 更新用户密码或角色
func (s *UserRPCServer) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	var pwd *string
	if in.Password != "" {
		pwd = &in.Password
	}
	var roles *[]string
	if len(in.Roles) > 0 {
		roles = &in.Roles
	}

	u, err := s.svc.UpdateUser(ctx, int64(in.Id), pwd, roles)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toUserResponse(u), nil
}

// DeleteUser 删除用户
func (s *UserRPCServer) DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*empty.Empty, error) {
	if err := s.svc.DeleteUser(ctx, int64(in.Id)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

// ValidateCredentials 校验用户名与密码
func (s *UserRPCServer) ValidateCredentials(ctx context.Context, in *pb.ValidateCredentialsRequest) (*pb.UserResponse, error) {
	log.Println("Access user service,call ValidateCredentials by rpc")
	u, err := s.svc.ValidatePassword(ctx, in.Username, in.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return s.toUserResponse(u), nil
}

// toUserResponse 将模型转换为 gRPC 响应
func (s *UserRPCServer) toUserResponse(u *model.User) *pb.UserResponse {
	var roles []string
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return &pb.UserResponse{
		Id:       int32(u.ID),
		Username: u.Username,
		Roles:    roles,
	}
}
