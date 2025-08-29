package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Role      string
	Redis     *redis.Client
	Hostname  string
	StartTime time.Time
}

func main() {
	role := env("ROLE", "web")
	addr := env("REDIS_ADDR", "redis:6379")
	db, _ := strconv.Atoi(env("REDIS_DB", "0"))

	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	host, _ := os.Hostname()
	app := &App{Role: role, Redis: rdb, Hostname: host, StartTime: time.Now()}

	switch role {
	case "worker":
		log.Printf("[worker %s] starting…", app.Hostname)
		app.runWorker()
	default:
		log.Printf("[web %s] starting…", app.Hostname)
		app.runWeb()
	}
}

func (a *App) runWeb() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sid := getOrSetSID(w, r)
		sessKey := fmt.Sprintf("sess:%s:count", sid)
		sessCount, _ := a.Redis.Incr(ctx, sessKey).Result()
		a.Redis.Expire(ctx, sessKey, 12*time.Hour)

		globalCount, _ := a.Redis.Incr(ctx, "global:count").Result()

		rand.Seed(time.Now().UnixNano() + int64(len(a.Hostname)))
		localNoise := rand.Intn(100000)

		resp := map[string]any{
			"served_by":     a.Hostname,
			"session_id":    sid,
			"session_count": sessCount,
			"global_count":  globalCount,
			"local_noise":   localNoise,
		}
		writeJSON(w, resp)
	})

	port := env("PORT", "8080")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (a *App) runWorker() {
	ctx := context.Background()
	for {
		res, err := a.Redis.BRPop(ctx, 0, "jobs").Result()
		if err == nil && len(res) == 2 {
			log.Printf("[worker %s] processed: %s", a.Hostname, res[1])
		}
	}
}

func getOrSetSID(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("sid")
	if err == nil && c.Value != "" {
		return c.Value
	}
	id := uuid.NewString()
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: id, Path: "/"})
	return id
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
