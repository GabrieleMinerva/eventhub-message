# go-worker-service

Worker service in Go for event-driven pipelines.

## Features
- Event consumption from Azure Event Hubs (`WORKER_MODE=eventhub`) or mock generator (`WORKER_MODE=mock`).
- Event validation and processing pipeline.
- Health/readiness probes for deployment platforms (Render/Kubernetes).
- Unit tests for core parsing, config and worker loop.

## Quickstart
```bash
go mod tidy
go test ./...
go run ./cmd/worker
```

By default it starts in mock mode and emits a synthetic event every 5 seconds.

## Configuration
| Variable | Default | Description |
|---|---:|---|
| `WORKER_MODE` | `mock` | `mock` or `eventhub` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `HTTP_PORT` | `8080` | Port for health endpoints |
| `MOCK_TICK_INTERVAL` | `5s` | Event generation interval in mock mode |
| `EVENTHUB_CONNECTION_STRING` | - | Required in `eventhub` mode |
| `EVENTHUB_NAME` | - | Required in `eventhub` mode |
| `EVENTHUB_CONSUMER_GROUP` | `$Default` | Consumer group |
| `CHECKPOINT_INTERVAL` | `30s` | Checkpoint cadence placeholder |
| `EVENTHUB_EVENTS_FILE` | - | Required in `eventhub` mode in this environment (NDJSON input) |
| `RECEIVE_ERROR_BACKOFF` | `250ms` | Backoff after receive errors |
| `DEDUP_WINDOW` | `10m` | In-memory deduplication TTL by event id |
| `DEDUP_MAX_ENTRIES` | `10000` | Maximum in-memory dedup cache size |

## HTTP Endpoints
- `GET /healthz` -> `ok`
- `GET /readyz` -> JSON with `processed` / `failed` / `skipped` counters (503 in degraded state)

## Render deployment
Use `render.yaml` in this repo to create a Background Worker service.
