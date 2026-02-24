package model

import "time"

// RequestStatus 表示需求单状态
type RequestStatus string

const (
	// RequestStatusPending 待分配
	RequestStatusPending   RequestStatus = "PENDING"
	// RequestStatusAssigned 已分配
	RequestStatusAssigned  RequestStatus = "ASSIGNED"
	// RequestStatusCompleted 已完成
	RequestStatusCompleted RequestStatus = "COMPLETED"
	// RequestStatusCancelled 已取消
	RequestStatusCancelled RequestStatus = "CANCELLED"
)

// ShipmentStatus 表示运输任务状态
type ShipmentStatus string

const (
	// ShipmentStatusNew 新建
	ShipmentStatusNew       ShipmentStatus = "NEW"
	// ShipmentStatusInTransit 运输中
	ShipmentStatusInTransit ShipmentStatus = "IN_TRANSIT"
	// ShipmentStatusDelivered 已送达
	ShipmentStatusDelivered ShipmentStatus = "DELIVERED"
	// ShipmentStatusCancelled 已取消
	ShipmentStatusCancelled ShipmentStatus = "CANCELLED"
)

// Request 需求单模型
type Request struct {
	ID         int64         `gorm:"primaryKey;autoIncrement"`
	Title      string        `gorm:"size:255;not null"`
	Location   string        `gorm:"size:255;not null"`
	Status     RequestStatus `gorm:"size:32;not null;default:'PENDING'"`
	AssignedTo *int64        `gorm:"index"`
	CreatedBy  int64         `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Items []RequestItem `gorm:"foreignKey:RequestID"`
}

func (Request) TableName() string {
	return "requests"
}

// RequestItem 需求明细
type RequestItem struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	RequestID int64 `gorm:"not null;index"`
	ItemID    int64 `gorm:"not null"`
	Quantity  int   `gorm:"not null"`
}

func (RequestItem) TableName() string {
	return "request_items"
}

// Shipment 运输任务
type Shipment struct {
	ID              int64          `gorm:"primaryKey;autoIncrement"`
	RequestID       int64          `gorm:"not null;index"`
	FromWarehouseID int64          `gorm:"not null"`
	ToLocation      string         `gorm:"size:255;not null"`
	Status          ShipmentStatus `gorm:"size:32;not null;default:'NEW'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Items    []ShipmentItem     `gorm:"foreignKey:ShipmentID"`
	Tracking []ShipmentTracking `gorm:"foreignKey:ShipmentID"`
}

func (Shipment) TableName() string {
	return "shipments"
}

// ShipmentItem 运输明细
type ShipmentItem struct {
	ID         int64 `gorm:"primaryKey;autoIncrement"`
	ShipmentID int64 `gorm:"not null;index"`
	ItemID     int64 `gorm:"not null"`
	Quantity   int   `gorm:"not null"`
}

func (ShipmentItem) TableName() string {
	return "shipment_items"
}

// ShipmentTracking 运输轨迹
type ShipmentTracking struct {
	ID         int64          `gorm:"primaryKey;autoIncrement"`
	ShipmentID int64          `gorm:"not null;index"`
	Status     ShipmentStatus `gorm:"size:32;not null"`
	Location   string         `gorm:"size:255"`
	RecordedAt time.Time      `gorm:"not null"`
}

func (ShipmentTracking) TableName() string {
	return "shipment_tracking"
}
