<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import TreasureMap from '@/components/TreasureMap.vue'
import { useTreasureStore } from '@/stores/treasures'

const store = useTreasureStore()
const router = useRouter()

const name = ref('')
const description = ref('')
const hint = ref('')
const picked = ref(null)
const center = ref({ lat: 20, lng: 0 })
const zoom = ref(3)
const error = ref('')
const locating = ref(false)
const saving = ref(false)

const canSave = computed(() => name.value.trim() && picked.value)

onMounted(useMyLocation)

function useMyLocation() {
  if (!navigator.geolocation) {
    error.value = 'Geolocation is not supported by this browser.'
    return
  }
  locating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const { latitude, longitude } = pos.coords
      picked.value = { lat: latitude, lng: longitude }
      center.value = { lat: latitude, lng: longitude }
      zoom.value = 16
      locating.value = false
    },
    (err) => {
      error.value = `Could not get your location: ${err.message}`
      locating.value = false
    },
    { enableHighAccuracy: true, timeout: 10000 },
  )
}

function onPick(latlng) {
  picked.value = { lat: latlng.lat, lng: latlng.lng }
}

async function save() {
  error.value = ''
  if (!canSave.value) return
  saving.value = true
  try {
    await store.bury({
      name: name.value.trim(),
      description: description.value.trim(),
      hint: hint.value.trim(),
      lat: picked.value.lat,
      lng: picked.value.lng,
    })
    router.push({ name: 'map' })
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="bury-layout">
    <aside class="panel">
      <h1>Bury a treasure</h1>
      <p style="color: var(--muted)">
        We use your current location by default. You can also click or drag the pin on the map to
        adjust exactly where you buried it.
      </p>

      <button class="secondary" :disabled="locating" @click="useMyLocation">
        {{ locating ? 'Locating…' : '📍 Use my current location' }}
      </button>

      <div v-if="picked" class="coords">
        📌 {{ picked.lat.toFixed(6) }}, {{ picked.lng.toFixed(6) }}
      </div>

      <label for="name">Name *</label>
      <input id="name" v-model="name" maxlength="120" placeholder="e.g. Grandpa's coin" />

      <label for="hint">Hint (shown publicly)</label>
      <input id="hint" v-model="hint" placeholder="e.g. Under the third oak" />

      <label for="desc">Description</label>
      <textarea id="desc" v-model="description" rows="3" placeholder="Notes about the treasure"></textarea>

      <p v-if="error" class="error">{{ error }}</p>

      <button style="margin-top: 1rem; width: 100%" :disabled="!canSave || saving" @click="save">
        {{ saving ? 'Burying…' : 'Bury it here' }}
      </button>
    </aside>

    <div class="map-wrap">
      <TreasureMap
        :treasures="[]"
        :center="center"
        :zoom="zoom"
        :pick-mode="true"
        :picked="picked"
        @pick="onPick"
      />
    </div>
  </div>
</template>

<style scoped>
.bury-layout {
  display: flex;
  height: 100%;
}
.panel {
  width: 360px;
  padding: 1rem 1.25rem;
  overflow-y: auto;
  border-right: 1px solid var(--border);
  background: var(--panel);
}
.panel h1 {
  margin-top: 0;
}
.map-wrap {
  flex: 1;
  min-width: 0;
}
.coords {
  margin-top: 0.75rem;
  padding: 0.5rem 0.7rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-family: monospace;
  font-size: 0.85rem;
}
@media (max-width: 720px) {
  .bury-layout {
    flex-direction: column;
  }
  .panel {
    width: 100%;
    max-height: 55%;
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
}
</style>
