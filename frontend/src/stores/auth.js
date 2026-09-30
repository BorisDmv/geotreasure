import { defineStore } from 'pinia'
import { request } from '@/api/client'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('gt_token') || '',
    user: JSON.parse(localStorage.getItem('gt_user') || 'null'),
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    persist() {
      if (this.token) {
        localStorage.setItem('gt_token', this.token)
        localStorage.setItem('gt_user', JSON.stringify(this.user))
      } else {
        localStorage.removeItem('gt_token')
        localStorage.removeItem('gt_user')
      }
    },
    async register(username, email, password) {
      const data = await request('/api/auth/register', {
        method: 'POST',
        auth: false,
        body: { username, email, password },
      })
      this.token = data.token
      this.user = data.user
      this.persist()
    },
    async login(email, password) {
      const data = await request('/api/auth/login', {
        method: 'POST',
        auth: false,
        body: { email, password },
      })
      this.token = data.token
      this.user = data.user
      this.persist()
    },
    logout() {
      this.token = ''
      this.user = null
      this.persist()
    },
  },
})
