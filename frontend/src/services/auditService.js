import api from './api'

const auditService = {
  getAll: (params = {}) => api.get('/audits', { params }),
  getById: (id) => api.get(`/audit/${id}`),
}

export default auditService
