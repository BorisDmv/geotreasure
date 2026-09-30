<script setup>
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push({ name: 'map' })
}
</script>

<template>
  <div class="app-shell">
    <nav class="navbar">
      <RouterLink to="/" class="brand">
        <span class="brand-icon">🗺️</span>
        <span class="brand-name">Geo Treasure</span>
      </RouterLink>
      <div class="nav-links">
        <RouterLink to="/" class="nav-link">Map</RouterLink>
        <RouterLink v-if="auth.isLoggedIn" to="/bury" class="nav-link">Bury treasure</RouterLink>
      </div>
      <span class="spacer"></span>
      <template v-if="auth.isLoggedIn">
        <span class="user-chip">
          <span class="user-avatar">{{ auth.user?.username?.[0] }}</span>
          <span class="user-name">{{ auth.user?.username }}</span>
        </span>
        <button class="ghost" @click="logout">Log out</button>
      </template>
      <template v-else>
        <RouterLink to="/login" class="nav-link">Log in</RouterLink>
        <RouterLink to="/register" class="nav-cta">Sign up</RouterLink>
      </template>
    </nav>

    <main class="content">
      <RouterView />
    </main>
  </div>
</template>
