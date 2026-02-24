package repository

import (
	"context"

	"github.com/leebrouse/ems/backend/warehouse/model"

	"gorm.io/gorm"
)

// WarehouseRepository 定义仓库数据访问接口
type WarehouseRepository interface {
	// Item operations
	CreateItem(ctx context.Context, item *model.Item) error
	GetItem(ctx context.Context, id int64) (*model.Item, error)
	UpdateItem(ctx context.Context, item *model.Item) error
	DeleteItem(ctx context.Context, id int64) error
	ListItems(ctx context.Context, page, size int, query string) ([]model.Item, int64, error)

	// Warehouse operations
	CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) error
	GetWarehouse(ctx context.Context, id int64) (*model.Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) error
	DeleteWarehouse(ctx context.Context, id int64) error
	ListWarehouses(ctx context.Context) ([]model.Warehouse, error)

	// Inventory operations
	GetInventory(ctx context.Context, warehouseID int64) ([]model.Inventory, error)
	GetInventoryByItem(ctx context.Context, warehouseID, itemID int64) (*model.Inventory, error)
	UpdateInventory(ctx context.Context, tx *gorm.DB, inventory *model.Inventory) error

	// Threshold operations
	SetThreshold(ctx context.Context, itemID int64, threshold int) error
	ListAlerts(ctx context.Context) ([]model.InventoryAlert, error)

	// Logs
	CreateInventoryLog(ctx context.Context, tx *gorm.DB, log *model.InventoryLog) error

	// Transaction helper
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// warehouseRepository 使用 GORM 实现 WarehouseRepository
type warehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建 WarehouseRepository 实例
func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &warehouseRepository{db: db}
}

// CreateItem 保存物资
func (r *warehouseRepository) CreateItem(ctx context.Context, item *model.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetItem 获取物资详情
func (r *warehouseRepository) GetItem(ctx context.Context, id int64) (*model.Item, error) {
	var item model.Item
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateItem 更新物资
func (r *warehouseRepository) UpdateItem(ctx context.Context, item *model.Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// DeleteItem 删除物资
func (r *warehouseRepository) DeleteItem(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Item{}, id).Error
}

// ListItems 分页查询物资列表
func (r *warehouseRepository) ListItems(ctx context.Context, page, size int, query string) ([]model.Item, int64, error) {
	var items []model.Item
	var total int64
	db := r.db.WithContext(ctx).Model(&model.Item{})
	if query != "" {
		db = db.Where("name LIKE ?", "%"+query+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CreateWarehouse 保存仓库
func (r *warehouseRepository) CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) error {
	return r.db.WithContext(ctx).Create(warehouse).Error
}

// GetWarehouse 获取仓库详情
func (r *warehouseRepository) GetWarehouse(ctx context.Context, id int64) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	if err := r.db.WithContext(ctx).First(&warehouse, id).Error; err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// UpdateWarehouse 更新仓库
func (r *warehouseRepository) UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) error {
	return r.db.WithContext(ctx).Save(warehouse).Error
}

// DeleteWarehouse 删除仓库
func (r *warehouseRepository) DeleteWarehouse(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Warehouse{}, id).Error
}

// ListWarehouses 获取仓库列表
func (r *warehouseRepository) ListWarehouses(ctx context.Context) ([]model.Warehouse, error) {
	var warehouses []model.Warehouse
	if err := r.db.WithContext(ctx).Find(&warehouses).Error; err != nil {
		return nil, err
	}
	return warehouses, nil
}

// GetInventory 获取仓库库存
func (r *warehouseRepository) GetInventory(ctx context.Context, warehouseID int64) ([]model.Inventory, error) {
	var inventory []model.Inventory
	if err := r.db.WithContext(ctx).Preload("Item").Where("warehouse_id = ?", warehouseID).Find(&inventory).Error; err != nil {
		return nil, err
	}
	return inventory, nil
}

// GetInventoryByItem 获取指定物资库存
func (r *warehouseRepository) GetInventoryByItem(ctx context.Context, warehouseID, itemID int64) (*model.Inventory, error) {
	var inventory model.Inventory
	err := r.db.WithContext(ctx).Where("warehouse_id = ? AND item_id = ?", warehouseID, itemID).First(&inventory).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.Inventory{WarehouseID: warehouseID, ItemID: itemID, Quantity: 0}, nil
		}
		return nil, err
	}
	return &inventory, nil
}

// UpdateInventory 更新库存并使用乐观锁控制
func (r *warehouseRepository) UpdateInventory(ctx context.Context, tx *gorm.DB, inventory *model.Inventory) error {
	db := tx
	if db == nil {
		db = r.db
	}
	// Using optimistic locking with version field
	result := db.WithContext(ctx).Save(inventory)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrInvalidData // Or a custom optimistic lock error
	}
	return nil
}

// SetThreshold 设置库存预警阈值
func (r *warehouseRepository) SetThreshold(ctx context.Context, itemID int64, threshold int) error {
	var t model.ItemThreshold
	result := r.db.WithContext(ctx).Where("item_id = ?", itemID).First(&t)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return r.db.WithContext(ctx).Create(&model.ItemThreshold{ItemID: itemID, Threshold: threshold}).Error
		}
		return result.Error
	}
	t.Threshold = threshold
	return r.db.WithContext(ctx).Save(&t).Error
}

// ListAlerts 获取库存预警列表
func (r *warehouseRepository) ListAlerts(ctx context.Context) ([]model.InventoryAlert, error) {
	var alerts []model.InventoryAlert
	err := r.db.WithContext(ctx).
		Table("item_thresholds").
		Select("item_thresholds.item_id, items.name, COALESCE(SUM(inventory.quantity), 0) as quantity, item_thresholds.threshold").
		Joins("JOIN items ON items.id = item_thresholds.item_id").
		Joins("LEFT JOIN inventory ON inventory.item_id = item_thresholds.item_id").
		Group("item_thresholds.item_id, items.name, item_thresholds.threshold").
		Having("COALESCE(SUM(inventory.quantity), 0) < item_thresholds.threshold").
		Scan(&alerts).Error
	return alerts, err
}

// CreateInventoryLog 创建库存变更日志
func (r *warehouseRepository) CreateInventoryLog(ctx context.Context, tx *gorm.DB, log *model.InventoryLog) error {
	db := tx
	if db == nil {
		db = r.db
	}
	return db.WithContext(ctx).Create(log).Error
}

// WithTransaction 在事务中执行回调
func (r *warehouseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
