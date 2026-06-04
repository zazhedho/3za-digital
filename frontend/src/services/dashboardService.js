import api from './api'

const dashboardService = {
  getSummary: (productType = 'smm') => api.get('/dashboard/summary', {
    params: { product_type: productType },
  }),
}

export default dashboardService
