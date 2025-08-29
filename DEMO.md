# Demo Script: golang-horizontal-scaling-poc

This is a step-by-step guide to demo the project live. Total time ~10–12 minutes.

---

## 0) Prep (1 min)
```bash
make down || true
make up
```
Brings up Redis, Traefik, the Go web app scaled to 3 replicas, the worker, and the Vue UI.

Optional checks:
- Traefik dashboard: http://localhost:8081 (Routers → `frontend` and `api` green)
- App UI: http://localhost:8080/

---

## 1) Show the UI + state persistence (2 min)
Open http://localhost:8080/ and narrate:

- **Fields:**
  - **Served by** (container ID rotates)
  - **Session ID** (stays constant per browser session)
  - **Session count** (increments each refresh)
  - **Global count** (increments for all users)

- Change **Auto refresh** to 1.5s or 5s.
- Explain: “Session state is persisted in Redis, so scaling doesn’t break sessions.”

---

## 2) Prove horizontal scale (2 min)
Scale up and down live:

```bash
make scale N=5
```
→ “Added 2 more replicas.”

```bash
make scale N=2
```
→ “Scaled down, session persists.”

---

## 3) Background jobs (1–2 min)
Enqueue jobs via UI or curl:

```bash
curl -b cookies.txt -c cookies.txt "http://localhost:8080/api/enqueue?job=demo-1"
```

Watch worker logs:

```bash
docker compose logs -f worker
```

Explain: “Async jobs decouple long tasks.”

---

## 4) Failure & resiliency (1–2 min)
Kill one app container:

```bash
docker compose ps
docker compose stop golang-horizontal-scaling-poc-app-1
```

Explain: “Losing a node doesn’t break sessions.”

Restart it:

```bash
docker compose up -d
```

---

## 5) API vs Frontend routing (1 min)
- Frontend: http://localhost:8080/  
- API: calls to `/api/*` in devtools  
- Traefik dashboard: see requests to `frontend` vs `api`.

---

## 6) Command-line session persistence (optional 1 min)
Demonstrate cookie jar with curl:

```bash
curl -c cookies.txt http://localhost:8080/api/
curl -b cookies.txt http://localhost:8080/api/
```

---

## 7) Apple Silicon note (15 sec)
Mention that Dockerfile builds for `linux/amd64` so it works on M-series Macs.

---

## 8) Wrap-up (30 sec)
- Stateful apps scale when state is externalized.
- Web tier becomes stateless & elastic.
- Traefik cleanly routes and balances traffic.

---

## Troubleshooting
- **404 on `/`**: Check Traefik routers; priorities set (frontend=1, api=100).
- **Frontend changes not showing**: Rebuild frontend (`docker compose build frontend`).
- **Redis errors**: Ensure `REDIS_ADDR=redis:6379`.
- **Session resets**: Frontend has `credentials: 'include'`; for curl use `-c/-b`.

---

## Cleanup
```bash
make down
```
