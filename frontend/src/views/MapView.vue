<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import TreasureMap from '@/components/TreasureMap.vue'
import { useTreasureStore } from '@/stores/treasures'
import { useAuthStore } from '@/stores/auth'

const store = useTreasureStore()
const auth = useAuthStore()
const router = useRouter()

const search = ref('')
const center = ref({ lat: 20, lng: 0 })
const mapRef = ref(null)
const mapWrapEl = ref(null)

const userLocation = ref(null)
const locating = ref(false)
const liveTracking = ref(false)
const locateError = ref('')
let watchId = null

onMounted(() => store.fetchAll())
onBeforeUnmount(stopLive)

let searchTimer = null
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => store.fetchAll(search.value.trim()), 300)
}

function focus(t) {
  center.value = { lat: t.lat, lng: t.lng }
  mapRef.value?.flyTo(t.lat, t.lng, 15)
  // On phones the map may be scrolled out of view; bring it back.
  if (window.matchMedia('(max-width: 720px)').matches) {
    mapWrapEl.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

function applyPosition(pos, recenter) {
  const { latitude, longitude, accuracy } = pos.coords
  userLocation.value = { lat: latitude, lng: longitude, accuracy }
  locateError.value = ''
  if (recenter) {
    center.value = { lat: latitude, lng: longitude }
    mapRef.value?.flyTo(latitude, longitude, 16)
  }
}

// locateMe takes a single fresh high-accuracy reading and recenters the map.
function locateMe() {
  if (!navigator.geolocation) {
    locateError.value = 'Geolocation is not supported by this browser.'
    return
  }
  locating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      applyPosition(pos, true)
      locating.value = false
    },
    (err) => {
      locateError.value = `Could not get your location: ${err.message}`
      locating.value = false
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
  )
}

// toggleLive follows the device position continuously as you move.
function toggleLive() {
  if (liveTracking.value) {
    stopLive()
    return
  }
  if (!navigator.geolocation) {
    locateError.value = 'Geolocation is not supported by this browser.'
    return
  }
  liveTracking.value = true
  let first = true
  watchId = navigator.geolocation.watchPosition(
    (pos) => {
      applyPosition(pos, first)
      first = false
    },
    (err) => {
      locateError.value = `Location tracking failed: ${err.message}`
      stopLive()
    },
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 },
  )
}

function stopLive() {
  if (watchId != null) {
    navigator.geolocation.clearWatch(watchId)
    watchId = null
  }
  liveTracking.value = false
}

// currentPosition resolves a single fresh high-accuracy GPS reading.
function currentPosition() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('Geolocation is not supported by this browser.'))
      return
    }
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: 10000,
      maximumAge: 0,
    })
  })
}

async function markFound(t) {
  try {
    const pos = await currentPosition()
    applyPosition(pos, false)
    const { latitude, longitude } = pos.coords
    await store.markFound(t.id, latitude, longitude)
  } catch (e) {
    alert(e.message)
  }
}

async function remove(t) {
  if (!confirm(`Delete "${t.name}"?`)) return
  try {
    await store.remove(t.id)
  } catch (e) {
    alert(e.message)
  }
}

function fmtDistance(m) {
  if (m == null) return ''
  return m < 1000 ? `${Math.round(m)} m` : `${(m / 1000).toFixed(1)} km`
}
</script>

