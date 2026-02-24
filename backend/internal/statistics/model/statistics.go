package model

// ItemStock 表示库存统计项
type ItemStock struct {
	ItemID        int32  `json:"itemId"`
	Name          string `json:"name"`
	TotalQuantity int32  `json:"totalQuantity"`
}

// RequestStats 表示需求统计
type RequestStats struct {
	StartDate     string           `json:"startDate"`
	EndDate       string           `json:"endDate"`
	TotalRequests int32            `json:"totalRequests"`
	ByStatus      map[string]int32 `json:"byStatus"`
}

// ShipmentCount 表示运输统计分组
type ShipmentCount struct {
	PeriodLabel string `json:"periodLabel"`
	Count       int32  `json:"count"`
}

// ShipmentStats 表示运输统计
type ShipmentStats struct {
	Period string           `json:"period"`
	Data   []*ShipmentCount `json:"data"`
}
