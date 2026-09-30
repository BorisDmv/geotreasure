<script setup>
import { ref } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push(route.query.redirect || { name: 'map' })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-card">
    <h1>Log in</h1>
    <form @submit.prevent="submit">
      <label for="email">Email</label>
      <input id="email" v-model="email" type="email" required autocomplete="email" />

      <label for="password">Password</label>
      <input
        id="password"
        v-model="password"
        type="password"
        required
        autocomplete="current-password"
      />

      <p v-if="error" class="error">{{ error }}</p>

      <button type="submit" :disabled="loading" style="margin-top: 1rem; width: 100%">
        {{ loading ? 'Logging in…' : 'Log in' }}
      </button>
    </form>
    <p style="margin-top: 1rem; color: var(--muted)">
      No account? <RouterLink to="/register">Sign up</RouterLink>
    </p>
  </div>
</template>
