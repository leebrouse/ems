package model

type ItemStock struct {
	ItemID        int32  `json:"itemId"`
	Name          string `json:"name"`
	TotalQuantity int32  `json:"totalQuantity"`
}

type RequestStats struct {
	StartDate     string           `json:"startDate"`
	EndDate       string           `json:"endDate"`
	TotalRequests int32            `json:"totalRequests"`
	ByStatus      map[string]int32 `json:"byStatus"`
}

type ShipmentCount struct {
	PeriodLabel string `json:"periodLabel"`
	Count       int32  `json:"count"`
}

type ShipmentStats struct {
	Period string           `json:"period"`
	Data   []*ShipmentCount `json:"data"`
}
