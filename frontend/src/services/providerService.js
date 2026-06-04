import api from './api'

const providerService = {
  getBalance: () => api.get('/provider/h2h/balance'),
  getApiLogs: (params = {}) => api.get('/provider/api-logs', { params }),
}

export default providerService
