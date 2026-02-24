import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
    baseURL: '/',
    timeout: 10000
})

// Request Interceptor
request.interceptors.request.use(
    (config) => {
        const authStore = useAuthStore()
        if (authStore.token) {
            config.headers.Authorization = `Bearer ${authStore.token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// Response Interceptor
request.interceptors.response.use(
    (response) => {
        return response.data
    },
    (error) => {
        if (error.response) {
            const { status } = error.response
            if (status === 401) {
                const authStore = useAuthStore()
                authStore.logout()
                router.push('/login')
                ElMessage.error('Session expired, please login again')
            } else {
                const message = error.response.data?.error || 'System error'
                ElMessage.error(message)
            }
        } else {
            ElMessage.error('Network error')
        }
        return Promise.reject(error)
    }
)

export default request
