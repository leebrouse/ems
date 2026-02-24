<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { 
  LayoutDashboard, 
  Package, 
  Truck, 
  Users, 
  Settings, 
  LogOut,
  Menu as MenuIcon,
  ChevronLeft,
  ChevronRight
} from 'lucide-vue-next'
import { ref } from 'vue'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()
const isCollapse = ref(false)

const menuItems = computed(() => {
  const items = [
    { title: 'Dashboard', index: '/dashboard', icon: LayoutDashboard },
    { title: 'Warehouse', index: '/warehouse', icon: Package, roles: ['Admin', 'WarehouseManager'] },
    { title: 'Scheduling', index: '/scheduling', icon: Truck, roles: ['Admin', 'Dispatcher'] },
    { title: 'Users', index: '/users', icon: Users, roles: ['Admin'] },
  ]
  
  return items.filter(item => {
    if (!item.roles) return true
    return item.roles.some(role => authStore.user?.roles.includes(role))
  })
})

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '240px'" class="aside">
      <div class="logo-container">
        <Truck class="logo-icon" />
        <span v-if="!isCollapse" class="logo-text">Rescue EMS</span>
      </div>
      
      <el-menu
        :default-active="route.path"
        class="el-menu-vertical"
        :collapse="isCollapse"
        router
        background-color="#0d1117"
        text-color="#8b949e"
        active-text-color="#58a6ff"
      >
        <el-menu-item v-for="item in menuItems" :key="item.index" :index="item.index">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>

      <div class="aside-footer">
        <el-button link @click="isCollapse = !isCollapse">
          <el-icon>
            <ChevronLeft v-if="!isCollapse" />
            <ChevronRight v-else />
          </el-icon>
        </el-button>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <span class="breadcrumb">{{ route.name }}</span>
        </div>
        <div class="header-right">
          <el-dropdown>
            <div class="user-profile">
              <el-avatar :size="32" class="avatar">{{ authStore.user?.username.charAt(0).toUpperCase() }}</el-avatar>
              <span class="username">{{ authStore.user?.username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>
                  <Settings :size="16" class="mr-2" /> Profile
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout" class="logout-item">
                  <LogOut :size="16" class="mr-2" /> Logout
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-container {
  height: 100vh;
}

.aside {
  background-color: #0d1117;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  overflow: hidden;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  gap: 12px;
  color: #58a6ff;
}

.logo-icon {
  width: 28px;
  height: 28px;
}

.logo-text {
  font-size: 1.2rem;
  font-weight: bold;
  white-space: nowrap;
}

.aside-footer {
  margin-top: auto;
  padding: 20px;
  display: flex;
  justify-content: center;
  border-top: 1px solid #30363d;
}

.header {
  background-color: #0d1117;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.header-left .breadcrumb {
  font-size: 1.1rem;
  font-weight: 500;
  color: #e6edf3;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.avatar {
  background-color: #58a6ff;
  color: white;
}

.username {
  color: #c9d1d9;
}

.main {
  background-color: #010409;
  padding: 24px;
}

.mr-2 {
  margin-right: 8px;
}

.logout-item {
  color: #f85149;
}

/* Animations */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
