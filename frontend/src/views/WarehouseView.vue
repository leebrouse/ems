<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Plus, Search, Edit, Trash2, RefreshCw, Minus, Plus as PlusIcon, AlertTriangle } from 'lucide-vue-next'
import request from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

type Item = { id: number; name: string; unit: string; description?: string }
type Warehouse = { id: number; name: string; location?: string }
type InventoryRow = { itemId: number; name: string; quantity: number }
type AlertRow = { itemId: number; name: string; quantity: number; threshold: number }

const authStore = useAuthStore()
const activeTab = ref<'items' | 'warehouses' | 'inventory' | 'alerts'>('items')

const itemsLoading = ref(false)
const itemsData = ref<Item[]>([])
const itemsTotal = ref(0)
const itemsQuery = reactive({ page: 1, size: 20, query: '' })
const itemsOptions = ref<Item[]>([])

const warehousesLoading = ref(false)
const warehousesData = ref<Warehouse[]>([])

const inventoryLoading = ref(false)
const inventoryData = ref<InventoryRow[]>([])
const selectedWarehouseId = ref<number | null>(null)

const alertsLoading = ref(false)
const alertsData = ref<AlertRow[]>([])

const loadItems = async () => {
  itemsLoading.value = true
  try {
    const res: any = await request.get('/api/v1/items', { params: { ...itemsQuery, query: itemsQuery.query || undefined } })
    itemsData.value = res?.items ?? []
    itemsTotal.value = Number(res?.total) || 0
  } finally {
    itemsLoading.value = false
  }
}

const loadItemsOptions = async () => {
  const res: any = await request.get('/api/v1/items', { params: { page: 1, size: 1000 } })
  itemsOptions.value = res?.items ?? []
}

const loadWarehouses = async () => {
  warehousesLoading.value = true
  try {
    const res: any = await request.get('/api/v1/warehouses')
    warehousesData.value = res ?? []
    const first = warehousesData.value[0]
    if (!selectedWarehouseId.value && first) selectedWarehouseId.value = first.id
  } finally {
    warehousesLoading.value = false
  }
}

const loadInventory = async () => {
  if (!selectedWarehouseId.value) return
  inventoryLoading.value = true
  try {
    const res: any = await request.get(`/api/v1/warehouses/${selectedWarehouseId.value}/inventory`)
    inventoryData.value = res ?? []
  } finally {
    inventoryLoading.value = false
  }
}

const loadAlerts = async () => {
  alertsLoading.value = true
  try {
    const res: any = await request.get('/api/v1/alerts')
    alertsData.value = res ?? []
  } finally {
    alertsLoading.value = false
  }
}

const itemDialogVisible = ref(false)
const itemDialogMode = ref<'create' | 'edit'>('create')
const itemForm = reactive({ id: 0, name: '', unit: '', description: '' })

const openCreateItem = () => {
  itemDialogMode.value = 'create'
  itemForm.id = 0
  itemForm.name = ''
  itemForm.unit = ''
  itemForm.description = ''
  itemDialogVisible.value = true
}

const openEditItem = (row: Item) => {
  itemDialogMode.value = 'edit'
  itemForm.id = row.id
  itemForm.name = row.name
  itemForm.unit = row.unit
  itemForm.description = row.description ?? ''
  itemDialogVisible.value = true
}

const saveItem = async () => {
  if (!itemForm.name || !itemForm.unit) {
    ElMessage.warning('请填写物资名称和单位')
    return
  }
  if (itemDialogMode.value === 'create') {
    await request.post('/api/v1/items', { name: itemForm.name, unit: itemForm.unit, description: itemForm.description || undefined })
    ElMessage.success('物资已创建')
  } else {
    await request.put(`/api/v1/items/${itemForm.id}`, { name: itemForm.name, unit: itemForm.unit, description: itemForm.description || undefined })
    ElMessage.success('物资已更新')
  }
  itemDialogVisible.value = false
  await loadItems()
  await loadItemsOptions()
}

const deleteItem = async (row: Item) => {
  await ElMessageBox.confirm(`确认删除物资「${row.name}」（ID: ${row.id}）？`, '提示', { type: 'warning' })
  await request.delete(`/api/v1/items/${row.id}`)
  ElMessage.success('已删除')
  await loadItems()
  await loadItemsOptions()
}

const warehouseDialogVisible = ref(false)
const warehouseDialogMode = ref<'create' | 'edit'>('create')
const warehouseForm = reactive({ id: 0, name: '', location: '' })

const openCreateWarehouse = () => {
  warehouseDialogMode.value = 'create'
  warehouseForm.id = 0
  warehouseForm.name = ''
  warehouseForm.location = ''
  warehouseDialogVisible.value = true
}

const openEditWarehouse = (row: Warehouse) => {
  warehouseDialogMode.value = 'edit'
  warehouseForm.id = row.id
  warehouseForm.name = row.name
  warehouseForm.location = row.location ?? ''
  warehouseDialogVisible.value = true
}

