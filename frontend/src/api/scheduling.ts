import request from './request'

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
