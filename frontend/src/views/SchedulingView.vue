<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { List, Map as MapIcon, Plus, RefreshCw } from 'lucide-vue-next'
import AMapLoader from '@amap/amap-jsapi-loader'
import { useAuthStore } from '@/stores/auth'
import request from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

type Item = { id: number; name: string; unit: string; description?: string }
type ItemQuantity = { itemId: number; quantity: number }
type RescueRequest = { id: number; title: string; location: string; status: string; items?: ItemQuantity[]; assignedTo?: number | null }
type ShipmentTracking = { status: string; location?: string; timestamp?: string }
type Shipment = { shipmentId: number; requestId: number; fromWarehouseId: number; toLocation: string; status: string; tracking?: ShipmentTracking[] }

const authStore = useAuthStore()
const activeTab = ref<'requests' | 'shipments' | 'map'>('requests')

const statusLabelMap: Record<string, string> = {
  PENDING: '待处理',
  ASSIGNED: '已指派',
  COMPLETED: '已完成',
  CANCELLED: '已取消',
  NEW: '新建',
  IN_TRANSIT: '运输中',
  DELIVERED: '已送达',
}

const items = ref<Item[]>([])
const loadItems = async () => {
  const res: any = await request.get('/api/v1/items', { params: { page: 1, size: 1000 } })
  items.value = res?.items ?? []
}

const requestsLoading = ref(false)
const requestsData = ref<RescueRequest[]>([])
const requestsTotal = ref(0)
const requestQuery = reactive({ page: 1, size: 10, status: '' })

const loadRequests = async () => {
  requestsLoading.value = true
  try {
    const res: any = await request.get('/api/v1/requests', { params: { ...requestQuery, status: requestQuery.status || undefined } })
    requestsData.value = res?.requests ?? []
    requestsTotal.value = Number(res?.total) || 0
  } finally {
    requestsLoading.value = false
  }
}

const shipmentsLoading = ref(false)
const shipmentsData = ref<Shipment[]>([])
const shipmentsTotal = ref(0)
const shipmentQuery = reactive({ page: 1, size: 10, status: '' })

const loadShipments = async () => {
  shipmentsLoading.value = true
  try {
    const res: any = await request.get('/api/v1/shipments', { params: { ...shipmentQuery, status: shipmentQuery.status || undefined } })
    shipmentsData.value = res?.shipments ?? []
    shipmentsTotal.value = Number(res?.total) || 0
  } finally {
    shipmentsLoading.value = false
  }
}

const requestDialogVisible = ref(false)
const requestDialogMode = ref<'create' | 'edit'>('create')
const requestForm = reactive({
  id: 0,
  title: '',
  location: '',
  status: '',
  assignedTo: undefined as number | undefined,
  items: [] as { itemId: number | null; quantity: number }[],
})

const openCreateRequest = () => {
  requestDialogMode.value = 'create'
  requestForm.id = 0
  requestForm.title = ''
  requestForm.location = ''
  requestForm.status = ''
  requestForm.assignedTo = undefined
  requestForm.items = [{ itemId: null, quantity: 1 }]
  requestDialogVisible.value = true
}

const openEditRequest = async (row: RescueRequest) => {
  requestDialogMode.value = 'edit'
  const res: any = await request.get(`/api/v1/requests/${row.id}`)
  requestForm.id = res?.id ?? row.id
  requestForm.title = res?.title ?? row.title
  requestForm.location = res?.location ?? row.location
  requestForm.status = res?.status ?? row.status
  requestForm.assignedTo = res?.assignedTo == null ? undefined : Number(res.assignedTo)
  requestForm.items = (res?.items ?? []).map((it: any) => ({ itemId: Number(it.itemId), quantity: Number(it.quantity) }))
  if (requestForm.items.length === 0) requestForm.items = [{ itemId: null, quantity: 1 }]
  requestDialogVisible.value = true
}

