# PulseDB

PulseDB is a lightweight uptime monitoring service built with Go. It lets you register HTTP endpoints, run periodic health checks, store check history, and expose operational metrics for Prometheus and Grafana.

## Why PulseDB?

- Monitor websites and APIs for availability and response time
- Track the history of checks for each endpoint
- Run locally with an in-memory store or connect to PostgreSQL
- Expose Prometheus metrics for dashboards and alerts
- Start quickly with Docker Compose for observability tooling

## Features

- CRUD REST API for monitors
- Background worker pool for periodic HTTP probes
- PostgreSQL-backed storage with an in-memory fallback for development
- Prometheus metrics for traffic and probe activity
- Docker Compose setup for PostgreSQL, Prometheus, and Grafana
- Graceful shutdown support with context cancellation

## Tech stack

- Go 1.25+
- Gin web framework
- PostgreSQL with pgx
- Prometheus client library
- Grafana
- Docker and Docker Compose

## Project structure

```text
pulseDb/
├── cmd/api/
├── internal/
├── docker-compose.yml
├── prometheus.yml
├── Dockerfile
├── go.mod
└── go.sum
```

## Getting started

### Prerequisites

- Go 1.25 or newer
- Docker and Docker Compose (optional, for the full observability stack)

### Run locally

```bash
git clone https://github.com/vampire321/PulseDB.git
cd PulseDB
go mod download
go run ./cmd/api
```

The application starts on port 8080. If PostgreSQL is unavailable, it automatically falls back to the in-memory repository.

### Run with Docker Compose

```bash
docker compose up -d
go run ./cmd/api
```

This starts:
- PostgreSQL on port 5432
- Prometheus on port 9090
- Grafana on port 3000

## API overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/monitors | Create a new monitor |
| GET | /api/v1/monitors | List all monitors |
| GET | /api/v1/monitors/:id | Get a monitor by ID |
| PUT | /api/v1/monitors/:id | Update a monitor |
| DELETE | /api/v1/monitors/:id | Delete a monitor |
| GET | /api/v1/monitors/:id/checks | Get check history |
| GET | /health | Health check endpoint |
| GET | /metrics | Prometheus metrics endpoint |

### Example: create a monitor

```bash
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -d '{"name":"Google","url":"https://www.google.com","interval_s":30}'
```

## Observability

PulseDB exposes Prometheus metrics at the /metrics endpoint, including:

- HTTP request counters and latency histograms
- Active monitor gauges
- Probe success and error counters
- Checker queue depth

You can point Prometheus to the app instance and visualize the data in Grafana.

## License

This project is intended for educational and experimental use.
