import React, { useEffect, useState } from 'react'

const fetchJSON = async (path) => {
  const res = await fetch(path, { headers: { 'Accept': 'application/json' } })
  if (!res.ok) throw new Error('Request failed')
  return res.json()
}

export default function App() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [jobName, setJobName] = useState('demo-job')

  const load = async () => {
    try {
      setLoading(true)
      const json = await fetchJSON('/api/')
      setData(json)
      setError(null)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    const id = setInterval(load, 1500)
    return () => clearInterval(id)
  }, [])

  const enqueue = async () => {
    try {
      const json = await fetchJSON('/api/enqueue?job=' + encodeURIComponent(jobName))
      alert('Enqueued: ' + json.enqueued)
    } catch (e) {
      alert('Failed to enqueue: ' + e)
    }
  }

  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', margin: '2rem', maxWidth: 800 }}>
      <h1>Golang Horizontal Scaling Demo</h1>
      <p>This UI hits <code>/api</code> which Traefik routes to the Go app. Refreshes every 1.5s.</p>

      <section style={{ padding: '1rem', border: '1px solid #ccc', borderRadius: 8, marginTop: '1rem' }}>
        <h2>Status</h2>
        {loading && <p>Loading…</p>}
        {error && <p style={{ color: 'crimson' }}>Error: {error}</p>}
        {data && (
          <div style={{ lineHeight: 1.6 }}>
            <div><strong>Served by:</strong> {data.served_by}</div>
            <div><strong>Session ID:</strong> {data.session_id}</div>
            <div><strong>Session count:</strong> {data.session_count}</div>
            <div><strong>Global count:</strong> {data.global_count}</div>
          </div>
        )}
      </section>

      <section style={{ padding: '1rem', border: '1px solid #ccc', borderRadius: 8, marginTop: '1rem' }}>
        <h2>Enqueue a job</h2>
        <input
          value={jobName}
          onChange={(e) => setJobName(e.target.value)}
          placeholder="job name"
          style={{ padding: '0.5rem', marginRight: '0.5rem' }}
        />
        <button onClick={enqueue} style={{ padding: '0.5rem 1rem' }}>Enqueue</button>
        <p>View worker logs with <code>docker compose logs -f worker</code></p>
      </section>

      <section style={{ padding: '1rem', border: '1px solid #ccc', borderRadius: 8, marginTop: '1rem' }}>
        <h2>Scale replicas</h2>
        <p>Try <code>make scale N=5</code> and watch <strong>Served by</strong> change as Traefik rotates between containers while your session count keeps increasing.</p>
      </section>
    </div>
  )
}
