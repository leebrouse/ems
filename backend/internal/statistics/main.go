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

func main() {
	// 1. Initialize gRPC clients
	warehouseAddr := viper.GetString("service.warehouse.grpc")
	if warehouseAddr == "" {
		warehouseAddr = "localhost:9002"
	}
	wConn, err := grpc.NewClient(warehouseAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to warehouse: %v", err)
	}
	defer wConn.Close()
	wClient := warehousepb.NewWarehouseServiceClient(wConn)

	// 
	schedulingAddr := viper.GetString("service.scheduling.grpc")
	if schedulingAddr == "" {
		schedulingAddr = "localhost:9003"
	}
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
