package main

import (
	"context"
	"log"
	"net"

	"github.com/leebrouse/ems/backend/common/database"
	"github.com/leebrouse/ems/backend/common/genopenapi/user"
	pb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"
	"github.com/leebrouse/ems/backend/common/observation"
	"github.com/leebrouse/ems/backend/user/handler"
	"github.com/leebrouse/ems/backend/user/model"
	"github.com/leebrouse/ems/backend/user/repository"
	"github.com/leebrouse/ems/backend/user/rpc"
	"github.com/leebrouse/ems/backend/user/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化用户服务依赖并启动 gRPC/REST 服务
func main() {
	shutdown, err := observation.InitFromViper(context.Background(), "user")
	if err != nil {
		log.Fatalf("failed to init observation: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	// 1. Initialize Database
	db := database.Connect("service.user.postgres",
		&model.User{},
		&model.Role{},
	)

	// 2. Initialize Layers
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)
	r := rpc.NewUserRPCServer(svc)

	// 3. Start Servers
	go startGRPCServer(r)

	startRESTServer(h)
}

// startGRPCServer 启动用户 gRPC 服务
func startGRPCServer(r *rpc.UserRPCServer) {
	port := viper.GetString("service.user.grpc")
	if port == "" {
		port = "9001"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(observation.GRPCServerOptions()...)
	pb.RegisterUserServiceServer(s, r)
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startRESTServer 启动用户 REST API 服务
func startRESTServer(h *handler.UserHandler) {
	port := viper.GetString("service.user.rest")
	if port == "" {
		port = "8081"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), observation.GinMiddleware("user"))
	user.RegisterHandlers(router, h)

	log.Printf("REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
