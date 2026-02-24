package service

import (
	"context"
	"errors"
	"time"

	"github.com/leebrouse/ems/backend/scheduling/model"
	"github.com/leebrouse/ems/backend/scheduling/repository"
	"gorm.io/gorm"
)

var (
	// ErrRequestNotPending 表示需求单状态不允许删除
	ErrRequestNotPending = errors.New("request is not in pending status")
)

// SchedulingService 定义调度领域业务能力
type SchedulingService interface {
	// Request
	CreateRequest(ctx context.Context, title, location string, items []model.RequestItem, createdBy int64) (*model.Request, error)
	GetRequest(ctx context.Context, id int64) (*model.Request, error)
	UpdateRequest(ctx context.Context, id int64, status model.RequestStatus, assignedTo *int64) (*model.Request, error)
	DeleteRequest(ctx context.Context, id int64) error
	ListRequests(ctx context.Context, page, size int, status string) ([]model.Request, int64, error)

	// Shipment
	CreateShipment(ctx context.Context, requestId int64, fromWarehouseId int64, toLocation string, items []model.ShipmentItem) (*model.Shipment, error)
	GetShipment(ctx context.Context, id int64) (*model.Shipment, error)
	UpdateShipmentStatus(ctx context.Context, id int64, status model.ShipmentStatus, location string) (*model.Shipment, error)
	ListShipments(ctx context.Context, page, size int, status string) ([]model.Shipment, int64, error)
}

// schedulingService 是 SchedulingService 的默认实现
type schedulingService struct {
	repo repository.SchedulingRepository
}

// NewSchedulingService 创建 SchedulingService 实例
func NewSchedulingService(repo repository.SchedulingRepository) SchedulingService {
	return &schedulingService{repo: repo}
}

// CreateRequest 创建需求单
func (s *schedulingService) CreateRequest(ctx context.Context, title, location string, items []model.RequestItem, createdBy int64) (*model.Request, error) {
	req := &model.Request{
		Title:     title,
		Location:  location,
		Status:    model.RequestStatusPending,
		Items:     items,
		CreatedBy: createdBy,
	}
	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

// GetRequest 获取需求单详情
func (s *schedulingService) GetRequest(ctx context.Context, id int64) (*model.Request, error) {
	return s.repo.GetRequest(ctx, id)
}

// UpdateRequest 更新需求单状态或负责人
func (s *schedulingService) UpdateRequest(ctx context.Context, id int64, status model.RequestStatus, assignedTo *int64) (*model.Request, error) {
	req, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if status != "" {
		req.Status = status
	}
	if assignedTo != nil {
		req.AssignedTo = assignedTo
	}
	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

// DeleteRequest 删除需求单
func (s *schedulingService) DeleteRequest(ctx context.Context, id int64) error {
	req, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != model.RequestStatusPending {
		return ErrRequestNotPending
	}
	return s.repo.DeleteRequest(ctx, id)
}

// ListRequests 分页查询需求单
func (s *schedulingService) ListRequests(ctx context.Context, page, size int, status string) ([]model.Request, int64, error) {
	return s.repo.ListRequests(ctx, page, size, status)
}

// CreateShipment 创建运输任务并写入轨迹
func (s *schedulingService) CreateShipment(ctx context.Context, requestId int64, fromWarehouseId int64, toLocation string, items []model.ShipmentItem) (*model.Shipment, error) {
	shipment := &model.Shipment{
		RequestID:       requestId,
		FromWarehouseID: fromWarehouseId,
		ToLocation:      toLocation,
		Status:          model.ShipmentStatusNew,
		Items:           items,
	}

	err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 1. Create shipment
		if err := s.repo.CreateShipment(ctx, shipment); err != nil {
			return err
		}

		// 2. Add initial tracking info
		tracking := &model.ShipmentTracking{
			ShipmentID: shipment.ID,
			Status:     model.ShipmentStatusNew,
			Location:   "Warehouse", // Simplified
			RecordedAt: time.Now(),
		}
		if err := s.repo.AddTracking(ctx, tracking); err != nil {
			return err
		}

		// 3. Update request status
		req, err := s.repo.GetRequest(ctx, requestId)
		if err != nil {
			return err
		}
		req.Status = model.RequestStatusAssigned // Or something similar
		if err := s.repo.UpdateRequest(ctx, req); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return shipment, nil
}

// GetShipment 获取运输任务详情
func (s *schedulingService) GetShipment(ctx context.Context, id int64) (*model.Shipment, error) {
	return s.repo.GetShipment(ctx, id)
}

// UpdateShipmentStatus 更新运输状态并返回最新数据
func (s *schedulingService) UpdateShipmentStatus(ctx context.Context, id int64, status model.ShipmentStatus, location string) (*model.Shipment, error) {
	if err := s.repo.UpdateShipmentStatus(ctx, id, status, location); err != nil {
		return nil, err
	}
	return s.repo.GetShipment(ctx, id)
}

// ListShipments 分页查询运输任务
func (s *schedulingService) ListShipments(ctx context.Context, page, size int, status string) ([]model.Shipment, int64, error) {
	return s.repo.ListShipments(ctx, page, size, status)
}
