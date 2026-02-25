import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

/**
 * axios 实例封装：
 * - baseURL 使用 '/'：开发环境由 Vite proxy 转发到后端各服务
 * - 请求拦截：自动附带 JWT Bearer Token
 * - 响应拦截：统一提取 data、处理 401 过期与错误提示
 */
const request = axios.create({
    baseURL: '/',
    timeout: 10000
})

// 请求拦截：注入 Authorization 头
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

// 响应拦截：统一返回 data；错误时做全局提示与登录态处理
request.interceptors.response.use(
    (response) => {
        return response.data
    },
    (error) => {
        if (error.response) {
            const { status } = error.response
            if (status === 401) {
                // 认证过期/无效：清理登录态并回到登录页
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
