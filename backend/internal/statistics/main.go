package main

import (
	"log"
	"net"

	"github.com/leebrouse/ems/backend/common/genopenapi/statistics"
	schedulingpb "github.com/leebrouse/ems/backend/common/genproto/scheduling/grpc"
	pb "github.com/leebrouse/ems/backend/common/genproto/statistics/grpc"
	warehousepb "github.com/leebrouse/ems/backend/common/genproto/warehouse/grpc"
	"github.com/leebrouse/ems/backend/statistics/handler"
	"github.com/leebrouse/ems/backend/statistics/rpc"
	"github.com/leebrouse/ems/backend/statistics/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/leebrouse/ems/backend/common/config"
)

// main 负责初始化统计服务依赖并启动 gRPC/REST 服务
func main() {
	// 1. Initialize gRPC clients
	warehouseHost := viper.GetString("service.warehouse.host")
	if warehouseHost == "" {
		warehouseHost = "localhost"
	}
	warehousePort := viper.GetString("service.warehouse.grpc")
	if warehousePort == "" {
		warehousePort = "9002"
	}
	warehouseAddr := warehouseHost + ":" + warehousePort
	wConn, err := grpc.NewClient(warehouseAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to warehouse: %v", err)
	}
	defer wConn.Close()
	wClient := warehousepb.NewWarehouseServiceClient(wConn)

	// 连接调度服务 gRPC 作为统计数据来源
	schedulingHost := viper.GetString("service.scheduling.host")
	if schedulingHost == "" {
		schedulingHost = "localhost"
	}
	schedulingPort := viper.GetString("service.scheduling.grpc")
	if schedulingPort == "" {
		schedulingPort = "9003"
	}
	schedulingAddr := schedulingHost + ":" + schedulingPort
	sConn, err := grpc.NewClient(schedulingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to scheduling: %v", err)
	}
	defer sConn.Close()
	sClient := schedulingpb.NewSchedulingServiceClient(sConn)

	// 2. Initialize Layers
	svc := service.NewStatisticsService(wClient, sClient)
	h := handler.NewStatisticsHandler(svc)
	r := rpc.NewStatisticsRPCServer(svc)

	// 3. Start Servers
	go startGRPCServer(r)

	startRESTServer(h)
}

// startGRPCServer 启动统计 gRPC 服务
func startGRPCServer(r *rpc.StatisticsRPCServer) {
	port := viper.GetString("service.statistics.grpc")
	if port == "" {
		port = "9004"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterStatisticsServiceServer(s, r)
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// startRESTServer 启动统计 REST API 服务
func startRESTServer(h *handler.StatisticsHandler) {
	port := viper.GetString("service.statistics.rest")
	if port == "" {
		port = "8084"
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	statistics.RegisterHandlers(router, h)

	log.Printf("REST server listening at :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