<template>
  <div class="map-layout">
    <aside class="sidebar">
      <section class="panel controls-panel">
        <div class="sidebar-head">
          <input v-model="search" placeholder="Search treasures or users…" @input="onSearch" />
          <button class="secondary icon-btn" title="Find my location" :disabled="locating" @click="locateMe">
            {{ locating ? '…' : '📍' }}
          </button>
        </div>

        <div class="locate-row">
          <button class="secondary locate-btn" :class="{ active: liveTracking }" @click="toggleLive">
            {{ liveTracking ? '🔴 Live: on' : '🛰️ Live tracking' }}
          </button>
          <span v-if="userLocation?.accuracy" class="accuracy">
            ±{{ Math.round(userLocation.accuracy) }} m
          </span>
        </div>
        <div v-if="userLocation" class="my-coords">
          📍 You: {{ userLocation.lat.toFixed(6) }}, {{ userLocation.lng.toFixed(6) }}
        </div>
        <p v-if="locateError" class="error">{{ locateError }}</p>

        <button
          v-if="auth.isLoggedIn"
          class="bury-btn"
          @click="router.push({ name: 'bury' })"
        >
          + Bury a treasure here
        </button>
      </section>

      <section class="panel list-panel">
        <p v-if="store.loading" class="muted-note">Loading…</p>
        <p v-if="store.error" class="error">{{ store.error }}</p>
        <p v-if="!store.loading && !store.items.length" class="muted-note">
          No treasures yet.
        </p>

        <ul class="treasure-list">
          <li v-for="t in store.items" :key="t.id" class="treasure">
            <div class="treasure-main" @click="focus(t)">
              <div class="treasure-name">
                {{ t.name }}
                <span v-if="t.foundByMe" title="You found this">✅</span>
              </div>
              <div class="treasure-meta">
                by {{ t.ownerName }} · {{ t.foundByCount }} found
                <span v-if="t.distanceMeters != null"> · {{ fmtDistance(t.distanceMeters) }}</span>
              </div>
              <div class="treasure-coords">📌 {{ t.lat.toFixed(6) }}, {{ t.lng.toFixed(6) }}</div>
              <div v-if="t.hint" class="treasure-hint">Hint: {{ t.hint }}</div>
            </div>
            <div class="treasure-actions" v-if="auth.isLoggedIn">
              <button
                v-if="!t.foundByMe && t.ownerId !== auth.user?.id"
                class="secondary"
                @click="markFound(t)"
              >
                I found it
              </button>
              <button v-if="t.ownerId === auth.user?.id" class="danger" @click="remove(t)">
                Delete
              </button>
            </div>
          </li>
        </ul>
      </section>
    </aside>

    <div class="map-wrap" ref="mapWrapEl">
      <TreasureMap
        ref="mapRef"
        :treasures="store.items"
        :center="center"
        :user-location="userLocation"
        @select="focus"
      />
    </div>
  </div>
</template>

<style scoped>
.map-layout {
  display: flex;
  gap: 1rem;
  height: 100%;
  padding: 1rem;
}
.sidebar {
  width: 360px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-height: 0;
}
.panel {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 1rem;
}
.controls-panel {
  flex-shrink: 0;
}
.list-panel {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
.sidebar-head {
  display: flex;
  gap: 0.5rem;
}
.icon-btn {
  flex-shrink: 0;
}
.locate-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
.locate-btn {
  flex: 1;
}
.locate-btn.active {
  background: var(--info);
  color: #fff;
}
.accuracy {
  font-size: 0.8rem;
  color: var(--muted);
  white-space: nowrap;
}
.my-coords {
  margin-top: 0.75rem;
  padding: 0.4rem 0.6rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  font-family: monospace;
  font-size: 0.8rem;
  color: var(--info-soft);
}
.bury-btn {
  width: 100%;
  margin-top: 0.75rem;
}
.muted-note {
  color: var(--muted);
  margin: 0.25rem 0;
}
.treasure-coords {
  font-family: monospace;
  font-size: 0.75rem;
  color: var(--muted);
  margin-top: 0.15rem;
}
.map-wrap {
  flex: 1;
  min-width: 0;
  border-radius: 18px;
  overflow: hidden;
  border: 1px solid var(--border);
}
.map-wrap :deep(.leaflet-container) {
  height: 100%;
  width: 100%;
}
.treasure-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.treasure {
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--border);
}
.treasure:first-child {
  padding-top: 0;
}
.treasure:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
.treasure-main {
  cursor: pointer;
}
.treasure-name {
  font-weight: 600;
}
.treasure-meta {
  font-size: 0.8rem;
  color: var(--muted);
  margin-top: 0.15rem;
}
.treasure-hint {
  font-size: 0.8rem;
  color: var(--accent);
  margin-top: 0.15rem;
}
.treasure-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}
.treasure-actions button {
  padding: 0.35rem 0.6rem;
  font-size: 0.8rem;
}
@media (max-width: 720px) {
  .map-layout {
    flex-direction: column;
    padding: 0.75rem;
    gap: 0.75rem;
    overflow-y: auto;
  }
  /* Promote the panels to layout items so the map can sit between controls and list. */
  .sidebar {
    display: contents;
  }
  .controls-panel {
    order: 1;
  }
  .map-wrap {
    order: 2;
    flex: none;
    height: 52vh;
    min-height: 300px;
    scroll-margin-top: 0.75rem;
  }
  .list-panel {
    order: 3;
    flex: none;
    max-height: none;
    overflow-y: visible;
  }
}
</style>
