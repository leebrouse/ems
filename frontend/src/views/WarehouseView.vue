<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus, Search, Filter, AlertTriangle, Edit, Trash2 } from 'lucide-vue-next'
import request from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

interface Item {
  id: number
  name: string
  unit: string
  category: string
  stock: number
}

const loading = ref(false)
const items = ref<Item[]>([])
const search = ref('')

const tableRowClassName = ({ row }: { row: Item }) => {
  if (row.stock < 10) {
    return 'warning-row'
  }
  return ''
}

const fetchItems = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/api/v1/items', { params: { page: 1, size: 100 } })
    items.value = res.items || []
  } catch (err) {
    // Mocking for demo if no backend items
    items.value = [
      { id: 101, name: 'Medical Kit A', unit: 'Box', category: 'Medical', stock: 5 },
      { id: 102, name: 'Bottled Water', unit: 'Case', category: 'Water', stock: 500 },
      { id: 103, name: 'Emergency Tent', unit: 'Piece', category: 'Shelter', stock: 2 },
      { id: 104, name: 'Dry Ration Pack', unit: 'Carton', category: 'Food', stock: 200 },
    ]
  } finally {
    loading.value = false
  }
}

const getStockStatus = (stock: number) => {
  if (stock < 10) return 'danger'
  if (stock < 50) return 'warning'
  return 'success'
}

onMounted(() => {
  fetchItems()
})
</script>

<template>
  <div class="warehouse-container">
    <div class="header-actions">
      <div class="search-box">
        <el-input
          v-model="search"
          placeholder="Search items..."
          :prefix-icon="Search"
          clearable
          class="custom-input"
        />
        <el-button :icon="Filter">Filters</el-button>
      </div>
      <el-button type="primary" :icon="Plus">Add New Item</el-button>
    </div>

    <el-card class="table-card" shadow="never">
      <el-table 
        :data="items" 
        style="width: 100%" 
        v-loading="loading"
        :row-class-name="tableRowClassName"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="Item Name" min-width="150" />
        <el-table-column prop="category" label="Category" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="stock" label="Stock" width="120">
          <template #default="{ row }">
            <div class="stock-cell">
              <span :class="['stock-value', `text-${getStockStatus(row.stock)}`]">{{ row.stock }}</span>
              <AlertTriangle v-if="row.stock < 10" :size="14" class="text-danger" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="Unit" width="100" />
        <el-table-column label="Actions" width="120" fixed="right">
          <template #default>
            <el-button link type="primary" :icon="Edit" />
            <el-button link type="danger" :icon="Trash2" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="pagination">
      <el-pagination background layout="prev, pager, next" :total="40" />
    </div>
  </div>
</template>

<style scoped>
.warehouse-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-box {
  display: flex;
  gap: 12px;
}

.custom-input {
  width: 300px;
}

.table-card {
  border: 1px solid #30363d;
  background-color: #0d1117;
}

.stock-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stock-value {
  font-weight: 600;
}

.text-danger { color: #ff7b72; }
.text-warning { color: #ffa657; }
.text-success { color: #7ee787; }

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

:deep(.el-table .warning-row) {
  --el-table-tr-bg-color: rgba(248, 81, 73, 0.1);
}
</style>
