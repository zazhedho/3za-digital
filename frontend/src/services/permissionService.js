import api from './api'

const permissionService = {
  getAll: (params = {}) => api.get('/permissions', { params }),
  getUserPermissions: () => api.get('/permissions/me'),
}

export default permissionService
