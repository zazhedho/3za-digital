import api from './api'

const supportService = {
  getSupportContact: () => api.get('/support'),
}

export default supportService
