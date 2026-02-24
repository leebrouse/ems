package repository

import (
	"context"
	"time"

	"github.com/leebrouse/ems/backend/scheduling/model"
	"gorm.io/gorm"
)

// SchedulingRepository 定义调度数据访问接口
type SchedulingRepository interface {
	// Request
	CreateRequest(ctx context.Context, req *model.Request) error
	GetRequest(ctx context.Context, id int64) (*model.Request, error)
	UpdateRequest(ctx context.Context, req *model.Request) error
	DeleteRequest(ctx context.Context, id int64) error
	ListRequests(ctx context.Context, page, size int, status string) ([]model.Request, int64, error)

	// Shipment
	CreateShipment(ctx context.Context, shipment *model.Shipment) error
	GetShipment(ctx context.Context, id int64) (*model.Shipment, error)
	UpdateShipmentStatus(ctx context.Context, id int64, status model.ShipmentStatus, location string) error
	ListShipments(ctx context.Context, page, size int, status string) ([]model.Shipment, int64, error)
	AddTracking(ctx context.Context, tracking *model.ShipmentTracking) error

	// Transaction helper
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// schedulingRepository 使用 GORM 实现 SchedulingRepository
type schedulingRepository struct {
	db *gorm.DB
}

// NewSchedulingRepository 创建 SchedulingRepository 实例
func NewSchedulingRepository(db *gorm.DB) SchedulingRepository {
	return &schedulingRepository{db: db}
}

// CreateRequest 保存需求单
func (r *schedulingRepository) CreateRequest(ctx context.Context, req *model.Request) error {
	return r.db.WithContext(ctx).Create(req).Error
}

// GetRequest 根据 ID 获取需求单并预加载明细
func (r *schedulingRepository) GetRequest(ctx context.Context, id int64) (*model.Request, error) {
	var req model.Request
	if err := r.db.WithContext(ctx).Preload("Items").First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// UpdateRequest 更新需求单
func (r *schedulingRepository) UpdateRequest(ctx context.Context, req *model.Request) error {
	return r.db.WithContext(ctx).Save(req).Error
}

// DeleteRequest 删除需求单
func (r *schedulingRepository) DeleteRequest(ctx context.Context, id int64) error {
	// Only allow deleting if PENDING? (Logic better in service, but repo facilitates)
	return r.db.WithContext(ctx).Delete(&model.Request{}, id).Error
}

// ListRequests 分页查询需求单
func (r *schedulingRepository) ListRequests(ctx context.Context, page, size int, status string) ([]model.Request, int64, error) {
	var reqs []model.Request
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Request{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Items").Offset((page - 1) * size).Limit(size).Find(&reqs).Error; err != nil {
		return nil, 0, err
	}
	return reqs, total, nil
}

// CreateShipment 保存运输任务
func (r *schedulingRepository) CreateShipment(ctx context.Context, shipment *model.Shipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

// GetShipment 获取运输任务并预加载明细与轨迹
func (r *schedulingRepository) GetShipment(ctx context.Context, id int64) (*model.Shipment, error) {
	var shipment model.Shipment
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Tracking").First(&shipment, id).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

// UpdateShipmentStatus 更新运输状态并追加轨迹
func (r *schedulingRepository) UpdateShipmentStatus(ctx context.Context, id int64, status model.ShipmentStatus, location string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Shipment{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		// Add tracking record
		tracking := &model.ShipmentTracking{
			ShipmentID: id,
			Status:     status,
			Location:   location,
			RecordedAt: time.Now(),
		}
		return tx.Create(tracking).Error
	})
}

// ListShipments 分页查询运输任务
func (r *schedulingRepository) ListShipments(ctx context.Context, page, size int, status string) ([]model.Shipment, int64, error) {
	var shipments []model.Shipment
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Shipment{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Items").Offset((page - 1) * size).Limit(size).Find(&shipments).Error; err != nil {
		return nil, 0, err
	}
	return shipments, total, nil
}

// AddTracking 添加运输轨迹
func (r *schedulingRepository) AddTracking(ctx context.Context, tracking *model.ShipmentTracking) error {
	return r.db.WithContext(ctx).Create(tracking).Error
}

// WithTransaction 在事务中执行回调
func (r *schedulingRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
