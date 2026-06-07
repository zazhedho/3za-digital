import api from './api'

const sessionService = {
  getActiveSessions: () => api.get('/user/sessions'),
  revokeSession: (sessionId) => api.delete(`/user/session/${sessionId}`),
  revokeAllOtherSessions: () => api.post('/user/sessions/revoke-others'),
}

export default sessionService
