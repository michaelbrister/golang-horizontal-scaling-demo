# Golang Horizontal Scaling PoC (Traefik + Redis + Docker Compose)

This proof‑of‑concept demonstrates how a **stateful** web app can scale **horizontally** by externalizing state to **Redis** and load‑balancing multiple app instances with **Traefik**.

## What’s inside

- **Traefik (v2.11)** – reverse proxy / load balancer, auto‑discovers containers via the Docker provider.
- **Go web app** – exposes a tiny API that:
  - sets a `sid` cookie,
  - increments a **per‑session** counter in Redis,
  - increments a **global** counter in Redis,
  - reports which **container** served your request.
- **Worker** – optional background consumer reading from a Redis list (`jobs`) to show async processing.
- **Redis** – shared state store.

```
.
├─ docker-compose.yml          # Traefik-based stack
├─ app/
│  ├─ Dockerfile
│  ├─ go.mod
│  └─ main.go
```

## Prereqs

- Docker Desktop or Docker Engine 24+
- Docker Compose V2

## Run (Traefik version)

Build and bring up the stack with **3 app replicas**:

```bash
docker compose up --build --scale app=3
```

Open the app:
- http://localhost:8080/

Now **refresh the page repeatedly**. You should see:
- `served_by` rotate across different container hostnames (Traefik load‑balancing).
- Your `session_id` stays constant and `session_count` **keeps increasing** (state is in Redis, not the container’s memory).
- `global_count` increments across all requests and users.

### Traefik Dashboard (optional)
A demo dashboard is enabled for convenience:
- http://localhost:8081

> ⚠️ Uses `--api.insecure=true`; **do not** use this flag in production.

## Scale elastically

Scale **up** and **down** live; your session will persist:
```bash
docker compose up -d --scale app=5
docker compose up -d --scale app=2
```

## Background jobs (optional)

Enqueue some jobs:
```bash
curl "http://localhost:8080/enqueue?job=test-1"
curl "http://localhost:8080/enqueue?job=test-2"
curl "http://localhost:8080/enqueue?job=test-3"
```

Watch the worker process them:
```bash
docker compose logs -f worker
```

## Shared key/value (optional)

```bash
curl "http://localhost:8080/set?key=color&value=blue"
curl "http://localhost:8080/get?key=color"
```

Any instance will return the same value because the data lives in Redis.

## Health checks (optional)

Traefik does passive health by default. To enable **active** checks against `/healthz`, add these labels to the `app` service in `docker-compose.yml`:

```yaml
labels:
  - "traefik.http.services.app.loadbalancer.healthcheck.path=/healthz"
  - "traefik.http.services.app.loadbalancer.healthcheck.interval=3s"
```

## Why this proves horizontal scaling for a “stateful” app

- **State externalized** (sessions & counters in Redis) → any node can serve any request.
- **Load-balanced replicas** (Traefik) → independent container instances scale out/in.
- **Async worker** → background processing scales separately from the web tier.

## Cleanup

```bash
docker compose down -v
```

This stops containers and removes the Redis volume created by this stack.
