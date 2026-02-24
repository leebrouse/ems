package main

import (
	"log"
	"net"
	"time"

	"github.com/leebrouse/ems/backend/auth/handler"
	"github.com/leebrouse/ems/backend/auth/model"
	"github.com/leebrouse/ems/backend/auth/repository"
	"github.com/leebrouse/ems/backend/auth/service"
	authjwt "github.com/leebrouse/ems/backend/common/auth"
	"github.com/leebrouse/ems/backend/common/database"
	"github.com/leebrouse/ems/backend/common/genopenapi/auth"
	userpb "github.com/leebrouse/ems/backend/common/genproto/user/grpc"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化认证服务依赖并启动 REST 服务
func main() {
	// 1. Initialize Database
	db := database.Connect("service.auth.postgres",
		&model.RefreshToken{},
	)

	// 2. Initialize gRPC clients
	userAddr := viper.GetString("service.user.grpc")
	if userAddr == "" {
		userAddr = "localhost:9001"
	} else if userAddr != "" && userAddr[0] != ':' {
		// Handle case where it's just a port or an address
		// viper.GetString might return just "9001" if it's treated as string
	}

	// In our config it's "9001", but grpc.Dial needs "host:port"
	if _, err := net.LookupHost("127.0.0.1"); err == nil {
		if !contains(userAddr, ":") {
			userAddr = "localhost:" + userAddr
		}
	}

	uConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// contains 判断字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// startRESTServer 启动认证 REST API 服务
func startRESTServer(h *handler.AuthHandler) {
	port := viper.GetString("service.auth.rest")
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	auth.RegisterHandlers(router, h)

	log.Printf("Auth REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
