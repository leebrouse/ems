<script setup lang="ts">
/**
 * 概览页：
 * - 聚合展示库存总量、运输中任务、待处理需求、库存预警
 * - 使用 ECharts 渲染运输趋势与库存 Top 分布
 * - 兼容后端返回字段大小写差异（normalize*）
 */
import { ref, onMounted } from "vue";
import * as echarts from "echarts";
import {
  Package,
  Truck,
  AlertTriangle,
  TrendingUp,
  Box,
} from "lucide-vue-next";
import request from "../api/request";

type InventoryStat = { itemId: number; name: string; totalQuantity: number };
type ShipmentStatPoint = { week?: string; periodLabel?: string; count: number };
type Warehouse = { id: number; name: string; location?: string };
type InventoryRow = { itemId: number; name: string; quantity: number };

// 兼容后端字段命名差异（例如 id/ID、name/Name）
const normalizeWarehouse = (raw: any): Warehouse => ({
  id: Number(raw?.id ?? raw?.ID ?? 0),
  name: String(raw?.name ?? raw?.Name ?? ""),
  location: raw?.location ?? raw?.Location ?? undefined,
});

const normalizeInventoryRow = (raw: any): InventoryRow => {
  const item = raw?.item ?? raw?.Item ?? {};
  return {
    itemId: Number(raw?.itemId ?? raw?.ItemID ?? item?.id ?? item?.ID ?? 0),
    name: String(raw?.name ?? raw?.Name ?? item?.name ?? item?.Name ?? ""),
    quantity: Number(raw?.quantity ?? raw?.Quantity ?? 0),
  };
};

const normalizeInventoryStat = (raw: any): InventoryStat => ({
  itemId: Number(raw?.itemId ?? raw?.ItemID ?? 0),
  name: String(raw?.name ?? raw?.Name ?? ""),
  totalQuantity: Number(
    raw?.totalQuantity ??
      raw?.TotalQuantity ??
      raw?.quantity ??
      raw?.Quantity ??
      0
  ),
});

const stats = ref({
  totalInventory: 0,
  inTransitShipments: 0,
  pendingRequests: 0,
  alerts: 0,
});

const barChartRef = ref<HTMLElement | null>(null);
const pieChartRef = ref<HTMLElement | null>(null);

// 根据接口数据初始化/更新图表
const initCharts = (
  shipmentSeries?: { labels: string[]; counts: number[] },
  inventorySeries?: { labels: string[]; values: number[] }
) => {
  if (barChartRef.value) {
    const chart = echarts.init(barChartRef.value, "dark");
    const labels = shipmentSeries?.labels ?? [];
    const counts = shipmentSeries?.counts ?? [];
    chart.setOption({
      backgroundColor: "transparent",
      tooltip: { trigger: "axis" },
      xAxis: { type: "category", data: labels },
      yAxis: { type: "value" },
      series: [
        {
          data: counts,
          type: "line",
          smooth: true,
          itemStyle: { color: "#58a6ff" },
        },
      ],
    });
  }

  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value, "dark");
    const labels = inventorySeries?.labels ?? [];
    const values = inventorySeries?.values ?? [];
    chart.setOption({
      backgroundColor: "transparent",
      tooltip: { trigger: "item" },
      series: [
        {
          type: "pie",
          radius: ["40%", "70%"],
          data: labels.map((name, i) => ({ name, value: values[i] })),
          itemStyle: {
            borderRadius: 8,
            borderColor: "#161b22",
            borderWidth: 2,
          },
        },
      ],
    });
  }
};

// 兜底库存统计：逐仓库拉取库存明细并聚合
const fetchInventoryFromWarehouses = async (): Promise<InventoryStat[]> => {
  const warehousesRes: any = await request.get("/api/v1/warehouses");
  const warehousesRaw =
    warehousesRes?.warehouses ??
    warehousesRes?.Warehouses ??
    warehousesRes ??
    [];
  const warehouses = Array.isArray(warehousesRaw)
    ? warehousesRaw.map(normalizeWarehouse)
    : [];
  if (!warehouses.length) return [];

  const inventoryResponses = await Promise.all(
    warehouses.map((warehouse) =>
      request
        .get(`/api/v1/warehouses/${warehouse.id}/inventory`)
        .catch(() => [])
    )
  );

  const aggregates = new Map<number, InventoryStat>();
  inventoryResponses.forEach((res: any) => {
    const list = res?.inventory ?? res?.Inventory ?? res ?? [];
    const rows = Array.isArray(list) ? list.map(normalizeInventoryRow) : [];
    rows.forEach((row) => {
      const current = aggregates.get(row.itemId);
      if (!current) {
        aggregates.set(row.itemId, {
          itemId: row.itemId,
          name: row.name,
          totalQuantity: row.quantity,
        });
        return;
      }
      aggregates.set(row.itemId, {
        itemId: row.itemId,
        name: current.name || row.name,
        totalQuantity: current.totalQuantity + row.quantity,
      });
    });
  });

  return [...aggregates.values()];
};

