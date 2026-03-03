import { reactive } from 'vue'

export const authStore = reactive({
  token: localStorage.getItem('token') || null,
  email: localStorage.getItem('email') || null,

  setAuth(token, email) {
    this.token = token
    this.email = email
    localStorage.setItem('token', token)
    localStorage.setItem('email', email)
  },

  logout() {
    this.token = null
    this.email = null
    localStorage.removeItem('token')
    localStorage.removeItem('email')
  },

  get isLoggedIn() {
    return !!this.token
  },
})
