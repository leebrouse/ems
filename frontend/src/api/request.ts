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
                ElMessage.error('登录已过期，请重新登录')
            } else {
                const message = error.response.data?.error || '系统错误'
                ElMessage.error(message)
            }
        } else {
            ElMessage.error('网络错误')
        }
        return Promise.reject(error)
    }
)

export default request
