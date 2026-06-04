import api from './api'

const appConfigService = {
  getAll: (params = {}) => api.get('/configs', { params }),
  getById: (id) => api.get(`/config/${id}`),
  update: (id, payload) => api.put(`/config/${id}`, payload),
}

export default appConfigService
