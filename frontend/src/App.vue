<template>
  <div class="container">
    <h1>Golang Horizontal Scaling Demo (Vue)</h1>
    <p>This UI hits <code>/api</code> via Traefik. Refreshes every ~1.5s.</p>

    <section class="card">
      <h2>Status</h2>
      <p v-if="loading">Loading…</p>
      <p v-else-if="error" style="color:crimson">Error: {{ error }}</p>
      <div v-else>
        <div><strong>Served by:</strong> {{ data.served_by }}</div>
        <div><strong>Session ID:</strong> {{ data.session_id }}</div>
        <div><strong>Session count:</strong> {{ data.session_count }}</div>
        <div><strong>Global count:</strong> {{ data.global_count }}</div>
      </div>
    </section>

    <section class="card">
      <h2>Enqueue a job</h2>
      <input v-model="jobName" placeholder="job name" />
      <button @click="enqueue">Enqueue</button>
      <p>View worker logs with <code>docker compose logs -f worker</code></p>
    </section>

    <section class="card">
      <h2>Scale replicas</h2>
      <p>Run <code>make scale N=5</code> and watch <strong>Served by</strong> change while your session count keeps increasing.</p>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const data = ref(null)
const loading = ref(true)
const error = ref(null)
const jobName = ref('demo-job')

const load = async () => {
  try {
    loading.value = true
    const res = await fetch('/api/')
    if (!res.ok) throw new Error('Request failed')
    data.value = await res.json()
    error.value = null
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

const enqueue = async () => {
  try {
    const res = await fetch('/api/enqueue?job=' + encodeURIComponent(jobName.value))
    const json = await res.json()
    alert('Enqueued: ' + json.enqueued)
  } catch (e) {
    alert('Failed to enqueue: ' + e.message)
  }
}

onMounted(() => {
  load()
  setInterval(load, 5000)
})
</script>

<style>
.container { font-family: system-ui, sans-serif; margin: 2rem; max-width: 800px; }
.card { padding: 1rem; border: 1px solid #ccc; border-radius: 8px; margin-top: 1rem; }
input { padding: 0.5rem; margin-right: 0.5rem; }
button { padding: 0.5rem 1rem; }
</style>
