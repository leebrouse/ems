import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

/**
 * 登录态 Store（Pinia）：
 * - token/user 会持久化到 LocalStorage，刷新页面后可恢复
 * - roles 用于路由与菜单的 RBAC 权限控制
 */
export interface User {
    id: number
    username: string
    roles: string[]
}

export const useAuthStore = defineStore('auth', () => {
    // 与 localStorage 的 key 对应：token / user
    const token = ref(localStorage.getItem('token') || '')
    const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

    const isAuthenticated = computed(() => !!token.value)
    const isAdmin = computed(() => user.value?.roles.includes('Admin'))
    const isWarehouseManager = computed(() => user.value?.roles.includes('WarehouseManager'))
    const isDispatcher = computed(() => user.value?.roles.includes('Dispatcher'))

    // 登录：更新内存态与持久化存储
    function setAuth(newToken: string, newUser: User) {
        token.value = newToken
        user.value = newUser
        localStorage.setItem('token', newToken)
        localStorage.setItem('user', JSON.stringify(newUser))
    }

    // 退出登录：清理内存态与持久化存储
    function logout() {
        token.value = ''
        user.value = null
        localStorage.removeItem('token')
        localStorage.removeItem('user')
    }

    return {
        token,
        user,
        isAuthenticated,
        isAdmin,
        isWarehouseManager,
        isDispatcher,
        setAuth,
        logout
    }
})
