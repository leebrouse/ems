package handler

import (
	"net/http"

	"github.com/leebrouse/ems/backend/common/genopenapi/warehouse"
	"github.com/leebrouse/ems/backend/warehouse/model"
	"github.com/leebrouse/ems/backend/warehouse/service"

	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct {
	svc service.WarehouseService
}

// NewWarehouseHandler creates a new WarehouseHandler
func NewWarehouseHandler(svc service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

func (h *WarehouseHandler) ListAlerts(c *gin.Context) {
	alerts, err := h.svc.ListAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *WarehouseHandler) SetThreshold(c *gin.Context) {
	var body warehouse.SetThresholdJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.svc.SetThreshold(c.Request.Context(), int64(body.ItemId), int(body.Threshold))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// 
func (h *WarehouseHandler) ListItems(c *gin.Context, params warehouse.ListItemsParams) {
	page := int(1)
	if params.Page != nil {
		page = int(*params.Page)
	}
	size := int(20)
	if params.Size != nil {
		size = int(*params.Size)
	}
	query := ""
	if params.Query != nil {
		query = *params.Query
	}

	items, total, err := h.svc.ListItems(c.Request.Context(), page, size, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// 
func (h *WarehouseHandler) CreateItem(c *gin.Context) {
	var body warehouse.CreateItemJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	desc := ""
	if body.Description != nil {
		desc = *body.Description
	}
	item := &model.Item{
		Name:        body.Name,
		Unit:        body.Unit,
		Description: desc,
	}
	created, err := h.svc.CreateItem(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// 
func (h *WarehouseHandler) DeleteItem(c *gin.Context, itemId int32) {
	err := h.svc.DeleteItem(c.Request.Context(), int64(itemId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// 
func (h *WarehouseHandler) GetItem(c *gin.Context, itemId int32) {
	item, err := h.svc.GetItem(c.Request.Context(), int64(itemId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

//
func (h *WarehouseHandler) UpdateItem(c *gin.Context, itemId int32) {
	var body warehouse.UpdateItemJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := ""
	if body.Name != nil {
		name = *body.Name
	}
	unit := ""
	if body.Unit != nil {
		unit = *body.Unit
	}
	desc := ""
	if body.Description != nil {
		desc = *body.Description
	}
	updated, err := h.svc.UpdateItem(c.Request.Context(), int64(itemId), name, unit, desc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

//
func (h *WarehouseHandler) ListWarehouses(c *gin.Context) {
	ws, err := h.svc.ListWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ws)
}

//
func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	var body warehouse.CreateWarehouseJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	loc := ""
	if body.Location != nil {
		loc = *body.Location
	}
	w, err := h.svc.CreateWarehouse(c.Request.Context(), body.Name, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, w)
}

func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context, id int32) {
	err := h.svc.DeleteWarehouse(c.Request.Context(), int64(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WarehouseHandler) GetWarehouse(c *gin.Context, id int32) {
	w, err := h.svc.GetWarehouse(c.Request.Context(), int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context, id int32) {
	var body warehouse.UpdateWarehouseJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	name := ""
	if body.Name != nil {
		name = *body.Name
	}
	loc := ""
	if body.Location != nil {
		loc = *body.Location
	}
	updated, err := h.svc.UpdateWarehouse(c.Request.Context(), int64(id), name, loc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *WarehouseHandler) GetInventory(c *gin.Context, id int32) {
	inv, err := h.svc.GetInventory(c.Request.Context(), int64(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *WarehouseHandler) AddInventory(c *gin.Context, id int32) {
	var body warehouse.AddInventoryJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inv, err := h.svc.AdjustInventory(c.Request.Context(), int64(id), int64(body.ItemId), int(body.Amount), "MANUAL", 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *WarehouseHandler) RemoveInventory(c *gin.Context, id int32) {
	var body warehouse.RemoveInventoryJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inv, err := h.svc.AdjustInventory(c.Request.Context(), int64(id), int64(body.ItemId), -int(body.Amount), "MANUAL", 0)
	if err != nil {
		if err == service.ErrInsufficientStock {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inv)
}
