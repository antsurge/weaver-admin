import { requestClient } from '#/api/request';

// 订单

export interface Order {
  id: string;
  name: string;
  sn: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface OrderPageParams {
  currentPage?: number;
  pageSize?: number;
}

export function listOrderApi(params: OrderPageParams) {
  return requestClient.get('/admin/v1/orders', { params });
}

export function getOrderApi(id: string) {
  return requestClient.get(`/admin/v1/orders/${id}`);
}

export function createOrderApi(data: Partial<Order>) {
  return requestClient.post('/admin/v1/orders', data);
}

export function updateOrderApi(id: string, data: Partial<Order>) {
  return requestClient.put(`/admin/v1/orders/${id}`, data);
}

export function updateOrderStatusApi(id: string, status: string) {
  return requestClient.put(`/admin/v1/orders/${id}/status`, { status });
}

export function deleteOrderApi(ids: string[]) {
  return requestClient.delete('/admin/v1/orders', { data: { ids } });
}

export function importOrderApi(data: FormData) {
  return requestClient.post('/admin/v1/orders/import', data, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
}

export function exportOrderApi(params: OrderPageParams) {
  return requestClient.post('/admin/v1/orders/export', params, {
    responseType: 'blob',
    responseReturn: 'raw',
    showFailMessage: false,
  });
}