const saveWarehouse = async () => {
  if (!warehouseForm.name) {
    ElMessage.warning('请填写仓库名称')
    return
  }
  if (warehouseDialogMode.value === 'create') {
    await request.post('/api/v1/warehouses', { name: warehouseForm.name, location: warehouseForm.location || undefined })
    ElMessage.success('仓库已创建')
  } else {
    await request.put(`/api/v1/warehouses/${warehouseForm.id}`, { name: warehouseForm.name, location: warehouseForm.location || undefined })
    ElMessage.success('仓库已更新')
  }
  warehouseDialogVisible.value = false
  await loadWarehouses()
}

const deleteWarehouse = async (row: Warehouse) => {
  await ElMessageBox.confirm(`确认删除仓库「${row.name}」（ID: ${row.id}）？`, '提示', { type: 'warning' })
  await request.delete(`/api/v1/warehouses/${row.id}`)
  ElMessage.success('已删除')
  await loadWarehouses()
}

const adjustDialogVisible = ref(false)
const adjustMode = ref<'add' | 'remove'>('add')
const adjustForm = reactive({ itemId: null as number | null, amount: 1 })

const openAdjust = (mode: 'add' | 'remove') => {
  if (!selectedWarehouseId.value) {
    ElMessage.warning('请先选择仓库')
    return
  }
  adjustMode.value = mode
  adjustForm.itemId = null
  adjustForm.amount = 1
  adjustDialogVisible.value = true
}

const saveAdjust = async () => {
  if (!selectedWarehouseId.value || !adjustForm.itemId || adjustForm.amount <= 0) {
    ElMessage.warning('请选择物资并填写数量')
    return
  }
  const url = adjustMode.value === 'add'
    ? `/api/v1/warehouses/${selectedWarehouseId.value}/inventory/add`
    : `/api/v1/warehouses/${selectedWarehouseId.value}/inventory/remove`

  await request.post(url, { itemId: adjustForm.itemId, amount: adjustForm.amount })
  ElMessage.success(adjustMode.value === 'add' ? '入库成功' : '出库成功')
  adjustDialogVisible.value = false
  await loadInventory()
  await loadAlerts()
}

const thresholdDialogVisible = ref(false)
const thresholdForm = reactive({ itemId: null as number | null, threshold: 10 })

const openSetThreshold = () => {
  thresholdForm.itemId = null
  thresholdForm.threshold = 10
  thresholdDialogVisible.value = true
}

const saveThreshold = async () => {
  if (!thresholdForm.itemId || thresholdForm.threshold < 0) {
    ElMessage.warning('请选择物资并填写阈值')
    return
  }
  await request.put('/api/v1/alerts/threshold', { itemId: thresholdForm.itemId, threshold: thresholdForm.threshold })
  ElMessage.success('阈值已设置')
  thresholdDialogVisible.value = false
  await loadAlerts()
}

const alertRowClassName = ({ row }: { row: AlertRow }) => {
  if (row.quantity < row.threshold) return 'warning-row'
  return ''
}

const canAdminManage = computed(() => authStore.isAdmin)
const canWarehouseManage = computed(() => authStore.isAdmin || authStore.isWarehouseManager)

watch(activeTab, async (tab) => {
  if (tab === 'items') await loadItems()
  if (tab === 'warehouses') await loadWarehouses()
  if (tab === 'inventory') {
    if (warehousesData.value.length === 0) await loadWarehouses()
    await loadInventory()
  }
  if (tab === 'alerts') await loadAlerts()
})

watch(selectedWarehouseId, async () => {
  if (activeTab.value === 'inventory') await loadInventory()
})

onMounted(async () => {
  await loadItems()
  await loadItemsOptions()
  await loadWarehouses()
})
</script>

