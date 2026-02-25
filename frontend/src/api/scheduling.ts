import request from './request'

/**
 * 调度/运输相关接口：
 * - Shipment 表示运输任务（由 scheduling 服务提供）
 * - 列表接口支持分页与状态筛选
 */
export interface Shipment {
    id: number
    requestId: number
    fromWarehouseId: number
    toLocation: string
    status: string
    items?: any[]
    tracking?: any[]
    createdAt?: string
    updatedAt?: string
}

export const listShipments = (params?: { page?: number; size?: number; status?: string }) => {
    return request.get<{ shipments: Shipment[]; total: number }>('/api/v1/shipments', { params })
}

export const getShipment = (id: number) => {
    return request.get<Shipment>(`/api/v1/shipments/${id}`)
}

export const updateShipmentStatus = (id: number, data: { status: string; location?: string; timestamp?: string }) => {
    return request.put(`/api/v1/shipments/${id}/status`, data)
}
