package service

import (
	"context"
	"errors"

	"github.com/leebrouse/ems/backend/warehouse/model"
	"github.com/leebrouse/ems/backend/warehouse/repository"

	"gorm.io/gorm"
)

var (
	// ErrInsufficientStock 表示库存不足
	ErrInsufficientStock = errors.New("insufficient stock")
	// ErrOptimisticLock 表示乐观锁更新失败
	ErrOptimisticLock    = errors.New("optimistic lock failed")
)

// WarehouseService 定义仓库领域业务能力
type WarehouseService interface {
	// Item
	CreateItem(ctx context.Context, item *model.Item) (*model.Item, error)
	GetItem(ctx context.Context, id int64) (*model.Item, error)
	UpdateItem(ctx context.Context, id int64, name, unit, description string) (*model.Item, error)
	DeleteItem(ctx context.Context, id int64) error
	ListItems(ctx context.Context, page, size int, query string) ([]model.Item, int64, error)

	// Warehouse
	CreateWarehouse(ctx context.Context, name, location string) (*model.Warehouse, error)
	GetWarehouse(ctx context.Context, id int64) (*model.Warehouse, error)
	UpdateWarehouse(ctx context.Context, id int64, name, location string) (*model.Warehouse, error)
	DeleteWarehouse(ctx context.Context, id int64) error
	ListWarehouses(ctx context.Context) ([]model.Warehouse, error)

	// Inventory
	GetInventory(ctx context.Context, warehouseID int64) ([]model.Inventory, error)
	AdjustInventory(ctx context.Context, warehouseID, itemID int64, amount int, referenceType string, referenceID int64) (*model.Inventory, error)

	// Alerts
	SetThreshold(ctx context.Context, itemID int64, threshold int) error
	ListAlerts(ctx context.Context) ([]model.ItemThreshold, error)
}

// warehouseService 是 WarehouseService 的默认实现
type warehouseService struct {
	repo repository.WarehouseRepository
}

// NewWarehouseService 创建 WarehouseService 实例
func NewWarehouseService(repo repository.WarehouseRepository) WarehouseService {
	return &warehouseService{repo: repo}
}

// CreateItem 创建物资
func (s *warehouseService) CreateItem(ctx context.Context, item *model.Item) (*model.Item, error) {
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// GetItem 获取物资详情
func (s *warehouseService) GetItem(ctx context.Context, id int64) (*model.Item, error) {
	return s.repo.GetItem(ctx, id)
}

// UpdateItem 更新物资信息
func (s *warehouseService) UpdateItem(ctx context.Context, id int64, name, unit, description string) (*model.Item, error) {
	item, err := s.repo.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		item.Name = name
	}
	if unit != "" {
		item.Unit = unit
	}
	if description != "" {
		item.Description = description
	}
	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteItem 删除物资
func (s *warehouseService) DeleteItem(ctx context.Context, id int64) error {
	return s.repo.DeleteItem(ctx, id)
}

// ListItems 分页查询物资列表
func (s *warehouseService) ListItems(ctx context.Context, page, size int, query string) ([]model.Item, int64, error) {
	return s.repo.ListItems(ctx, page, size, query)
}

// CreateWarehouse 创建仓库
func (s *warehouseService) CreateWarehouse(ctx context.Context, name, location string) (*model.Warehouse, error) {
	w := &model.Warehouse{Name: name, Location: location}
	if err := s.repo.CreateWarehouse(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

// GetWarehouse 获取仓库详情
func (s *warehouseService) GetWarehouse(ctx context.Context, id int64) (*model.Warehouse, error) {
	return s.repo.GetWarehouse(ctx, id)
}

// UpdateWarehouse 更新仓库信息
func (s *warehouseService) UpdateWarehouse(ctx context.Context, id int64, name, location string) (*model.Warehouse, error) {
	w, err := s.repo.GetWarehouse(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		w.Name = name
	}
	if location != "" {
		w.Location = location
	}
	if err := s.repo.UpdateWarehouse(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

// DeleteWarehouse 删除仓库
func (s *warehouseService) DeleteWarehouse(ctx context.Context, id int64) error {
	return s.repo.DeleteWarehouse(ctx, id)
}

// ListWarehouses 获取仓库列表
func (s *warehouseService) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	return s.repo.ListWarehouses(ctx)
}

// GetInventory 获取仓库库存
func (s *warehouseService) GetInventory(ctx context.Context, warehouseID int64) ([]model.Inventory, error) {
	return s.repo.GetInventory(ctx, warehouseID)
}

// AdjustInventory 调整库存并写入日志
func (s *warehouseService) AdjustInventory(ctx context.Context, warehouseID, itemID int64, amount int, referenceType string, referenceID int64) (*model.Inventory, error) {
	var updatedInventory *model.Inventory
	err := s.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 1. Get current inventory
		inv, err := s.repo.GetInventoryByItem(ctx, warehouseID, itemID)
		if err != nil {
			return err
		}

		// 2. Check if subtraction is possible
		beforeQty := inv.Quantity
		afterQty := beforeQty + amount
		if afterQty < 0 {
			return ErrInsufficientStock
		}

		// 3. Update inventory
		inv.Quantity = afterQty
		inv.Version++
		if err := s.repo.UpdateInventory(ctx, tx, inv); err != nil {
			return err
		}

		// 4. Create log
		log := &model.InventoryLog{
			WarehouseID:   warehouseID,
			ItemID:        itemID,
			ChangeAmount:  amount,
			BeforeQty:     beforeQty,
			AfterQty:      afterQty,
			ReferenceType: referenceType,
			ReferenceID:   referenceID,
		}
		if err := s.repo.CreateInventoryLog(ctx, tx, log); err != nil {
			return err
		}

		updatedInventory = inv
		return nil
	})

	if err != nil {
		return nil, err
	}
	return updatedInventory, nil
}

// SetThreshold 设置库存预警阈值
func (s *warehouseService) SetThreshold(ctx context.Context, itemID int64, threshold int) error {
	return s.repo.SetThreshold(ctx, itemID, threshold)
}

// ListAlerts 查询库存预警
func (s *warehouseService) ListAlerts(ctx context.Context) ([]model.ItemThreshold, error) {
	return s.repo.ListAlerts(ctx)
}