<template>
  <div class="warehouse-container">
    <el-tabs v-model="activeTab" type="card" class="tabs">
      <el-tab-pane label="物资管理" name="items" />
      <el-tab-pane label="仓库管理" name="warehouses" />
      <el-tab-pane label="库存管理" name="inventory" />
      <el-tab-pane label="库存预警" name="alerts" />
    </el-tabs>

    <div v-if="activeTab === 'items'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input v-model="itemsQuery.query" placeholder="搜索物资名称" :prefix-icon="Search" clearable style="width: 320px" />
          <el-button :icon="RefreshCw" @click="loadItems">刷新</el-button>
        </div>
        <el-button v-if="canWarehouseManage" type="primary" :icon="Plus" @click="openCreateItem">新建物资</el-button>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="itemsData" v-loading="itemsLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="unit" label="单位" width="120" />
          <el-table-column prop="description" label="描述" min-width="200" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canWarehouseManage" link type="primary" :icon="Edit" @click="openEditItem(row)" />
              <el-button v-if="canAdminManage" link type="danger" :icon="Trash2" @click="deleteItem(row)" />
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <div class="pagination">
        <el-pagination
          background
          layout="prev, pager, next"
          :page-size="itemsQuery.size"
          :current-page="itemsQuery.page"
          :total="itemsTotal"
          @current-change="(p:number) => { itemsQuery.page = p; loadItems() }"
        />
      </div>
    </div>

    <div v-else-if="activeTab === 'warehouses'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="RefreshCw" @click="loadWarehouses">刷新</el-button>
        </div>
        <el-button v-if="canAdminManage" type="primary" :icon="Plus" @click="openCreateWarehouse">新建仓库</el-button>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="warehousesData" v-loading="warehousesLoading" style="width: 100%">
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="location" label="地址" min-width="180" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canAdminManage" link type="primary" :icon="Edit" @click="openEditWarehouse(row)" />
              <el-button v-if="canAdminManage" link type="danger" :icon="Trash2" @click="deleteWarehouse(row)" />
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <div v-else-if="activeTab === 'inventory'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select v-model="selectedWarehouseId" placeholder="选择仓库" style="width: 240px">
            <el-option v-for="w in warehousesData" :key="w.id" :label="`${w.name}（${w.location || '-'}）`" :value="w.id" />
          </el-select>
          <el-button :icon="RefreshCw" @click="loadInventory">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-button v-if="canWarehouseManage" type="success" :icon="PlusIcon" @click="openAdjust('add')">入库</el-button>
          <el-button v-if="canWarehouseManage" type="warning" :icon="Minus" @click="openAdjust('remove')">出库</el-button>
        </div>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="inventoryData" v-loading="inventoryLoading" style="width: 100%">
          <el-table-column prop="itemId" label="物资ID" width="110" />
          <el-table-column prop="name" label="物资名称" min-width="160" />
          <el-table-column prop="quantity" label="库存数量" width="120" />
        </el-table>
      </el-card>
    </div>

    <div v-else class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="RefreshCw" @click="loadAlerts">刷新</el-button>
        </div>
        <el-button v-if="canWarehouseManage" type="primary" @click="openSetThreshold">设置阈值</el-button>
      </div>

      <el-card class="table-card" shadow="never">
        <el-table :data="alertsData" v-loading="alertsLoading" style="width: 100%" :row-class-name="alertRowClassName">
          <el-table-column prop="itemId" label="物资ID" width="110" />
          <el-table-column prop="name" label="物资名称" min-width="160" />
          <el-table-column prop="quantity" label="当前库存" width="120">
            <template #default="{ row }">
              <div class="quantity-cell">
                <span>{{ row.quantity }}</span>
                <AlertTriangle :size="14" class="text-danger" />
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="threshold" label="阈值" width="120" />
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="itemDialogVisible" :title="itemDialogMode === 'create' ? '新建物资' : '编辑物资'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="itemForm.name" placeholder="请输入物资名称" />
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="itemForm.unit" placeholder="例如 箱/件/顶" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="itemForm.description" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="itemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveItem">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="warehouseDialogVisible" :title="warehouseDialogMode === 'create' ? '新建仓库' : '编辑仓库'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="warehouseForm.name" placeholder="请输入仓库名称" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="warehouseForm.location" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="warehouseDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveWarehouse">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="adjustDialogVisible" :title="adjustMode === 'add' ? '入库' : '出库'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="物资">
          <el-select v-model="adjustForm.itemId" placeholder="选择物资" filterable style="width: 100%">
            <el-option v-for="it in itemsOptions" :key="it.id" :label="`${it.name}（${it.unit}）`" :value="it.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数量">
          <el-input-number v-model="adjustForm.amount" :min="1" controls-position="right" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAdjust">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="thresholdDialogVisible" title="设置库存预警阈值" width="520px">
      <el-form label-width="90px">
        <el-form-item label="物资">
          <el-select v-model="thresholdForm.itemId" placeholder="选择物资" filterable style="width: 100%">
            <el-option v-for="it in itemsOptions" :key="it.id" :label="`${it.name}（${it.unit}）`" :value="it.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值">
          <el-input-number v-model="thresholdForm.threshold" :min="0" controls-position="right" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="thresholdDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveThreshold">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.warehouse-container {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.tabs :deep(.el-tabs__header) {
  margin: 0;
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

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.table-card {
  border: 1px solid #30363d;
  background-color: #0d1117;
}

.text-danger { color: #ff7b72; }

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

:deep(.el-table) {
  --el-table-header-bg-color: #161b22;
}

:deep(.el-tabs__item) {
  color: #8b949e;
}

:deep(.el-tabs__item.is-active) {
  color: #58a6ff;
}

:deep(.el-tabs__active-bar) {
  background-color: #58a6ff;
}

:deep(.el-table .warning-row) {
  --el-table-tr-bg-color: rgba(248, 81, 73, 0.1);
}

.quantity-cell {
  display: inline-flex;
  gap: 8px;
  align-items: center;
}
</style>
