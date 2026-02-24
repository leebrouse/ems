package handler

import (
	"log"
	"net/http"

	"github.com/leebrouse/ems/backend/common/genopenapi/scheduling"
	"github.com/leebrouse/ems/backend/scheduling/model"
	"github.com/leebrouse/ems/backend/scheduling/service"

	"github.com/gin-gonic/gin"
)

// SchedulingHandler 处理调度相关的 HTTP 请求
type SchedulingHandler struct {
	svc service.SchedulingService
}

// NewSchedulingHandler 创建 SchedulingHandler 实例
func NewSchedulingHandler(svc service.SchedulingService) *SchedulingHandler {
	return &SchedulingHandler{svc: svc}
}

// Ensure SchedulingHandler implements scheduling.ServerInterface
var _ scheduling.ServerInterface = (*SchedulingHandler)(nil)

// ListRequests 分页查询需求单
func (h *SchedulingHandler) ListRequests(c *gin.Context, params scheduling.ListRequestsParams) {
	page := int(1)
	if params.Page != nil {
		page = int(*params.Page)
	}
	size := int(20)
	if params.Size != nil {
		size = int(*params.Size)
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}

	reqs, total, err := h.svc.ListRequests(c.Request.Context(), page, size, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"requests": reqs,
		"total":    total,
		"page":     page,
		"size":     size,
	})
}

// CreateRequest 创建需求单
func (h *SchedulingHandler) CreateRequest(c *gin.Context) {
	var body scheduling.CreateRequestJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []model.RequestItem
	for _, item := range body.Items {
		id := int64(0)
		if item.ItemId != nil {
			id = int64(*item.ItemId)
		}
		qty := 0
		if item.Quantity != nil {
			qty = int(*item.Quantity)
		}
		items = append(items, model.RequestItem{
			ItemID:   id,
			Quantity: qty,
		})
	}

	// For now, we don't have user authentication context here, using 0 as createdBy.
	// In reality, this should come from JWT.
	req, err := h.svc.CreateRequest(c.Request.Context(), body.Title, body.Location, items, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

// DeleteRequest 删除需求单
func (h *SchedulingHandler) DeleteRequest(c *gin.Context, id int32) {
	err := h.svc.DeleteRequest(c.Request.Context(), int64(id))
	if err != nil {
		if err == service.ErrRequestNotPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetRequest 获取需求单详情
func (h *SchedulingHandler) GetRequest(c *gin.Context, id int32) {
	req, err := h.svc.GetRequest(c.Request.Context(), int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// UpdateRequest 更新需求单
func (h *SchedulingHandler) UpdateRequest(c *gin.Context, id int32) {
	var body scheduling.UpdateRequestJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var status model.RequestStatus
	if body.Status != nil {
		status = model.RequestStatus(*body.Status)
	}

	var assignedTo *int64
	if body.AssignedTo != nil {
		val := int64(*body.AssignedTo)
		assignedTo = &val
	}

	req, err := h.svc.UpdateRequest(c.Request.Context(), int64(id), status, assignedTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// ListShipments 分页查询运输任务
func (h *SchedulingHandler) ListShipments(c *gin.Context, params scheduling.ListShipmentsParams) {
	page := int(1)
	if params.Page != nil {
		page = int(*params.Page)
	}
	size := int(20)
	if params.Size != nil {
		size = int(*params.Size)
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}

	shipments, total, err := h.svc.ListShipments(c.Request.Context(), page, size, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shipments": shipments,
		"total":     total,
		"page":      page,
		"size":      size,
	})
}

// CreateShipment 创建运输任务
func (h *SchedulingHandler) CreateShipment(c *gin.Context) {
	var body scheduling.CreateShipmentJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []model.ShipmentItem
	for _, item := range body.Items {
		id := int64(0)
		if item.ItemId != nil {
			id = int64(*item.ItemId)
		}
		qty := 0
		if item.Quantity != nil {
			qty = int(*item.Quantity)
		}
		items = append(items, model.ShipmentItem{
			ItemID:   id,
			Quantity: qty,
		})
	}

	shipment, err := h.svc.CreateShipment(c.Request.Context(), int64(body.RequestId), int64(body.FromWarehouseId), body.ToLocation, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, shipment)
}

// GetShipment 获取运输任务详情
func (h *SchedulingHandler) GetShipment(c *gin.Context, shipmentId int32) {
	log.Printf("GetShipment: %v", shipmentId)
	shipment, err := h.svc.GetShipment(c.Request.Context(), int64(shipmentId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipment)
}

// UpdateShipmentStatus 更新运输任务状态
func (h *SchedulingHandler) UpdateShipmentStatus(c *gin.Context, shipmentId int32) {
	var body scheduling.UpdateShipmentStatusJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loc := ""
	if body.Location != nil {
		loc = *body.Location
	}

	shipment, err := h.svc.UpdateShipmentStatus(c.Request.Context(), int64(shipmentId), model.ShipmentStatus(body.Status), loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipment)
}
