import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
    id: number
    username: string
    roles: string[]
}

export const useAuthStore = defineStore('auth', () => {
    const token = ref(localStorage.getItem('token') || '')
    const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

    const isAuthenticated = computed(() => !!token.value)
    const isAdmin = computed(() => user.value?.roles.includes('Admin'))
    const isWarehouseManager = computed(() => user.value?.roles.includes('WarehouseManager'))
    const isDispatcher = computed(() => user.value?.roles.includes('Dispatcher'))

    function setAuth(newToken: string, newUser: User) {
        token.value = newToken
        user.value = newUser
        localStorage.setItem('token', newToken)
        localStorage.setItem('user', JSON.stringify(newUser))
    }

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
