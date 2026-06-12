import api from './api'

const MY_WALLET_CACHE_TTL = 30 * 1000

let myWalletCache = null
let myWalletCacheTime = 0
let myWalletRequest = null

const isMyWalletCacheFresh = () => (
  myWalletCache && Date.now() - myWalletCacheTime < MY_WALLET_CACHE_TTL
)

export const clearMyWalletCache = () => {
  myWalletCache = null
  myWalletCacheTime = 0
  myWalletRequest = null
}

const invalidateMyWalletAfter = (request) => request.then((response) => {
  clearMyWalletCache()
  return response
})

const walletService = {
  getMyWallet: ({ force = false } = {}) => {
    if (!force && isMyWalletCacheFresh()) return Promise.resolve(myWalletCache)
    if (!force && myWalletRequest) return myWalletRequest

    myWalletRequest = api.get('/wallet/me')
      .then((response) => {
        myWalletCache = response
        myWalletCacheTime = Date.now()
        return response
      })
      .finally(() => {
        myWalletRequest = null
      })

    return myWalletRequest
  },
  getMyTransactions: (params = {}) => api.get('/wallet/transactions', { params }),
  getDepositSettings: () => api.get('/deposits/settings'),
  createDeposit: (payload) => invalidateMyWalletAfter(api.post('/deposits', payload)),
  getMyDeposits: (params = {}) => api.get('/deposits', { params }),
  getMyDepositById: (id) => api.get(`/deposits/${id}`),
  getWallets: (params = {}) => api.get('/admin/wallets', { params }),
  adminTopup: (userId, payload) => invalidateMyWalletAfter(api.post(`/admin/wallets/${userId}/topup`, payload)),
  adminAdjust: (userId, payload) => invalidateMyWalletAfter(api.post(`/admin/wallets/${userId}/adjust`, payload)),
  getDeposits: (params = {}) => api.get('/admin/deposits', { params }),
  getDepositById: (id) => api.get(`/admin/deposits/${id}`),
  updateDepositStatus: (id, payload) => invalidateMyWalletAfter(api.post(`/admin/deposits/${id}/status`, payload)),
  clearMyWalletCache,
}

export default walletService
