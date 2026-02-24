<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { 
  Package, 
  Truck, 
  AlertTriangle, 
  CheckCircle2,
  TrendingUp,
  Box
} from 'lucide-vue-next'
import request from '@/api/request'

const stats = ref({
  totalItems: 0,
  activeShipments: 0,
  pendingRequests: 0,
  alerts: 0
})

const barChartRef = ref<HTMLElement | null>(null)
const pieChartRef = ref<HTMLElement | null>(null)

const initCharts = () => {
  if (barChartRef.value) {
    const chart = echarts.init(barChartRef.value, 'dark')
    chart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'] },
      yAxis: { type: 'value' },
      series: [{
        data: [120, 200, 150, 80, 70, 110, 130],
        type: 'bar',
        itemStyle: { color: '#58a6ff' }
      }]
    })
  }
  
  if (pieChartRef.value) {
    const chart = echarts.init(pieChartRef.value, 'dark')
    chart.setOption({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        data: [
          { value: 1048, name: 'Medical' },
          { value: 735, name: 'Food' },
          { value: 580, name: 'Shelter' },
          { value: 484, name: 'Water' }
        ],
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
    const res = await request.get('/api/v1/stats/summary')
    // Mocking for now if endpoint not ready
    stats.value = {
      totalItems: 1250,
      activeShipments: 8,
      pendingRequests: 12,
      alerts: 3
    }
  } catch (err) {
    // Fallback/Mock
    stats.value = {
      totalItems: 1250,
      activeShipments: 8,
      pendingRequests: 12,
      alerts: 3
    }
  }
}

onMounted(() => {
  fetchStats()
  initCharts()
})
</script>

<template>
  <div class="dashboard-container">
    <div class="stats-grid">
      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>Total Inventory</span>
            <Box class="icon blue" />
          </div>
        </template>
        <div class="card-value">{{ stats.totalItems }}</div>
        <div class="card-desc"><TrendingUp :size="14" /> +12% from last week</div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>In Transit</span>
            <Truck class="icon purple" />
          </div>
        </template>
        <div class="card-value">{{ stats.activeShipments }}</div>
        <div class="card-desc">Active logistics tasks</div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>Pending Requests</span>
            <Package class="icon orange" />
          </div>
        </template>
        <div class="card-value">{{ stats.pendingRequests }}</div>
        <div class="card-desc">Awaiting approval</div>
      </el-card>

      <el-card class="dashboard-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>Critical Alerts</span>
            <AlertTriangle class="icon red" />
          </div>
        </template>
        <div class="card-value">{{ stats.alerts }}</div>
        <div class="card-desc">Stock below threshold</div>
      </el-card>
    </div>

    <div class="charts-grid">
      <el-card class="chart-card" shadow="never">
        <template #header>Material Consumption Trend</template>
        <div ref="barChartRef" class="chart-container"></div>
      </el-card>
      
      <el-card class="chart-card" shadow="never">
        <template #header>Category Distribution</template>
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
