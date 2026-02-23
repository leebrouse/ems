package repository

import (
	"context"
	"time"

	"github.com/leebrouse/ems/backend/scheduling/model"
	"gorm.io/gorm"
)

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

type schedulingRepository struct {
	db *gorm.DB
}

func NewSchedulingRepository(db *gorm.DB) SchedulingRepository {
	return &schedulingRepository{db: db}
}

func (r *schedulingRepository) CreateRequest(ctx context.Context, req *model.Request) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *schedulingRepository) GetRequest(ctx context.Context, id int64) (*model.Request, error) {
	var req model.Request
	if err := r.db.WithContext(ctx).Preload("Items").First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *schedulingRepository) UpdateRequest(ctx context.Context, req *model.Request) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *schedulingRepository) DeleteRequest(ctx context.Context, id int64) error {
	// Only allow deleting if PENDING? (Logic better in service, but repo facilitates)
	return r.db.WithContext(ctx).Delete(&model.Request{}, id).Error
}

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

func (r *schedulingRepository) CreateShipment(ctx context.Context, shipment *model.Shipment) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

func (r *schedulingRepository) GetShipment(ctx context.Context, id int64) (*model.Shipment, error) {
	var shipment model.Shipment
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Tracking").First(&shipment, id).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

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

func (r *schedulingRepository) AddTracking(ctx context.Context, tracking *model.ShipmentTracking) error {
	return r.db.WithContext(ctx).Create(tracking).Error
}

func (r *schedulingRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
