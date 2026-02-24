package rpc

import (
	"context"

	pb "github.com/leebrouse/ems/backend/common/genproto/statistics/grpc"
	"github.com/leebrouse/ems/backend/statistics/service"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatisticsRPCServer 提供统计服务的 gRPC 接口实现
type StatisticsRPCServer struct {
	pb.UnimplementedStatisticsServiceServer
	svc service.StatisticsService
}

// NewStatisticsRPCServer 创建 StatisticsRPCServer 实例
func NewStatisticsRPCServer(svc service.StatisticsService) *StatisticsRPCServer {
	return &StatisticsRPCServer{svc: svc}
}

// GetInventoryStats 获取库存统计
func (s *StatisticsRPCServer) GetInventoryStats(ctx context.Context, in *empty.Empty) (*pb.InventoryStatsResponse, error) {
	items, err := s.svc.GetInventoryStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.ItemStock
	for _, item := range items {
		res = append(res, &pb.ItemStock{
			ItemId:        item.ItemID,
			Name:          item.Name,
			TotalQuantity: item.TotalQuantity,
		})
	}
	return &pb.InventoryStatsResponse{Items: res}, nil
}

// GetRequestStats 获取需求统计
func (s *StatisticsRPCServer) GetRequestStats(ctx context.Context, in *pb.StatsRequest) (*pb.RequestStatsResponse, error) {
	stats, err := s.svc.GetRequestStats(ctx, in.StartDate, in.EndDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RequestStatsResponse{
		StartDate:     stats.StartDate,
		EndDate:       stats.EndDate,
		TotalRequests: stats.TotalRequests,
		ByStatus:      stats.ByStatus,
	}, nil
}

// GetShipmentStats 获取运输统计
func (s *StatisticsRPCServer) GetShipmentStats(ctx context.Context, in *pb.StatsRequest) (*pb.ShipmentStatsResponse, error) {
	stats, err := s.svc.GetShipmentStats(ctx, in.StartDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.ShipmentCount
	for _, d := range stats.Data {
		res = append(res, &pb.ShipmentCount{
			PeriodLabel: d.PeriodLabel,
			Count:       d.Count,
		})
	}
	return &pb.ShipmentStatsResponse{
		Period: stats.Period,
		Data:   res,
	}, nil
}
