import request from './request'

/**
 * 认证接口：
 * - 登录成功返回 token + user（由前端存储到 Pinia/LocalStorage）
 * - 登出用于服务端清理（若后端实现了 logout）
 */
export const login = (data: any) => {
    return request.post('/api/v1/auth/login', data)
}

export const logout = () => {
    return request.post('/api/v1/auth/logout')
}
