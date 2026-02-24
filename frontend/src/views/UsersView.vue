<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Search, UserCheck, Shield, Trash2, Edit } from 'lucide-vue-next'
import request from '@/api/request'
import { ElMessage } from 'element-plus'

interface User {
  id: number
  username: string
  roles: string[]
}

const loading = ref(false)
const users = ref<User[]>([])

const fetchUsers = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/api/v1/users')
    users.value = res.users || []
  } catch (err) {
    // Mocking
    users.value = [
      { id: 1, username: 'admin', roles: ['Admin'] },
      { id: 2, username: 'manager1', roles: ['WarehouseManager'] },
      { id: 3, username: 'dispatcher1', roles: ['Dispatcher'] },
    ]
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchUsers()
})
</script>

<template>
  <div class="users-container">
    <div class="header-actions">
      <h2>User Management</h2>
      <el-button type="primary" :icon="Plus">Create User</el-button>
    </div>

    <el-card class="table-card" shadow="never">
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="Username" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="24" class="mr-2">{{ row.username[0].toUpperCase() }}</el-avatar>
              {{ row.username }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="roles" label="Roles" min-width="200">
          <template #default="{ row }">
            <div class="roles-tags">
              <el-tag 
                v-for="role in row.roles" 
                :key="role" 
                size="small" 
                :type="role === 'Admin' ? 'danger' : 'info'"
                effect="plain"
              >
                {{ role }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="150" fixed="right">
          <template #default>
            <el-button link type="primary" :icon="Edit" />
            <el-button link type="danger" :icon="Trash2" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.users-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #e6edf3;
}

.user-cell {
  display: flex;
  align-items: center;
}

.mr-2 { margin-right: 8px; }

.roles-tags {
  display: flex;
  gap: 4px;
}

.table-card {
  border: 1px solid #30363d;
  background-color: #0d1117;
}

:deep(.el-table) {
  --el-table-header-bg-color: #161b22;
}
</style>
