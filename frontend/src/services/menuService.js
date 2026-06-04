import api from './api'

const menuService = {
  getAll: (params = {}) => api.get('/menus', { params }),
  getUserMenus: () => api.get('/menus/me'),
  getById: (id) => api.get(`/menu/${id}`),
  update: (id, payload) => api.put(`/menu/${id}`, payload),
}

export default menuService
