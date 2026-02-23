package rpc

import (
	"context"

	pb "github.com/leebrouse/ems/backend/common/genproto/statistics/grpc"
	"github.com/leebrouse/ems/backend/statistics/service"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatisticsRPCServer struct {
	pb.UnimplementedStatisticsServiceServer
	svc service.StatisticsService
}

func NewStatisticsRPCServer(svc service.StatisticsService) *StatisticsRPCServer {
	return &StatisticsRPCServer{svc: svc}
}

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

func (s *StatisticsRPCServer) GetShipmentStats(ctx context.Context, in *pb.StatsRequest) (*pb.ShipmentStatsResponse, error) {
	// Note: gRPC StatsRequest uses startDate/endDate, but REST uses period.
	// We'll use startDate as period for now or handle accordingly.
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
