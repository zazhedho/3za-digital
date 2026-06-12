import api from './api'
import { clearMyWalletCache } from './walletService'

const invalidateWalletAfter = (request) => request.then((response) => {
  clearMyWalletCache()
  return response
})

const smmService = {
  getServices: (params = {}) => api.get('/smm/services', { params }),
  syncServices: (payload = {}) => api.post('/smm/services/sync', payload),
  getOrders: (params = {}) => api.get('/smm/orders', { params }),
  getOrder: (id) => api.get(`/smm/orders/${id}`),
  createOrder: (payload) => invalidateWalletAfter(api.post('/smm/orders', payload)),
  refreshOrderStatus: (id) => api.post(`/smm/orders/${id}/refresh-status`),
  getOrderStatusLogs: (id) => api.get(`/smm/orders/${id}/status-logs`),
}

export default smmService
