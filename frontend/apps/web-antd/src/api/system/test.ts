import { requestClient } from '#/api/request';

// 测试2

export interface test {
  id: string;
  id: number;
  status: number;
  createdAt: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface testPageParams {
  currentPage?: number;
  pageSize?: number;
  status?: number;
}

export function listtestApi(params: testPageParams) {
  return requestClient.get('/admin/v1/tests', { params });
}

export function gettestApi(id: string) {
  return requestClient.get(`/admin/v1/tests/${id}`);
}

export function createtestApi(data: Partial<test>) {
  return requestClient.post('/admin/v1/tests', data);
}

export function updatetestApi(id: string, data: Partial<test>) {
  return requestClient.put(`/admin/v1/tests/${id}`, data);
}

export function updatetestStatusApi(id: string, status: string) {
  return requestClient.put(`/admin/v1/tests/${id}/status`, { status });
}

export function deletetestApi(ids: string[]) {
  return requestClient.delete('/admin/v1/tests', { data: { ids } });
}

export function importtestApi(data: FormData) {
  return requestClient.post('/admin/v1/tests/import', data, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
}

export function exporttestApi(params: testPageParams) {
  return requestClient.post('/admin/v1/tests/export', params, {
    responseType: 'blob',
    responseReturn: 'raw',
    showFailMessage: false,
  });
}
