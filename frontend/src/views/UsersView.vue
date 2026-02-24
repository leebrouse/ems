<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Plus, Trash2, Edit } from 'lucide-vue-next'
import request from '@/api/request'
import { ElMessage, ElMessageBox } from 'element-plus'

interface User {
  id: number
  username: string
  roles: string[]
}

interface Role {
  name: string
  description?: string
}

const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const query = reactive({ page: 1, size: 20 })

const roles = ref<Role[]>([])
const roleNameToLabel: Record<string, string> = {
  Admin: '系统管理员',
  WarehouseManager: '仓库管理员',
  Dispatcher: '调度员',
}
const getRoleLabel = (role: string) => roleNameToLabel[role] || role

const loadRoles = async () => {
  const res: any = await request.get('/api/v1/roles')
  roles.value = res ?? []
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/api/v1/users', { params: query })
    users.value = res?.users ?? []
    total.value = Number(res?.total) || 0
  } finally {
    loading.value = false
  }
}

const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const form = reactive({
  id: 0,
  username: '',
  password: '',
  roles: [] as string[],
})

const openCreate = () => {
  dialogMode.value = 'create'
  form.id = 0
  form.username = ''
  form.password = ''
  form.roles = []
  dialogVisible.value = true
}

const openEdit = (row: User) => {
  dialogMode.value = 'edit'
  form.id = row.id
  form.username = row.username
  form.password = ''
  form.roles = [...(row.roles ?? [])]
  dialogVisible.value = true
}

const save = async () => {
  if (dialogMode.value === 'create') {
    if (!form.username || !form.password) {
      ElMessage.warning('请输入用户名和密码')
      return
    }
    await request.post('/api/v1/users', { username: form.username, password: form.password, roles: form.roles })
    ElMessage.success('用户已创建')
  } else {
    const payload: any = {}
    if (form.password) payload.password = form.password
    if (form.roles) payload.roles = form.roles
    await request.put(`/api/v1/users/${form.id}`, payload)
    ElMessage.success('用户已更新')
  }
  dialogVisible.value = false
  await fetchUsers()
}

const removeUser = async (row: User) => {
  await ElMessageBox.confirm(`确认删除用户 ${row.username}（ID: ${row.id}）？`, '提示', { type: 'warning' })
  await request.delete(`/api/v1/users/${row.id}`)
  ElMessage.success('已删除')
  await fetchUsers()
}

onMounted(() => {
  loadRoles()
  fetchUsers()
})
</script>

<template>
  <div class="users-container">
    <div class="header-actions">
      <h2>用户管理</h2>
      <el-button type="primary" :icon="Plus" @click="openCreate">新建用户</el-button>
    </div>

    <el-card class="table-card" shadow="never">
      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="24" class="mr-2">{{ row.username[0].toUpperCase() }}</el-avatar>
              {{ row.username }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="roles" label="角色" min-width="240">
          <template #default="{ row }">
            <div class="roles-tags">
              <el-tag 
                v-for="role in row.roles" 
                :key="role" 
                size="small" 
                :type="role === 'Admin' ? 'danger' : 'info'"
                effect="plain"
              >
                {{ getRoleLabel(role) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Edit" @click="openEdit(row)" />
            <el-button link type="danger" :icon="Trash2" @click="removeUser(row)" />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="pagination">
      <el-pagination
        background
        layout="prev, pager, next"
        :page-size="query.size"
        :current-page="query.page"
        :total="total"
        @current-change="(p:number) => { query.page = p; fetchUsers() }"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新建用户' : '编辑用户'" width="520px">
      <el-form label-width="90px">
        <el-form-item label="用户名">
          <el-input v-model="form.username" :disabled="dialogMode === 'edit'" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item :label="dialogMode === 'create' ? '密码' : '新密码'">
          <el-input v-model="form.password" type="password" show-password :placeholder="dialogMode === 'create' ? '请输入密码' : '不修改可留空'" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.roles" multiple filterable placeholder="选择角色" style="width: 100%">
            <el-option v-for="r in roles" :key="r.name" :label="`${getRoleLabel(r.name)}（${r.name}）`" :value="r.name" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
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

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

:deep(.el-table) {
  --el-table-header-bg-color: #161b22;
}
</style>
