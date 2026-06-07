import api from './api'

const authService = {
  login: (payload) => api.post('/user/login', payload),
  googleLogin: (idToken) => api.post('/user/google/login', { id_token: idToken }),
  register: (payload) => api.post('/user/register', payload),
  registerStatus: () => api.get('/user/register/status'),
  sendRegisterOTP: (email) => api.post('/user/register/otp/send', { email }),
  logout: () => api.post('/user/logout'),
  me: () => api.get('/user'),
  updateProfile: (payload) => api.put('/user', payload),
  changePassword: (payload) => api.put('/user/change/password', payload),
  forgotPassword: (email) => api.post('/user/forgot-password', { email }),
  resetPassword: (token, newPassword) => api.post('/user/reset-password', {
    token,
    new_password: newPassword,
  }),
}

export default authService
