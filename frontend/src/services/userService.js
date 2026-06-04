import api from './api'

const userService = {
  getAll: (params = {}) => api.get('/users', { params }),
  create: (payload) => api.post('/user', payload),
  getById: (id) => api.get(`/user/${id}`),
  update: (id, payload) => api.put(`/user/${id}`, payload),
  delete: (id) => api.delete(`/user/${id}`),
}

export default userService
