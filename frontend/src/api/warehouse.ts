import request from './request'

/**
 * 仓库相关接口：
 * - REST 路径由后端 warehouse 服务提供（/api/v1/warehouses）
 */
export interface Warehouse {
    id: number
    name: string
    location: string
}

export const listWarehouses = () => {
    return request.get<Warehouse[]>('/api/v1/warehouses')
}

export const getWarehouse = (id: number) => {
    return request.get<Warehouse>(`/api/v1/warehouses/${id}`)
}
