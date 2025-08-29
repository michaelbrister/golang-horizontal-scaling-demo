# Golang Horizontal Scaling PoC (Traefik + Redis + Docker Compose)

This proof‑of‑concept demonstrates how a **stateful** web app can scale **horizontally** by externalizing state to **Redis (latest)** and load‑balancing multiple app instances with **Traefik**.

## What’s inside

- **Traefik (latest)** – reverse proxy / load balancer, auto‑discovers containers via the Docker provider.
- **Go web app** (built from Golang `latest` base image) – exposes a tiny API that:
  - sets a `sid` cookie,
  - increments a **per‑session** counter in Redis,
  - increments a **global** counter in Redis,
  - reports which **container** served your request.
- **Worker** – optional background consumer reading from a Redis list (`jobs`) to show async processing.
- **Redis (latest)** – shared state store.

```

make scale N=5   # Scale to 5 replicas
make scale N=2   # Scale down to 2 replicas
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

> 💡 **Apple Silicon users (M1/M2/M3 Macs):**
> The Dockerfile is configured with `GOARCH=amd64` so your Go app builds for Linux/amd64 inside Docker.
> This avoids cross-architecture issues during `go build`. No extra flags needed when running `docker compose up`.

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


## Troubleshooting

### Build fails at `go build`
- Ensure you’re on a recent Docker + Compose version.
- If you see `no required module provides package ...`, run `go mod tidy` locally inside `app/` and rebuild.
- On corporate networks, `proxy.golang.org` may be blocked. Try setting:
  ```bash
  export GOPROXY=direct
  docker compose build --no-cache
  ```

### Architecture issues on Apple Silicon (M1/M2/M3)
- This project forces `GOARCH=amd64` in the Dockerfile so Go builds target Linux/amd64, matching the rest of the stack.
- If you need to run **arm64** everywhere instead, update the Dockerfile and compose to use `arm64` images (e.g., `golang:latest` already supports arm64).

### Redis connection errors
- Ensure the `redis` container is healthy (`docker compose ps` shows `healthy`).
- If the app starts too quickly, rebuild with `--no-cache` to apply the `depends_on` health check.

### General tips
- Rebuild with full logs to debug:
  ```bash
  docker compose build --no-cache --progress=plain
  ```
- Check container logs:
  ```bash
  docker compose logs -f app
  docker compose logs -f worker
  ```


## Quickstart with Makefile

This project includes a simple **Makefile** to make common tasks easier:

```bash
make up       # Build and start the stack with 3 app replicas
make down     # Stop and remove containers + volumes
make build    # Rebuild images without cache
make logs     # Follow logs for all services
make ps       # List running containers and their status
```

> You can still run `docker compose` commands directly, but the Makefile provides handy shortcuts.


## Web Frontend

This demo now includes a small **Vue 3** UI (built with Vite and served by **nginx**) to visualize scaling:
- Traefik routes **`/api`** to the Go backend (prefix stripped), and routes **`/`** to the frontend.
- The dashboard polls `/api/` every ~1.5s to show which container served your request, plus session/global counters.
- You can enqueue jobs from the UI and watch the worker process them in logs.

**Open:** http://localhost:8080/ (frontend) — it calls `/api/*` behind Traefik.

### Build & run
```bash
make up
# or: docker compose up --build --scale app=3
```



### Testing sessions with curl

If you test the backend directly with `curl`, remember that `curl` does **not** keep cookies by default.  
Use a *cookie jar* file to persist the `sid` cookie between requests:

```bash
# First request saves cookies to cookies.txt
curl -c cookies.txt http://localhost:8080/api/

# Subsequent requests send cookies from cookies.txt
curl -b cookies.txt http://localhost:8080/api/
```

- `-c cookies.txt` → write response cookies into a file  
- `-b cookies.txt` → read cookies from that file and include them in the next request  

This way, you’ll see the **same `session_id`** and `session_count` increasing, just like in the web frontend.
