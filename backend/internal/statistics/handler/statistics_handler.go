package handler

import (
	"net/http"

	"github.com/leebrouse/ems/backend/common/genopenapi/statistics"
	"github.com/leebrouse/ems/backend/statistics/service"

	"github.com/gin-gonic/gin"
)

// StatisticsHandler 处理统计相关的 HTTP 请求
type StatisticsHandler struct {
	svc service.StatisticsService
}

// NewStatisticsHandler 创建 StatisticsHandler 实例
func NewStatisticsHandler(svc service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{svc: svc}
}

// Ensure StatisticsHandler implements statistics.ServerInterface
var _ statistics.ServerInterface = (*StatisticsHandler)(nil)

// GetInventoryStats 获取库存统计
func (h *StatisticsHandler) GetInventoryStats(c *gin.Context) {
	stats, err := h.svc.GetInventoryStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetRequestStats 获取需求统计
func (h *StatisticsHandler) GetRequestStats(c *gin.Context, params statistics.GetRequestStatsParams) {
	start := ""
	if params.StartDate != nil {
		start = *params.StartDate
	}
	end := ""
	if params.EndDate != nil {
		end = *params.EndDate
	}

	stats, err := h.svc.GetRequestStats(c.Request.Context(), start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetShipmentStats 获取运输统计
func (h *StatisticsHandler) GetShipmentStats(c *gin.Context, params statistics.GetShipmentStatsParams) {
	period := "weekly"
	if params.Period != nil {
		period = string(*params.Period)
	}

	stats, err := h.svc.GetShipmentStats(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