const saveRequest = async () => {
  const payloadItems = requestForm.items
    .filter(it => it.itemId != null && it.quantity > 0)
    .map(it => ({ itemId: Number(it.itemId), quantity: Number(it.quantity) }))

  if (requestDialogMode.value === 'create') {
    if (!requestForm.title || !requestForm.location || payloadItems.length === 0) {
      ElMessage.warning('请填写标题、地点和物资明细')
      return
    }
    await request.post('/api/v1/requests', { title: requestForm.title, location: requestForm.location, items: payloadItems })
    ElMessage.success('需求单已创建')
  } else {
    const payload: any = {}
    if (requestForm.status) payload.status = requestForm.status
    if (requestForm.assignedTo != null) payload.assignedTo = requestForm.assignedTo
    await request.put(`/api/v1/requests/${requestForm.id}`, payload)
    ElMessage.success('需求单已更新')
  }
  requestDialogVisible.value = false
  await loadRequests()
}

const deleteRequest = async (row: RescueRequest) => {
  await ElMessageBox.confirm(`确认删除需求单 #${row.id}？仅未指派状态可删除。`, '提示', { type: 'warning' })
  await request.delete(`/api/v1/requests/${row.id}`)
  ElMessage.success('已删除')
  await loadRequests()
}

const shipmentDialogVisible = ref(false)
const shipmentDialogMode = ref<'create' | 'status'>('create')
const shipmentForm = reactive({
  shipmentId: 0,
  requestId: undefined as number | undefined,
  fromWarehouseId: undefined as number | undefined,
  toLocation: '',
  status: 'IN_TRANSIT',
  location: '',
  timestamp: '',
  items: [] as { itemId: number | null; quantity: number }[],
})

const openCreateShipment = () => {
  shipmentDialogMode.value = 'create'
  shipmentForm.shipmentId = 0
  shipmentForm.requestId = undefined
  shipmentForm.fromWarehouseId = undefined
  shipmentForm.toLocation = ''
  shipmentForm.items = [{ itemId: null, quantity: 1 }]
  shipmentDialogVisible.value = true
}

const loadRequestForShipment = async () => {
  if (!shipmentForm.requestId) {
    ElMessage.warning('请先输入需求单 ID')
    return
  }
  const res: any = await request.get(`/api/v1/requests/${shipmentForm.requestId}`)
  shipmentForm.toLocation = res?.location ?? ''
  shipmentForm.items = (res?.items ?? []).map((it: any) => ({ itemId: Number(it.itemId), quantity: Number(it.quantity) }))
  if (shipmentForm.items.length === 0) shipmentForm.items = [{ itemId: null, quantity: 1 }]
  ElMessage.success('已加载需求单明细')
}

const openUpdateShipmentStatus = (row: Shipment) => {
  shipmentDialogMode.value = 'status'
  shipmentForm.shipmentId = row.shipmentId
  shipmentForm.status = row.status
  shipmentForm.location = ''
  shipmentForm.timestamp = ''
  shipmentDialogVisible.value = true
}

const saveShipment = async () => {
  if (shipmentDialogMode.value === 'create') {
    const payloadItems = shipmentForm.items
      .filter(it => it.itemId != null && it.quantity > 0)
      .map(it => ({ itemId: Number(it.itemId), quantity: Number(it.quantity) }))

    if (!shipmentForm.requestId || !shipmentForm.fromWarehouseId || !shipmentForm.toLocation || payloadItems.length === 0) {
      ElMessage.warning('请填写需求单ID、出发仓库、目的地和物资明细')
      return
    }
    await request.post('/api/v1/shipments', {
      requestId: shipmentForm.requestId,
      fromWarehouseId: shipmentForm.fromWarehouseId,
      toLocation: shipmentForm.toLocation,
      items: payloadItems,
    })
    ElMessage.success('运输任务已创建')
  } else {
    if (!shipmentForm.status) {
      ElMessage.warning('请选择状态')
      return
    }
    const payload: any = { status: shipmentForm.status }
    if (shipmentForm.location) payload.location = shipmentForm.location
    if (shipmentForm.timestamp) payload.timestamp = shipmentForm.timestamp
    await request.put(`/api/v1/shipments/${shipmentForm.shipmentId}/status`, payload)
    ElMessage.success('运输状态已更新')
  }
  shipmentDialogVisible.value = false
  await loadShipments()
  if (activeTab.value === 'map' && selectedShipmentId.value) {
    await loadSelectedShipment()
    await renderShipmentOnMap()
  }
}

const selectedShipmentId = ref<number | null>(null)
const selectedShipment = ref<Shipment | null>(null)

const mapContainer = ref<HTMLElement | null>(null)
let map: any = null

