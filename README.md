# PingPulse: Simple Uptime & Health Pinger

A minimal, portable Go app to check HTTP, ping, and database endpoints with Prometheus metrics. Runs anywhere: binary, Docker, Kubernetes.

## Features

- **HTTP, Ping, and DB checks** (Postgres, MySQL)
- **Parallel checks** with sensible defaults
- **Prometheus metrics**: up/down, response time, SSL expiry, success/failure counts
- **YAML config** with minimal required fields
- **Maintenance mode**: skip checks via config
- **Graceful config reload**: update checks without restart

## Quick Start

### 1. Build & Run (Locally)

```sh
git clone https://github.com/timmyb824/PingPulse
cd PingPulse
go build -o pingpulse
./pingpulse example-config.yaml
```

### 2. Run with Docker

```sh
docker build -t pingpulse .
# IMPORTANT: Add --cap-add=NET_RAW (docker) or --cap-add=net_raw (podman) for ping checks

# Example: Run with your own config file
# (Replace /path/to/your/config.yaml with your actual config location)
docker run --cap-add=NET_RAW -v /path/to/your/config.yaml:/app/config.yaml pingpulse /app/config.yaml
```

### 3. Run with Podman (for local testing only)

Podman is stricter about file permissions and SELinux when mounting files. Use the following command for reliable local testing:

```sh
podman build -t pingpulse .
podman run --rm --security-opt label=disable \
  -v "$(pwd)/config.yaml:/config.yaml:ro,Z" \
  localhost/pingpulse \
  /app/pingpulse /config.yaml
```

- `--security-opt label=disable` disables SELinux labels that can block execution.
- `:ro,Z` makes the mount read-only and relabels it for container sharing.
- This approach is recommended for Podman only. For production, use Docker or Kubernetes.

### 4. Kubernetes

Create a ConfigMap from your YAML config and mount it to `/app/example-config.yaml` in the container.

## Configuration Example (`example-config.yaml`)

```yaml
maintenance_mode: false
interval_seconds: 30
retries: 2
http_checks:
  - name: Google
    url: https://www.google.com
    timeout: 5
    accept_status_codes: [200, 301, 302]
ping_checks:
  - name: CloudflareDNS
    host: 1.1.1.1
    timeout: 2
db_checks:
  - name: ExamplePostgres
    driver: postgres
    dsn: "host=localhost port=5432 user=postgres password=secret dbname=postgres sslmode=disable"
    timeout: 3
  - name: ExampleMySQL
    driver: mysql
    dsn: "user:password@tcp(127.0.0.1:3306)/mysql"
    timeout: 3
```

## Prometheus Metrics

- `uptime_check_up{type, name}`: 1=up, 0=down
- `uptime_check_response_seconds{type, name}`: response time
- `uptime_check_ssl_days_left{name}`: SSL cert expiry (HTTP only)
- `uptime_check_success_total{type, name}`
- `uptime_check_failure_total{type, name}`
- `uptime_check_ssl_error_total{name}`

## Maintenance Mode

Set `maintenance_mode: true` in your config to skip all checks and mark all targets as down in metrics.

## Extending & Organizing Code

- Metrics are now defined in `metrics.go` for better organization.

## License

MIT
