import { defineStore } from 'pinia'
import { request } from '@/api/client'

export const useTreasureStore = defineStore('treasures', {
  state: () => ({
    items: [],
    loading: false,
    error: '',
  }),
  actions: {
    async fetchAll(query = '') {
      this.loading = true
      this.error = ''
      try {
        const q = query ? `?q=${encodeURIComponent(query)}` : ''
        this.items = await request(`/api/treasures${q}`)
      } catch (e) {
        this.error = e.message
      } finally {
        this.loading = false
      }
    },
    async bury({ name, description, hint, lat, lng }) {
      const created = await request('/api/treasures', {
        method: 'POST',
        body: { name, description, hint, lat, lng },
      })
      this.items.unshift(created)
      return created
    },
    async markFound(id, lat, lng) {
      await request(`/api/treasures/${id}/find`, {
        method: 'POST',
        body: { lat, lng },
      })
      const t = this.items.find((x) => x.id === id)
      if (t && !t.foundByMe) {
        t.foundByMe = true
        t.foundByCount += 1
      }
    },
    async remove(id) {
      await request(`/api/treasures/${id}`, { method: 'DELETE' })
      this.items = this.items.filter((x) => x.id !== id)
    },
  },
})