// 优先使用 stats/inventory，失败则降级到仓库库存聚合
const fetchInventoryStats = async (): Promise<InventoryStat[]> => {
  try {
    const inventoryStatsRes: any = await request.get("/api/v1/stats/inventory");
    const list = Array.isArray(inventoryStatsRes)
      ? inventoryStatsRes
      : inventoryStatsRes?.items ?? inventoryStatsRes?.Items ?? [];
    const stats = Array.isArray(list) ? list.map(normalizeInventoryStat) : [];
    const total = stats.reduce(
      (sum, item) => sum + (Number(item.totalQuantity) || 0),
      0
    );
    if (stats.length && total > 0) return stats;
  } catch {}
  return await fetchInventoryFromWarehouses();
};

// 拉取仪表盘所需的全部数据，并驱动图表渲染
const fetchStats = async () => {
  try {
    const [
      inventoryStats,
      inTransitShipmentsRes,
      pendingRequestsRes,
      alertsRes,
      shipmentStatsRes,
    ] = await Promise.all([
      fetchInventoryStats(),
      request.get("/api/v1/shipments", {
        params: { page: 1, size: 1, status: "IN_TRANSIT" },
      }),
      request.get("/api/v1/requests", {
        params: { page: 1, size: 1, status: "PENDING" },
      }),
      request.get("/api/v1/alerts"),
      request.get("/api/v1/stats/shipments", { params: { period: "weekly" } }),
    ]);

    const inTransitTotal = Number((inTransitShipmentsRes as any)?.total) || 0;
    const pendingTotal = Number((pendingRequestsRes as any)?.total) || 0;
    const shipmentStats = shipmentStatsRes as any as {
      period: string;
      data: ShipmentStatPoint[];
    };

    const totalInventory = (inventoryStats ?? []).reduce(
      (sum: number, item: InventoryStat) =>
        sum + (Number(item.totalQuantity) || 0),
      0
    );
    const alertsList = Array.isArray(alertsRes)
      ? alertsRes
      : (alertsRes as any)?.alerts ?? (alertsRes as any)?.Alerts ?? [];
    const alertsCount = Array.isArray(alertsList) ? alertsList.length : 0;

    const shipmentPoints = Array.isArray(shipmentStats?.data)
      ? shipmentStats.data
      : [];
    const shipmentLabels = shipmentPoints
      .map((p) => p.week || p.periodLabel || "")
      .filter(Boolean);
    const shipmentCounts = shipmentPoints.map((p) => Number(p.count) || 0);

    const topInventory = [...(inventoryStats ?? [])]
      .sort(
        (a, b) =>
          (Number(b.totalQuantity) || 0) - (Number(a.totalQuantity) || 0)
      )
      .slice(0, 6);
    const inventoryLabels = topInventory.map((i) => i.name);
    const inventoryValues = topInventory.map(
      (i) => Number(i.totalQuantity) || 0
    );

    stats.value = {
      totalInventory,
      inTransitShipments: inTransitTotal,
      pendingRequests: pendingTotal,
      alerts: alertsCount,
    };

    initCharts(
      { labels: shipmentLabels, counts: shipmentCounts },
      { labels: inventoryLabels, values: inventoryValues }
    );
  } catch {
    initCharts({ labels: [], counts: [] }, { labels: [], values: [] });
  }
};

// 页面加载后拉取统计数据
onMounted(() => {
  fetchStats();
});
</script>

<template>
  <div class="dashboard-container">
    <div class="stats-grid">
      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>库存总量</span>
            <Box class="icon blue" />
          </div>
        </template>
        <div class="card-value">{{ stats.totalInventory }}</div>
        <div class="card-desc">
          <TrendingUp :size="14" /> 统计来自仓库库存汇总
        </div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>运输中任务</span>
            <Truck class="icon purple" />
          </div>
        </template>
        <div class="card-value">{{ stats.inTransitShipments }}</div>
        <div class="card-desc">状态为 IN_TRANSIT</div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>待处理需求</span>
            <Package class="icon orange" />
          </div>
        </template>
        <div class="card-value">{{ stats.pendingRequests }}</div>
        <div class="card-desc">状态为 PENDING</div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>库存预警</span>
            <AlertTriangle class="icon red" />
          </div>
        </template>
        <div class="card-value">{{ stats.alerts }}</div>
        <div class="card-desc">低于阈值的物资</div>
      </el-card>
    </div>

    <div class="charts-grid">
      <el-card class="chart-card" shadow="never">
        <template #header>运输量趋势（按周）</template>
        <div ref="barChartRef" class="chart-container"></div>
      </el-card>

      <el-card class="chart-card" shadow="never">
        <template #header>库存分布（Top）</template>
        <div ref="pieChartRef" class="chart-container"></div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.dashboard-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #8b949e;
  font-size: 0.9rem;
}

.card-value {
  font-size: 2rem;
  font-weight: bold;
  margin: 8px 0;
  color: #e6edf3;
}

.card-desc {
  font-size: 0.8rem;
  color: #8b949e;
  display: flex;
  align-items: center;
  gap: 4px;
}

.icon {
  width: 20px;
  height: 20px;
}

.blue {
  color: #58a6ff;
}
.purple {
  color: #bc8cff;
}
.orange {
  color: #ffa657;
}
.red {
  color: #ff7b72;
}

.charts-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
}

.chart-card {
  height: 400px;
}

.chart-container {
  height: 320px;
  width: 100%;
}

:deep(.el-card__header) {
  border-bottom: 1px solid #30363d;
  padding: 12px 20px;
  font-weight: 500;
}
</style>