const loadSelectedShipment = async () => {
  if (!selectedShipmentId.value) return
  const res: any = await request.get(`/api/v1/shipments/${selectedShipmentId.value}`)
  selectedShipment.value = res
}

const initMap = async () => {
  if (!mapContainer.value) return
  const AMap = await AMapLoader.load({
    key: 'bff42bf37382382b61e29c13b4964ad4',
    version: '2.0',
    plugins: ['AMap.Geocoder'],
  })

  map = new AMap.Map(mapContainer.value, {
    viewMode: '3D',
    zoom: 5,
    center: [104.195397, 35.86166],
    mapStyle: 'amap://styles/dark',
  })
}

const renderShipmentOnMap = async () => {
  if (!selectedShipment.value) return
  if (!map) {
    await nextTick()
    await initMap()
  }
  if (!map) return
  map.clearMap()

  const AMap = (window as any).AMap
  const geocoder = new AMap.Geocoder()

  const tracking = selectedShipment.value.tracking ?? []
  const locations = tracking.map(t => t.location).filter(Boolean) as string[]
  if (locations.length === 0) {
    ElMessage.info('该运输任务暂无可用位置轨迹')
    return
  }

  const points: any[] = []
  for (const loc of locations) {
    const p = await new Promise<any | null>((resolve) => {
      geocoder.getLocation(loc, (status: string, result: any) => {
        if (status === 'complete' && result?.geocodes?.[0]?.location) resolve(result.geocodes[0].location)
        else resolve(null)
      })
    })
    if (p) points.push(p)
  }

  if (points.length === 0) {
    ElMessage.warning('无法将轨迹地点解析为坐标，请检查地点文本')
    return
  }

  const polyline = new AMap.Polyline({
    path: points,
    strokeColor: '#3b82f6',
    strokeOpacity: 1,
    strokeWeight: 6,
    lineJoin: 'round',
    lineCap: 'round',
  })
  map.add(polyline)

  points.forEach((p, idx) => {
    new AMap.Marker({
      position: p,
      map,
      title: idx === 0 ? '起点' : idx === points.length - 1 ? '终点' : `节点${idx + 1}`,
    })
  })
  map.setFitView()
}

const canDeleteRequest = computed(() => authStore.isAdmin)

watch(activeTab, async (tab) => {
  if (tab === 'requests') await loadRequests()
  if (tab === 'shipments') await loadShipments()
  if (tab === 'map') {
    await nextTick()
    if (!map) await initMap()
    if (selectedShipmentId.value) {
      await loadSelectedShipment()
      await renderShipmentOnMap()
    }
  }
})

onMounted(async () => {
  await loadItems()
  await loadRequests()
})
</script>

