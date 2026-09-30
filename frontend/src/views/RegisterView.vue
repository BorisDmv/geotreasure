<script setup>
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.register(username.value, email.value, password.value)
    router.push({ name: 'map' })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-card">
    <h1>Sign up</h1>
    <form @submit.prevent="submit">
      <label for="username">Username</label>
      <input id="username" v-model="username" required autocomplete="username" />

      <label for="email">Email</label>
      <input id="email" v-model="email" type="email" required autocomplete="email" />

      <label for="password">Password (min 8 chars)</label>
      <input
        id="password"
        v-model="password"
        type="password"
        required
        minlength="8"
        autocomplete="new-password"
      />

      <p v-if="error" class="error">{{ error }}</p>

      <button type="submit" :disabled="loading" style="margin-top: 1rem; width: 100%">
        {{ loading ? 'Creating account…' : 'Create account' }}
      </button>
    </form>
    <p style="margin-top: 1rem; color: var(--muted)">
      Already have an account? <RouterLink to="/login">Log in</RouterLink>
    </p>
  </div>
</template>
