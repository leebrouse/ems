package main

import (
	"log"
	"net"

	"github.com/leebrouse/ems/backend/common/database"
	"github.com/leebrouse/ems/backend/common/genopenapi/warehouse"
	pb "github.com/leebrouse/ems/backend/common/genproto/warehouse/grpc"
	"github.com/leebrouse/ems/backend/warehouse/handler"
	"github.com/leebrouse/ems/backend/warehouse/model"
	"github.com/leebrouse/ems/backend/warehouse/repository"
	"github.com/leebrouse/ems/backend/warehouse/rpc"
	"github.com/leebrouse/ems/backend/warehouse/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化仓库服务依赖并启动 gRPC/REST 服务
func main() {
	// 1. Initialize Database
	db := database.Connect("service.warehouse.postgres",
		&model.Item{},
		&model.Warehouse{},
		&model.Inventory{},
		&model.InventoryLog{},
		&model.ItemThreshold{},
	)

	// 2. Initialize Layers
	repo := repository.NewWarehouseRepository(db)
	svc := service.NewWarehouseService(repo)
	h := handler.NewWarehouseHandler(svc)
	r := rpc.NewWarehouseRPCServer(svc)

	// 3. Start Servers
	go startGRPCServer(r)

	startRESTServer(h)
}

// startGRPCServer 启动仓库 gRPC 服务
func startGRPCServer(r *rpc.WarehouseRPCServer) {
	// 从配置读取 gRPC 端口，未配置时使用默认值
	port := viper.GetString("service.warehouse.grpc")
	if port == "" {
		port = "9002"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWarehouseServiceServer(s, r)
	// 注册 gRPC 服务实现
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startRESTServer 启动仓库 REST API 服务
func startRESTServer(h *handler.WarehouseHandler) {
	port := viper.GetString("service.warehouse.rest")
	if port == "" {
		port = "8002"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	warehouse.RegisterHandlers(router, h)

	log.Printf("REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
