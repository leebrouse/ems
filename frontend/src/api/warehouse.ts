import request from './request'

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
