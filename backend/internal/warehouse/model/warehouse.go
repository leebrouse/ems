package model

import "time"

// Item 物品模型
type Item struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"size:128;not null;uniqueIndex:uk_item_name"`
	Unit        string `gorm:"size:32;not null"`
	Description string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Item) TableName() string {
	return "items"
}

// Warehouse 仓库模型
type Warehouse struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:128;not null"`
	Location string `gorm:"size:255"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Warehouse) TableName() string {
	return "warehouses"
}

// Inventory 库存模型
type Inventory struct {
	ID int64 `gorm:"primaryKey;autoIncrement"`

	WarehouseID int64 `gorm:"not null;index:uk_warehouse_item,unique"`
	ItemID      int64 `gorm:"not null;index:uk_warehouse_item,unique"`

	Quantity int `gorm:"not null;default:0"`
	Version  int `gorm:"not null;default:0;version"` // 乐观锁版本字段

	UpdatedAt time.Time

	// 关联
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID"`
	Item      Item      `gorm:"foreignKey:ItemID"`
}

func (Inventory) TableName() string {
	return "inventory"
}

// InventoryLog 库存日志模型
type InventoryLog struct {
	ID int64 `gorm:"primaryKey;autoIncrement"`

	WarehouseID int64 `gorm:"not null"`
	ItemID      int64 `gorm:"not null"`

	ChangeAmount int `gorm:"not null"` // 正数入库，负数出库
	BeforeQty    int `gorm:"not null"`
	AfterQty     int `gorm:"not null"`

	ReferenceType string `gorm:"size:32"`
	ReferenceID   int64
	OperatorID    int64

	CreatedAt time.Time

	// 关联
	Warehouse Warehouse `gorm:"foreignKey:WarehouseID"`
	Item      Item      `gorm:"foreignKey:ItemID"`
}

func (InventoryLog) TableName() string {
	return "inventory_logs"
}

// ItemThreshold 库存阈值模型
type ItemThreshold struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	ItemID    int64 `gorm:"not null;uniqueIndex"`
	Threshold int   `gorm:"not null"`
	UpdatedAt time.Time

	Item Item `gorm:"foreignKey:ItemID"`
}

func (ItemThreshold) TableName() string {
	return "item_thresholds"
}
