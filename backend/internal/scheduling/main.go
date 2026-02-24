package main

import (
	"log"
	"net"

	"github.com/leebrouse/ems/backend/common/database"
	"github.com/leebrouse/ems/backend/common/genopenapi/scheduling"
	pb "github.com/leebrouse/ems/backend/common/genproto/scheduling/grpc"
	"github.com/leebrouse/ems/backend/scheduling/handler"
	"github.com/leebrouse/ems/backend/scheduling/model"
	"github.com/leebrouse/ems/backend/scheduling/repository"
	"github.com/leebrouse/ems/backend/scheduling/rpc"
	"github.com/leebrouse/ems/backend/scheduling/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化调度服务依赖并启动 gRPC/REST 服务
func main() {
	// 1. Initialize Database
	db := database.Connect("service.scheduling.postgres",
		&model.Request{},
		&model.RequestItem{},
		&model.Shipment{},
		&model.ShipmentItem{},
		&model.ShipmentTracking{},
	)

	// 2. Initialize Layers
	repo := repository.NewSchedulingRepository(db)
	svc := service.NewSchedulingService(repo)
	h := handler.NewSchedulingHandler(svc)
	r := rpc.NewSchedulingRPCServer(svc)

	// 3. Start Servers
	go startGRPCServer(r)

	startRESTServer(h)
}

// startGRPCServer 启动调度 gRPC 服务
func startGRPCServer(r *rpc.SchedulingRPCServer) {
	port := viper.GetString("service.scheduling.grpc")
	if port == "" {
		port = "9003"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSchedulingServiceServer(s, r)
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startRESTServer 启动调度 REST API 服务
func startRESTServer(h *handler.SchedulingHandler) {
	port := viper.GetString("service.scheduling.rest")
	if port == "" {
		port = "8083"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	scheduling.RegisterHandlers(router, h)

	log.Printf("REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
