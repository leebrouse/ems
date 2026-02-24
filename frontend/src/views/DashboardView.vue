<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { 
  Package, 
  Truck, 
  AlertTriangle, 
  TrendingUp,
  Box
} from 'lucide-vue-next'
import request from '@/api/request'

type InventoryStat = { itemId: number; name: string; totalQuantity: number }
type ShipmentStatPoint = { week?: string; periodLabel?: string; count: number }

const stats = ref({
  totalInventory: 0,
  inTransitShipments: 0,
  pendingRequests: 0,
  alerts: 0,
})

const barChartRef = ref<HTMLElement | null>(null)
const pieChartRef = ref<HTMLElement | null>(null)

const initCharts = (shipmentSeries?: { labels: string[]; counts: number[] }, inventorySeries?: { labels: string[]; values: number[] }) => {
  if (barChartRef.value) {
    const chart = echarts.init(barChartRef.value, 'dark')
    const labels = shipmentSeries?.labels ?? []
    const counts = shipmentSeries?.counts ?? []
    chart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: labels },
      yAxis: { type: 'value' },
      series: [{
        data: counts,
        type: 'line',
        smooth: true,
        itemStyle: { color: '#58a6ff' }
      }]
    })
  }
  
  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value, 'dark')
    const labels = inventorySeries?.labels ?? []
    const values = inventorySeries?.values ?? []
    chart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        data: labels.map((name, i) => ({ name, value: values[i] })),
        itemStyle: {
          borderRadius: 8,
          borderColor: '#161b22',
          borderWidth: 2
        }
      }]
    })
  }
}

const fetchStats = async () => {
  try {
    const [inventoryStatsRes, inTransitShipmentsRes, pendingRequestsRes, alertsRes, shipmentStatsRes] = await Promise.all([
      request.get('/api/v1/stats/inventory'),
      request.get('/api/v1/shipments', { params: { page: 1, size: 1, status: 'IN_TRANSIT' } }),
      request.get('/api/v1/requests', { params: { page: 1, size: 1, status: 'PENDING' } }),
      request.get('/api/v1/alerts'),
      request.get('/api/v1/stats/shipments', { params: { period: 'weekly' } }),
    ])

    const inventoryStats = (inventoryStatsRes as any) as InventoryStat[]
    const inTransitTotal = Number((inTransitShipmentsRes as any)?.total) || 0
    const pendingTotal = Number((pendingRequestsRes as any)?.total) || 0
    const shipmentStats = (shipmentStatsRes as any) as { period: string; data: ShipmentStatPoint[] }

    const totalInventory = (inventoryStats ?? []).reduce((sum: number, item: InventoryStat) => sum + (Number(item.totalQuantity) || 0), 0)
    const alertsCount = Array.isArray(alertsRes) ? (alertsRes as any[]).length : 0

    const shipmentPoints = Array.isArray(shipmentStats?.data) ? shipmentStats.data : []
    const shipmentLabels = shipmentPoints.map(p => p.week || p.periodLabel || '').filter(Boolean)
    const shipmentCounts = shipmentPoints.map(p => Number(p.count) || 0)

    const topInventory = [...(inventoryStats ?? [])]
      .sort((a, b) => (Number(b.totalQuantity) || 0) - (Number(a.totalQuantity) || 0))
      .slice(0, 6)
    const inventoryLabels = topInventory.map(i => i.name)
    const inventoryValues = topInventory.map(i => Number(i.totalQuantity) || 0)

    stats.value = {
      totalInventory,
      inTransitShipments: inTransitTotal,
      pendingRequests: pendingTotal,
      alerts: alertsCount,
    }

    initCharts(
      { labels: shipmentLabels, counts: shipmentCounts },
      { labels: inventoryLabels, values: inventoryValues },
    )
  } catch (err) {
    initCharts({ labels: [], counts: [] }, { labels: [], values: [] })
  }
}

onMounted(() => {
  fetchStats()
})
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
        <div class="card-desc"><TrendingUp :size="14" /> 统计来自库存汇总接口</div>
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

.blue { color: #58a6ff; }
.purple { color: #bc8cff; }
.orange { color: #ffa657; }
.red { color: #ff7b72; }

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
