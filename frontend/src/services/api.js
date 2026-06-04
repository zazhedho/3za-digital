import axios from 'axios'

const API_BASE_URL = window.ENV_CONFIG?.API_URL || import.meta.env.VITE_API_URL || '/api'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.warn('Unauthorized request:', error.config?.url)
    }
    return Promise.reject(error)
  },
)

export const getErrorMessage = (error, fallback = 'Request failed') => {
  const payload = error?.response?.data
  if (payload?.error?.message) return payload.error.message
  if (payload?.message) return payload.message
  if (Array.isArray(payload?.error)) return payload.error.map((item) => item.message).join(', ')
  return fallback
}

export const getListPayload = (response) => {
  const body = response?.data || {}
  const rows = Array.isArray(body.data) ? body.data : []
  return {
    rows,
    total: body.total_data || body.total || body.meta?.total || rows.length,
    page: body.current_page || body.page || body.meta?.page || 1,
    limit: body.limit || body.meta?.limit || rows.length,
    totalPages: body.total_pages || body.meta?.total_pages || 1,
    nextPage: Boolean(body.next_page),
    prevPage: Boolean(body.prev_page),
  }
}

export default api
