# healthcheck - API Health Monitor CLI

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/badge/CI-passing-brightgreen?style=flat-square)]()

A fast, lightweight CLI tool for monitoring API endpoint health. Supports concurrent checks, configurable intervals, Slack/webhook notifications, and colored terminal output. Built for SRE and DevOps workflows.

## Demo

```
$ healthcheck check https://api.example.com https://status.example.com/health

  ENDPOINT                              STATUS    TIME
  ────────────────────────────────────────────────────────
  ✔ https://api.example.com             200 OK    45ms
  ✔ https://status.example.com/health   200 OK    112ms

  2/2 healthy — total time 156ms
```

```
$ healthcheck watch --interval 30s --config endpoints.yaml

  [14:32:01] Checking 4 endpoints...
  ✔ https://api.example.com             200 OK      45ms
  ✔ https://status.example.com/health   200 OK     112ms
  ✘ https://staging.example.com/api     503 ERROR  2015ms
  ✔ https://cdn.example.com/ping        200 OK      23ms

  3/4 healthy — next check in 30s

  [14:32:31] Checking 4 endpoints...
  ✔ https://api.example.com             200 OK      41ms
  ✔ https://status.example.com/health   200 OK      98ms
  ✔ https://staging.example.com/api     200 OK     187ms
  ✔ https://cdn.example.com/ping        200 OK      19ms

  4/4 healthy — next check in 30s
```

## Features

- **Concurrent health checks** — all endpoints checked in parallel using goroutines
- **YAML/JSON config** — define endpoints, headers, and thresholds in a config file
- **Slack webhook alerts** — get notified on failures via Slack incoming webhooks
- **Colored terminal output** — green for healthy, red for failures, at a glance
- **Exit codes for CI/CD** — returns non-zero when any endpoint is unhealthy
- **Response time tracking** — measure and display latency for every request
- **Custom headers & auth** — set Authorization, API keys, or any custom header
- **Retry with backoff** — configurable retries with exponential backoff on failure

## Installation

```bash
go install github.com/enriquevaldivia1988/api-health-cli@latest
```

Or build from source:

```bash
git clone https://github.com/enriquevaldivia1988/api-health-cli.git
cd healthcheck
make build
```

## Usage

### Single URL check

```bash
healthcheck check https://api.example.com/health
```

### Multiple endpoints

```bash
healthcheck check https://api.example.com https://status.example.com/health
```

### Using a config file

```bash
healthcheck check --config endpoints.yaml
```

### Watch mode (continuous monitoring)

```bash
healthcheck watch --config endpoints.yaml --interval 30s
```

### CI/CD mode (exit code reflects health)

```bash
healthcheck check --config endpoints.yaml --ci
# Exit code 0 = all healthy, 1 = at least one failure
```

### With custom timeout and retries

```bash
healthcheck check https://api.example.com --timeout 10s --retries 3
```

## Config File

Create a YAML config to define endpoints and their settings:

```yaml
# endpoints.yaml
endpoints:
  - url: https://api.example.com/health
    method: GET
    timeout: 5s
    retries: 2
    headers:
      Authorization: "Bearer ${API_TOKEN}"

  - url: https://status.example.com/health
    method: GET
    timeout: 10s

  - url: https://internal.example.com/readiness
    method: GET
    timeout: 3s
    expected_status: 200

notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    on_failure: true
```

See [`config.example.yaml`](config.example.yaml) for a full example.

## Flags

| Flag         | Short | Default | Description                        |
|--------------|-------|---------|------------------------------------|
| `--config`   | `-c`  |         | Path to YAML/JSON config file      |
| `--timeout`  | `-t`  | `5s`    | HTTP request timeout               |
| `--retries`  | `-r`  | `0`     | Number of retries on failure       |
| `--interval` | `-i`  | `30s`   | Check interval for watch mode      |
| `--ci`       |       | `false` | CI mode: exit 1 on any failure     |
| `--slack`    |       |         | Slack webhook URL for notifications|
| `--no-color` |       | `false` | Disable colored output             |

## Exit Codes

| Code | Meaning                     |
|------|-----------------------------|
| `0`  | All endpoints healthy       |
| `1`  | One or more endpoints failed|
| `2`  | Configuration error         |

## Author

**Enrique Valdivia Rios**
- GitHub: [@enriquevaldivia1988](https://github.com/enriquevaldivia1988)
- LinkedIn: [enrique-valdivia-rios](https://linkedin.com/in/enrique-valdivia-rios)
- Web: [enriquevaldivia.dev](https://enriquevaldivia.dev)

## License

[MIT](LICENSE)
