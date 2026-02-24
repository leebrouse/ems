import request from './request'

export const login = (data: any) => {
    return request.post('/api/v1/auth/login', data)
}

export const logout = () => {
    return request.post('/api/v1/auth/logout')
}
