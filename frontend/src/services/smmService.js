import api from './api'

const smmService = {
  getServices: (params = {}) => api.get('/smm/services', { params }),
  syncServices: (payload = {}) => api.post('/smm/services/sync', payload),
  getOrders: (params = {}) => api.get('/smm/orders', { params }),
  getOrder: (id) => api.get(`/smm/orders/${id}`),
  createOrder: (payload) => api.post('/smm/orders', payload),
  refreshOrderStatus: (id) => api.post(`/smm/orders/${id}/refresh-status`),
  getOrderStatusLogs: (id) => api.get(`/smm/orders/${id}/status-logs`),
}

export default smmService
