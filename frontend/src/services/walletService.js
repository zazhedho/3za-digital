import api from './api'

const walletService = {
  getMyWallet: () => api.get('/wallet/me'),
  getMyTransactions: (params = {}) => api.get('/wallet/transactions', { params }),
  getDepositSettings: () => api.get('/deposits/settings'),
  createDeposit: (payload) => api.post('/deposits', payload),
  getMyDeposits: (params = {}) => api.get('/deposits', { params }),
  getMyDepositById: (id) => api.get(`/deposits/${id}`),
  getWallets: (params = {}) => api.get('/admin/wallets', { params }),
  adminTopup: (userId, payload) => api.post(`/admin/wallets/${userId}/topup`, payload),
  adminAdjust: (userId, payload) => api.post(`/admin/wallets/${userId}/adjust`, payload),
  getDeposits: (params = {}) => api.get('/admin/deposits', { params }),
  getDepositById: (id) => api.get(`/admin/deposits/${id}`),
  updateDepositStatus: (id, payload) => api.post(`/admin/deposits/${id}/status`, payload),
}

export default walletService
