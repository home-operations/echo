# echo

Echoes each HTTP request back as JSON, which is handy for inspecting what a
client or proxy actually sends. Similar to mendhak/docker-http-https-echo.

```console
$ curl -s 'http://localhost:8080/hello?name=world' -d 'hi'
{
  "timestamp": "2026-06-14T19:00:00Z",
  "protocol": "http",
  "method": "POST",
  "host": "localhost:8080",
  "hostname": "localhost",
  "path": "/hello",
  "url": "/hello?name=world",
  "query": { "name": ["world"] },
  "headers": { "Content-Type": ["application/x-www-form-urlencoded"] },
  "body": "hi",
  "remoteAddr": "127.0.0.1:54321",
  "ip": "127.0.0.1",
  "os": { "hostname": "echo-7d9c" }
}
```

- Any path or method is echoed as JSON.
- `/ws` upgrades to a WebSocket and echoes each message; a plain request to `/ws` is echoed normally.
- `/healthz` returns `{"status":"ok"}`; `/metrics` serves Prometheus metrics on `:9090`.
- Plain HTTP only; terminate TLS at the ingress. `protocol` is read from `X-Forwarded-Proto`.
- Responses are `application/json` with `X-Content-Type-Options: nosniff`. Bodies are capped (1 MiB default) and flagged when truncated.
- The client IP is read from `X-Forwarded-For` for trusted proxies.
- With `ECHO_KUBERNETES=true` (chart `config.kubernetes`), adds a `kubernetes` block (pod, namespace, IP, node) from the Downward API. Off by default.

## Configuration

Set via environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `ECHO_HTTP_PORT` | `8080` | HTTP listen port |
| `ECHO_METRICS_ENABLED` | `true` | Expose Prometheus metrics |
| `ECHO_METRICS_ADDR` | `:9090` | Metrics listen address |
| `ECHO_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `ECHO_LOG_FORMAT` | `json` | `json` or `text` |
| `ECHO_DISABLE_REQUEST_LOGS` | `false` | Silence the per-request access log |
| `ECHO_BACK_TO_CLIENT` | `true` | Return the JSON body, or `204` when false |
| `ECHO_MAX_BODY_BYTES` | `1048576` | Max request body bytes read and echoed |
| `ECHO_WS_ENABLED` | `true` | Serve the WebSocket echo at `/ws` |
| `ECHO_WS_ALLOWED_ORIGINS` | _(empty)_ | Allowed WebSocket Origin host patterns (comma-separated); empty allows any |
| `ECHO_TRUSTED_PROXIES` | _(empty)_ | Trusted-proxy CIDRs for `X-Forwarded-For` (comma-separated) |
| `ECHO_KUBERNETES` | `false` | Add a `kubernetes` block (pod/node identity via the Downward API) |
| `ECHO_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown timeout |

## Running

Container:

```bash
docker run --rm -p 8080:8080 ghcr.io/home-operations/echo:rolling
```

Helm (Cosign-signed OCI chart; values are documented in charts/echo):

```bash
helm install echo oci://ghcr.io/home-operations/charts/echo --version <version>
```

## Development

mise manages the toolchain.

```bash
mise install            # install pinned tools (go, golangci-lint, helm, etc.)
mise run build          # go build ./...
mise run test           # go test -race
mise run lint           # golangci-lint
mise run helm-unittest  # chart tests
mise run generate       # regenerate chart README + schema
```

## License

See LICENSE.
