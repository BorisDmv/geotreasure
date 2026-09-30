<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import L from 'leaflet'
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'

// Vite bundles images as URLs, so point Leaflet's default icon at them.
L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
})

const props = defineProps({
  treasures: { type: Array, default: () => [] },
  center: { type: Object, default: () => ({ lat: 20, lng: 0 }) },
  zoom: { type: Number, default: 3 },
  pickMode: { type: Boolean, default: false },
  picked: { type: Object, default: null },
  userLocation: { type: Object, default: null },
})

const emit = defineEmits(['pick', 'select'])

const mapEl = ref(null)
let map = null
let markerLayer = null
let pickMarker = null
let userDot = null
let userAccuracy = null

const goldIcon = new L.Icon({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
})

// buryIcon is a self-contained SVG pin (no images or emoji glyphs), so it renders
// crisply and identically in every browser: gradient-gold teardrop, X marks the spot.
const buryIcon = L.divIcon({
  className: 'bury-pin',
  html: `
    <svg width="36" height="48" viewBox="0 0 36 48" xmlns="http://www.w3.org/2000/svg"
         style="filter: drop-shadow(0 2px 2px rgba(10, 16, 31, 0.45))">
      <defs>
        <linearGradient id="bury-pin-grad" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stop-color="#fcd34d"/>
          <stop offset="1" stop-color="#f59e0b"/>
        </linearGradient>
      </defs>
      <path d="M18 1.5C9.2 1.5 2 8.7 2 17.5c0 11.6 13.2 26.7 14.8 28.5a1.6 1.6 0 0 0 2.4 0C20.8 44.2 34 29.1 34 17.5 34 8.7 26.8 1.5 18 1.5z"
            fill="url(#bury-pin-grad)" stroke="#92400e" stroke-width="1.5"/>
      <circle cx="18" cy="17.5" r="9" fill="#fff" fill-opacity="0.95"/>
      <path d="M14 13.5l8 8M22 13.5l-8 8" stroke="#b45309" stroke-width="2.8" stroke-linecap="round"/>
    </svg>`,
  iconSize: [36, 48],
  iconAnchor: [18, 48],
  popupAnchor: [0, -44],
})

function renderTreasures() {
  if (!markerLayer) return
  markerLayer.clearLayers()
  for (const t of props.treasures) {
    const found = t.foundByMe ? ' ✅' : ''
    const marker = L.marker([t.lat, t.lng], { icon: goldIcon })
      .bindPopup(
        `<strong>${escapeHtml(t.name)}${found}</strong><br/>` +
          `by ${escapeHtml(t.ownerName)}<br/>` +
          (t.hint ? `<em>Hint: ${escapeHtml(t.hint)}</em><br/>` : '') +
          `📌 ${t.lat.toFixed(6)}, ${t.lng.toFixed(6)}<br/>` +
          `Found by ${t.foundByCount} hunter(s)`,
      )
    marker.on('click', () => emit('select', t))
    marker.addTo(markerLayer)
  }
}

function escapeHtml(s) {
  return String(s ?? '').replace(
    /[&<>"']/g,
    (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c],
  )
}

function setPickMarker(latlng) {
  if (!map) return
  if (pickMarker) {
    pickMarker.setLatLng(latlng)
  } else {
    pickMarker = L.marker(latlng, { draggable: true, icon: buryIcon })
      .addTo(map)
      .bindPopup('Treasure spot — drag to adjust')
    pickMarker.on('dragend', () => emit('pick', pickMarker.getLatLng()))
  }
}

// renderUser draws the blue "you are here" dot and a translucent circle showing
// the GPS accuracy radius, updating in place as the location changes.
function renderUser() {
  if (!map) return
  const loc = props.userLocation
  if (!loc) {
    if (userDot) {
      map.removeLayer(userDot)
      userDot = null
    }
    if (userAccuracy) {
      map.removeLayer(userAccuracy)
      userAccuracy = null
    }
    return
  }

  const latlng = [loc.lat, loc.lng]
  const accuracy = loc.accuracy || 0
  const popupHtml =
    `<strong>You are here</strong><br/>📌 ${loc.lat.toFixed(6)}, ${loc.lng.toFixed(6)}` +
    (accuracy ? `<br/>±${Math.round(accuracy)} m` : '')

  if (!userDot) {
    userAccuracy = L.circle(latlng, {
      radius: accuracy,
      color: '#2563eb',
      weight: 1,
      fillColor: '#3b82f6',
      fillOpacity: 0.15,
    }).addTo(map)
    userDot = L.circleMarker(latlng, {
      radius: 8,
      color: '#ffffff',
      weight: 2,
      fillColor: '#2563eb',
      fillOpacity: 1,
    })
      .addTo(map)
      .bindPopup(popupHtml)
  } else {
    userDot.setLatLng(latlng)
    userDot.setPopupContent(popupHtml)
    userAccuracy.setLatLng(latlng)
    userAccuracy.setRadius(accuracy)
  }
}

onMounted(() => {
  map = L.map(mapEl.value).setView([props.center.lat, props.center.lng], props.zoom)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  markerLayer = L.layerGroup().addTo(map)
  renderTreasures()
  renderUser()

  if (props.pickMode) {
    map.on('click', (e) => {
      setPickMarker(e.latlng)
      emit('pick', e.latlng)
    })
    if (props.picked) setPickMarker(props.picked)
  }
})

onBeforeUnmount(() => {
  if (map) map.remove()
})

watch(() => props.treasures, renderTreasures, { deep: true })

watch(
  () => props.center,
  (c) => {
    if (map && c) map.setView([c.lat, c.lng], Math.max(map.getZoom(), 13))
  },
)

watch(
  () => props.picked,
  (p) => {
    if (props.pickMode && p) setPickMarker(p)
  },
)

watch(() => props.userLocation, renderUser, { deep: true })

defineExpose({
  flyTo(lat, lng, zoom = 15) {
    if (map) map.flyTo([lat, lng], zoom)
  },
})
</script>

<template>
  <div ref="mapEl" class="map"></div>
</template>

<style scoped>
.map {
  width: 100%;
  height: 100%;
}
</style>
