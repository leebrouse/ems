package main

import (
	"context"
	"log"
	"time"

	"github.com/leebrouse/ems/backend/auth/handler"
	"github.com/leebrouse/ems/backend/auth/model"
	"github.com/leebrouse/ems/backend/auth/repository"
	"github.com/leebrouse/ems/backend/auth/service"
	authjwt "github.com/leebrouse/ems/backend/common/auth"
	"github.com/leebrouse/ems/backend/common/database"
	"github.com/leebrouse/ems/backend/common/genopenapi/auth"
	userpb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"
	"github.com/leebrouse/ems/backend/common/observation"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化认证服务依赖并启动 REST 服务
func main() {
	shutdown, err := observation.InitFromViper(context.Background(), "auth")
	if err != nil {
		log.Fatalf("failed to init observation: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	// 1. Initialize Database
	db := database.Connect("service.auth.postgres",
		&model.RefreshToken{},
	)

	// 2. Initialize gRPC clients
	userHost := viper.GetString("service.user.host")
	if userHost == "" {
		userHost = "localhost"
	}
	userPort := viper.GetString("service.user.grpc")
	if userPort == "" {
		userPort = "9001"
	}
	userAddr := userHost + ":" + userPort
	uConn, err := grpc.NewClient(userAddr, observation.GRPCDialOptions()...)
	if err != nil {
		log.Fatalf("did not connect to user service: %v", err)
	}
	defer uConn.Close()
	uClient := userpb.NewUserServiceClient(uConn)

	// 3. Initialize JWT Handler
	secret := viper.GetString("security.encryption_key")
	if secret == "" {
		secret = "emergency-system-aes-key-32chars"
	}
	jwtHandler := authjwt.NewJWTHandler(secret, "rescue-system", 24*time.Hour)

	// 4. Initialize Layers
	repo := repository.NewAuthRepository(db)
	svc := service.NewAuthService(repo, uClient, jwtHandler)
	h := handler.NewAuthHandler(svc)

	// 5. Start Server
	startRESTServer(h)
}

// startRESTServer 启动认证 REST API 服务
func startRESTServer(h *handler.AuthHandler) {
	port := viper.GetString("service.auth.rest")
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), observation.GinMiddleware("auth"))
	auth.RegisterHandlers(router, h)

	log.Printf("Auth REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
