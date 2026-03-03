import { api } from './client.js'

export const authApi = {
  login: (email, password) => api.post('/auth/login', { email, password }),
  register: (name, email, password) => api.post('/auth/register', { name, email, password }),
  createAccount: () => api.post('/account', {}),
}
