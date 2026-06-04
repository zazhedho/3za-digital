import api from './api'

const roleService = {
  getAll: (params = {}) => api.get('/roles', { params }),
  getById: (id) => api.get(`/role/${id}`),
  create: (payload) => api.post('/role', payload),
  update: (id, payload) => api.put(`/role/${id}`, payload),
  delete: (id) => api.delete(`/role/${id}`),
  assignPermissions: (id, payload) => api.post(`/role/${id}/permissions`, payload),
}

export default roleService
