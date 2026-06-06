# job-queue

A background job processing system written in Go. Accepts jobs over HTTP, queues them in Redis, and processes them asynchronously with automatic retries, exponential backoff, and a dead letter queue.

Follows the pattern behind Sidekiq, BullMQ, and Faktory.

---

## How it works

```
POST /jobs
     |
     v
server stores job in SQLite (status: pending)
     |
     v
job ID pushed to Redis queue (LPUSH)
     |
     v
worker blocks on Redis (BRPOP) — wakes instantly when job arrives
     |
     v
job processed in a goroutine
     |
     +-- success --> status: completed
     |
     +-- failure --> exponential backoff (2s, 4s, 8s) --> retry
                          |
                          v
                   3 attempts exhausted
                          |
                          v
                   status: failed (dead letter queue)
                          |
                          v
                   POST /jobs/:id/retry --> requeued
```

Server and worker are separate binaries. They share the same SQLite database via a mounted volume. Redis is the signal layer — only job IDs are queued, full payloads live in SQLite.

---

## Job types

| Type     | Behavior                                      |
|----------|-----------------------------------------------|
| `email`  | Simulates sending an email (500ms)            |
| `report` | Simulates generating a report (1s)            |
| `fail`   | Always fails — for testing DLQ and retry      |

Adding a new job type is one case in the `handle()` switch in `cmd/worker/main.go`.

---

## API

### Submit a job

```
POST /jobs
Content-Type: application/json

{
  "type": "email",
  "payload": {"to": "user@example.com", "subject": "welcome"}
}
```

Returns `202 Accepted` with the job ID. Processing is async.

### Check job status

```
GET /jobs/:id
```

```json
{
  "ID": "f1db57b5-...",
  "Type": "email",
  "Status": "completed",
  "Attempts": 1,
  "MaxAttempts": 3,
  "Error": ""
}
```

### List all jobs

```
GET /jobs
```

### Retry a failed job

```
POST /jobs/:id/retry
```

Resets attempts to 0, pushes back to Redis. Worker picks it up within milliseconds.

---

## Running locally

Set environment variables:

```bash
export REDIS_URL=rediss://...   # Upstash or any Redis instance
export DB_PATH=jobs.db
```

Start server and worker in separate terminals:

```bash
go run cmd/server/main.go
go run cmd/worker/main.go
```

Test the full cycle:

```bash
# successful job
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"email","payload":{"to":"x@example.com"}}'

# job that hits DLQ after 3 attempts
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"fail","payload":{}}'

# check status
curl http://localhost:8080/jobs/:id

# replay from DLQ
curl -X POST http://localhost:8080/jobs/:id/retry
```

---

## Docker

```bash
docker-compose up
```

Starts the server on `:8080` and the worker as a separate container. Both share a volume for SQLite. Pass your Redis URL via `.env`:

```
REDIS_URL=rediss://...
```

---

## Configuration

| Variable    | Default    | Description                        |
|-------------|------------|------------------------------------|
| `REDIS_URL` | required   | Redis connection string            |
| `DB_PATH`   | `jobs.db`  | Path to SQLite database file       |

---

## Stack

- Go 1.26
- Redis via [go-redis](https://github.com/redis/go-redis) — LPUSH/BRPOP queue
- SQLite via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure Go, no CGO
- Standard library only for HTTP

---

## Project structure

```
cmd/
  server/     HTTP API — accepts jobs, enqueues, serves status
  worker/     job processor — blocks on Redis, retries, updates status
internal/
  db/         schema, migrations, query functions
  queue/      Redis push/pop wrapper
```