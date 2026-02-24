<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, List, Map as MapIcon, ChevronRight } from 'lucide-vue-next'
import AMapLoader from '@amap/amap-jsapi-loader'
import { useAuthStore } from '@/stores/auth'

const activeTab = ref('list')
const shipments = ref([
  { id: 5001, status: 'IN_TRANSIT', from: 'Main Warehouse', to: 'Chengdu Office', progress: 65 },
  { id: 5002, status: 'DELIVERED', from: 'Northeast Depot', to: 'Harbin Rescue Center', progress: 100 },
  { id: 5003, status: 'NEW', from: 'Central Logistics', to: 'Wuhan Shelter 3', progress: 0 },
])

const mapContainer = ref<HTMLElement | null>(null)
let map: any = null

const initMap = async () => {
  if (activeTab.value !== 'map' || !mapContainer.value) return

  try {
    const AMap = await AMapLoader.load({
      key: 'bff42bf37382382b61e29c13b4964ad4',
      version: '2.0',
      plugins: ['AMap.Driving', 'AMap.PolyEditor']
    })

    map = new AMap.Map(mapContainer.value, {
      viewMode: '3D',
      zoom: 11,
      center: [116.397428, 39.90923], // Beijing
      mapStyle: 'amap://styles/dark'
    })

    // Mock path
    const path = [
      [116.368904, 39.913423],
      [116.382122, 39.901176],
      [116.387271, 39.912501],
      [116.398258, 39.904600]
    ]

    const polyline = new AMap.Polyline({
      path: path,
      isOutline: true,
      outlineColor: '#ffeeff',
      borderWeight: 1,
      strokeColor: '#3b82f6',
      strokeOpacity: 1,
      strokeWeight: 6,
      strokeStyle: 'solid',
      lineJoin: 'round',
      lineCap: 'round',
    })

    map.add(polyline)
    map.setFitView()
    
    // Add markers for start and end
    new AMap.Marker({
      position: path[0],
      map: map,
      title: 'Start'
    })
    new AMap.Marker({
      position: path[path.length - 1],
      map: map,
      title: 'End'
    })

  } catch (e) {
    console.error(e)
  }
}

const handleTabChange = (tab: string) => {
  activeTab.value = tab
  if (tab === 'map') {
    setTimeout(initMap, 100)
  }
}
</script>

<template>
  <div class="scheduling-container">
    <div class="tabs-header">
      <div 
        v-for="tab in ['list', 'map']" 
        :key="tab"
        :class="['tab-item', { active: activeTab === tab }]"
        @click="handleTabChange(tab)"
      >
        <component :is="tab === 'list' ? List : MapIcon" :size="18" />
        {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
      </div>
    </div>

    <div v-if="activeTab === 'list'" class="shipment-list">
      <el-card v-for="s in shipments" :key="s.id" class="shipment-card" shadow="never">
        <div class="card-content">
          <div class="shipment-info">
            <div class="id-tag">#{{ s.id }}</div>
            <div class="route">
              <span class="loc">{{ s.from }}</span>
              <ChevronRight :size="16" class="arrow" />
              <span class="loc">{{ s.to }}</span>
            </div>
          </div>
          <div class="status-info">
            <el-tag :type="s.status === 'DELIVERED' ? 'success' : 'primary'" effect="dark">
              {{ s.status }}
            </el-tag>
            <div class="progress-container">
              <el-progress :percentage="s.progress" :show-text="false" />
            </div>
          </div>
          <el-button link type="primary" @click="handleTabChange('map')">View Track</el-button>
        </div>
      </el-card>
    </div>

    <div v-show="activeTab === 'map'" class="map-view">
      <div ref="mapContainer" class="map-container"></div>
      <div class="map-overlay glass-panel">
        <h3>Live Tracking</h3>
        <p>Shipment: #5001</p>
        <div class="info-row">
          <span>Speed:</span>
          <span>45 km/h</span>
        </div>
        <div class="info-row">
          <span>Est. Arrival:</span>
          <span>14:30</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scheduling-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  gap: 20px;
}

.tabs-header {
  display: flex;
  gap: 4px;
  background: #0d1117;
  padding: 4px;
  border-radius: 8px;
  width: fit-content;
}

.tab-item {
  padding: 8px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 6px;
  cursor: pointer;
  color: #8b949e;
  transition: all 0.2s;
}

.tab-item:hover { color: #e6edf3; }
.tab-item.active {
  background: #21262d;
  color: #58a6ff;
}

.shipment-list {
  display: grid;
  gap: 16px;
}

.shipment-card {
  border: 1px solid #30363d;
  background: #161b22;
}

.card-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.shipment-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.id-tag {
  font-family: monospace;
  color: #8b949e;
}

.route {
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 500;
  color: #e6edf3;
}

.arrow { color: #8b949e; }

.status-info {
  width: 200px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.progress-container { width: 100%; }

.map-view {
  flex: 1;
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #30363d;
}

.map-container {
  width: 100%;
  height: 100%;
}

.map-overlay {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 240px;
  padding: 20px;
  border-radius: 12px;
  color: white;
}

.map-overlay h3 { margin-top: 0; font-size: 1.1rem; }
.map-overlay p { margin: 8px 0; color: #8b949e; }

.info-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
  margin-top: 8px;
}
</style>