<template>
  <div class="scheduling-container">
    <div class="tabs-header">
      <div
        v-for="tab in ['requests', 'shipments', 'map']"
        :key="tab"
        :class="['tab-item', { active: activeTab === tab }]"
        @click="activeTab = tab as any"
      >
        <component :is="tab === 'map' ? MapIcon : List" :size="18" />
        {{ tab === 'requests' ? '需求单' : tab === 'shipments' ? '运输任务' : '地图追踪' }}
      </div>
    </div>

    <div v-if="activeTab === 'requests'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select v-model="requestQuery.status" placeholder="状态筛选" clearable style="width: 180px" @change="loadRequests">
            <el-option label="待处理 (PENDING)" value="PENDING" />
            <el-option label="已指派 (ASSIGNED)" value="ASSIGNED" />
            <el-option label="已完成 (COMPLETED)" value="COMPLETED" />
            <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
          </el-select>
          <el-button :icon="RefreshCw" @click="loadRequests">刷新</el-button>
        </div>
        <el-button type="primary" :icon="Plus" @click="openCreateRequest">新建需求单</el-button>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="requestsData" v-loading="requestsLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="title" label="标题" min-width="160" />
          <el-table-column prop="location" label="地点" min-width="120" />
          <el-table-column prop="status" label="状态" width="140">
            <template #default="{ row }">
              <el-tag effect="plain">{{ statusLabelMap[row.status] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="170" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openEditRequest(row)">查看/更新</el-button>
              <el-button v-if="canDeleteRequest" link type="danger" @click="deleteRequest(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <div class="pagination">
        <el-pagination
          background
          layout="prev, pager, next"
          :page-size="requestQuery.size"
          :current-page="requestQuery.page"
          :total="requestsTotal"
          @current-change="(p:number) => { requestQuery.page = p; loadRequests() }"
        />
      </div>
    </div>

    <div v-else-if="activeTab === 'shipments'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select v-model="shipmentQuery.status" placeholder="状态筛选" clearable style="width: 180px" @change="loadShipments">
            <el-option label="新建 (NEW)" value="NEW" />
            <el-option label="运输中 (IN_TRANSIT)" value="IN_TRANSIT" />
            <el-option label="已送达 (DELIVERED)" value="DELIVERED" />
            <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
          </el-select>
          <el-button :icon="RefreshCw" @click="loadShipments">刷新</el-button>
        </div>
        <el-button type="primary" :icon="Plus" @click="openCreateShipment">新建运输任务</el-button>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="shipmentsData" v-loading="shipmentsLoading" style="width: 100%">
          <el-table-column prop="shipmentId" label="运输ID" width="110" />
          <el-table-column prop="requestId" label="需求ID" width="110" />
          <el-table-column prop="fromWarehouseId" label="出发仓库" width="110" />
          <el-table-column prop="toLocation" label="目的地" min-width="140" />
          <el-table-column prop="status" label="状态" width="140">
            <template #default="{ row }">
              <el-tag effect="plain">{{ statusLabelMap[row.status] || row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openUpdateShipmentStatus(row)">更新状态</el-button>
              <el-button
                link
                type="primary"
                @click="async () => { selectedShipmentId = row.shipmentId; activeTab = 'map'; await loadSelectedShipment(); await renderShipmentOnMap() }"
              >
                查看轨迹
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <div class="pagination">
        <el-pagination
          background
          layout="prev, pager, next"
          :page-size="shipmentQuery.size"
          :current-page="shipmentQuery.page"
          :total="shipmentsTotal"
          @current-change="(p:number) => { shipmentQuery.page = p; loadShipments() }"
        />
      </div>
    </div>

    <div v-else class="map-view">
      <div class="map-toolbar">
        <el-input-number v-model="selectedShipmentId" :min="1" controls-position="right" placeholder="运输ID" />
        <el-button
          type="primary"
          @click="async () => { await loadSelectedShipment(); await renderShipmentOnMap() }"
          :disabled="!selectedShipmentId"
        >
          加载并渲染
        </el-button>
      </div>

      <div class="map-body">
        <div ref="mapContainer" class="map-container"></div>
        <div class="map-overlay glass-panel">
          <h3>轨迹信息</h3>
          <p v-if="selectedShipment">运输任务：#{{ selectedShipment.shipmentId }}</p>
          <p v-else>请选择运输任务</p>
          <div v-if="selectedShipment" class="tracking-list">
            <div v-for="(t, idx) in (selectedShipment.tracking || [])" :key="idx" class="tracking-item">
              <div class="tracking-title">{{ statusLabelMap[t.status] || t.status }}</div>
              <div class="tracking-sub">{{ t.location || '-' }} · {{ t.timestamp || '-' }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="requestDialogVisible" :title="requestDialogMode === 'create' ? '新建需求单' : '查看/更新需求单'" width="720px">
      <el-form label-width="90px">
        <el-form-item label="标题" v-if="requestDialogMode === 'create'">
          <el-input v-model="requestForm.title" placeholder="请输入需求标题" />
        </el-form-item>
        <el-form-item label="地点" v-if="requestDialogMode === 'create'">
          <el-input v-model="requestForm.location" placeholder="请输入需求地点" />
        </el-form-item>

        <el-form-item label="状态" v-if="requestDialogMode === 'edit'">
          <el-select v-model="requestForm.status" placeholder="选择状态" clearable style="width: 220px">
            <el-option label="待处理 (PENDING)" value="PENDING" />
            <el-option label="已指派 (ASSIGNED)" value="ASSIGNED" />
            <el-option label="已完成 (COMPLETED)" value="COMPLETED" />
            <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派给" v-if="requestDialogMode === 'edit'">
          <el-input-number v-model="requestForm.assignedTo" :min="1" controls-position="right" placeholder="用户ID" />
        </el-form-item>

        <el-form-item label="物资明细" v-if="requestDialogMode === 'create'">
          <div class="items-editor">
            <div v-for="(it, idx) in requestForm.items" :key="idx" class="items-row">
              <el-select v-model="it.itemId" placeholder="选择物资" filterable style="width: 360px">
                <el-option v-for="opt in items" :key="opt.id" :label="`${opt.name}（${opt.unit}）`" :value="opt.id" />
              </el-select>
              <el-input-number v-model="it.quantity" :min="1" controls-position="right" />
              <el-button link type="danger" :disabled="requestForm.items.length <= 1" @click="requestForm.items.splice(idx, 1)">移除</el-button>
            </div>
            <el-button link type="primary" @click="requestForm.items.push({ itemId: null, quantity: 1 })">添加一行</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="requestDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRequest">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="shipmentDialogVisible" :title="shipmentDialogMode === 'create' ? '新建运输任务' : '更新运输状态'" width="760px">
      <el-form label-width="110px">
        <template v-if="shipmentDialogMode === 'create'">
          <el-form-item label="需求单ID">
            <el-input-number v-model="shipmentForm.requestId" :min="1" controls-position="right" />
            <el-button link type="primary" @click="loadRequestForShipment" style="margin-left: 8px">加载需求明细</el-button>
          </el-form-item>
          <el-form-item label="出发仓库ID">
            <el-input-number v-model="shipmentForm.fromWarehouseId" :min="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="目的地">
            <el-input v-model="shipmentForm.toLocation" placeholder="请输入目的地" />
          </el-form-item>
          <el-form-item label="物资明细">
            <div class="items-editor">
              <div v-for="(it, idx) in shipmentForm.items" :key="idx" class="items-row">
                <el-select v-model="it.itemId" placeholder="选择物资" filterable style="width: 360px">
                  <el-option v-for="opt in items" :key="opt.id" :label="`${opt.name}（${opt.unit}）`" :value="opt.id" />
                </el-select>
                <el-input-number v-model="it.quantity" :min="1" controls-position="right" />
                <el-button link type="danger" :disabled="shipmentForm.items.length <= 1" @click="shipmentForm.items.splice(idx, 1)">移除</el-button>
              </div>
              <el-button link type="primary" @click="shipmentForm.items.push({ itemId: null, quantity: 1 })">添加一行</el-button>
            </div>
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="状态">
            <el-select v-model="shipmentForm.status" placeholder="选择状态" style="width: 240px">
              <el-option label="新建 (NEW)" value="NEW" />
              <el-option label="运输中 (IN_TRANSIT)" value="IN_TRANSIT" />
              <el-option label="已送达 (DELIVERED)" value="DELIVERED" />
              <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
            </el-select>
          </el-form-item>
          <el-form-item label="当前位置">
            <el-input v-model="shipmentForm.location" placeholder="可选：例如 成都/西安/北京" />
          </el-form-item>
          <el-form-item label="时间戳">
            <el-input v-model="shipmentForm.timestamp" placeholder="可选：ISO8601，例如 2025-12-01T14:30:00Z" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="shipmentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveShipment">保存</el-button>
      </template>
    </el-dialog>
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

.panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.table-card {
  border: 1px solid #30363d;
  background: #161b22;
}

.pagination {
  display: flex;
  justify-content: flex-end;
}

.map-view {
  flex: 1;
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #30363d;
}

.map-toolbar {
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 5;
  display: flex;
  gap: 10px;
  align-items: center;
}

.map-body {
  width: 100%;
  height: 100%;
}

.map-container {
  width: 100%;
  height: 100%;
}

.map-overlay {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 320px;
  max-height: calc(100% - 40px);
  overflow: auto;
  padding: 20px;
  border-radius: 12px;
  color: white;
}

.map-overlay h3 { margin-top: 0; font-size: 1.1rem; }
.map-overlay p { margin: 8px 0; color: #8b949e; }

.tracking-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tracking-item {
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(13, 17, 23, 0.45);
  border: 1px solid rgba(48, 54, 61, 0.8);
}

.tracking-title {
  font-weight: 600;
  color: #e6edf3;
}

.tracking-sub {
  margin-top: 4px;
  font-size: 0.85rem;
  color: #8b949e;
}

.items-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.items-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
