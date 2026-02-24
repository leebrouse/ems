package service

import (
	"context"

	schedulingpb "github.com/leebrouse/ems/backend/common/genproto/scheduling/grpc"
	warehousepb "github.com/leebrouse/ems/backend/common/genproto/warehouse/grpc"
	"github.com/leebrouse/ems/backend/statistics/model"
)

// StatisticsService 定义统计聚合业务能力
type StatisticsService interface {
	GetInventoryStats(ctx context.Context) ([]*model.ItemStock, error)
	GetRequestStats(ctx context.Context, startDate, endDate string) (*model.RequestStats, error)
	GetShipmentStats(ctx context.Context, period string) (*model.ShipmentStats, error)
}

// statisticsService 是 StatisticsService 的默认实现
type statisticsService struct {
	warehouseClient  warehousepb.WarehouseServiceClient
	schedulingClient schedulingpb.SchedulingServiceClient
}

// NewStatisticsService 创建 StatisticsService 实例
func NewStatisticsService(warehouseClient warehousepb.WarehouseServiceClient, schedulingClient schedulingpb.SchedulingServiceClient) StatisticsService {
	return &statisticsService{
		warehouseClient:  warehouseClient,
		schedulingClient: schedulingClient,
	}
}

// GetInventoryStats 获取库存统计数据
func (s *statisticsService) GetInventoryStats(ctx context.Context) ([]*model.ItemStock, error) {
	// Call warehouse service to list items
	itemsRes, err := s.warehouseClient.ListItems(ctx, &warehousepb.ListItemsRequest{Page: 1, Size: 1000})
	if err != nil {
		return nil, err
	}

	var stats []*model.ItemStock
	for _, item := range itemsRes.Items {

		stats = append(stats, &model.ItemStock{
			ItemID:        item.Id,
			Name:          item.Name,
			TotalQuantity: 0, // Need to fill this
		})
	}

	return stats, nil
}

// GetRequestStats 获取需求单统计数据
func (s *statisticsService) GetRequestStats(ctx context.Context, startDate, endDate string) (*model.RequestStats, error) {
	// Call scheduling service
	reqsRes, err := s.schedulingClient.ListRequests(ctx, &schedulingpb.ListRequestsProto{Page: 1, Size: 1000})
	if err != nil {
		return nil, err
	}

	stats := &model.RequestStats{
		StartDate: startDate,
		EndDate:   endDate,
		ByStatus:  make(map[string]int32),
	}

	for _, req := range reqsRes.Requests {
		stats.TotalRequests++
		stats.ByStatus[req.Status]++
	}

	return stats, nil
}

// GetShipmentStats 获取运输统计数据
func (s *statisticsService) GetShipmentStats(ctx context.Context, period string) (*model.ShipmentStats, error) {
	// Call scheduling service
	shipmentsRes, err := s.schedulingClient.ListShipments(ctx, &schedulingpb.ListShipmentsProto{Page: 1, Size: 1000})
	if err != nil {
		return nil, err
	}

	stats := &model.ShipmentStats{
		Period: period,
		Data:   []*model.ShipmentCount{},
	}

	// Group by week/month based on period
	// This is simplified. In a real system, you'd parse dates.
	counts := make(map[string]int32)
	for _, sh := range shipmentsRes.Shipments {
		// Mock grouping based on shipment status or ID for demonstration
		// since CreatedAt is not available in ShipmentResponse proto currently.
		label := "2024-W01"
		if sh.ShipmentId%2 == 0 {
			label = "2024-W02"
		}
		counts[label]++
	}

	for label, count := range counts {
		stats.Data = append(stats.Data, &model.ShipmentCount{
			PeriodLabel: label,
			Count:       count,
		})
	}

	return stats, nil
}
