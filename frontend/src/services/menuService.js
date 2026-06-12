import api from './api'

const USER_MENUS_CACHE_TTL = 5 * 60 * 1000

let userMenusCache = null
let userMenusCacheTime = 0
let userMenusRequest = null

const isUserMenusCacheFresh = () => (
  userMenusCache && Date.now() - userMenusCacheTime < USER_MENUS_CACHE_TTL
)

export const clearUserMenusCache = () => {
  userMenusCache = null
  userMenusCacheTime = 0
  userMenusRequest = null
}

const menuService = {
  getAll: (params = {}) => api.get('/menus', { params }),
  getUserMenus: ({ force = false } = {}) => {
    if (!force && isUserMenusCacheFresh()) return Promise.resolve(userMenusCache)
    if (!force && userMenusRequest) return userMenusRequest

    userMenusRequest = api.get('/menus/me')
      .then((response) => {
        userMenusCache = response
        userMenusCacheTime = Date.now()
        return response
      })
      .finally(() => {
        userMenusRequest = null
      })

    return userMenusRequest
  },
  getById: (id) => api.get(`/menu/${id}`),
  update: (id, payload) => api.put(`/menu/${id}`, payload).then((response) => {
    clearUserMenusCache()
    return response
  }),
}

export default menuService
