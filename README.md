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
- `/healthz` (liveness) and `/readyz` (readiness) return `{"status":"ok"}` on the main echo port; `/metrics` serves Prometheus metrics on its own optional port (`:8081`), separate from the echo port.
- Plain HTTP only; terminate TLS at the ingress. `protocol` is read from `X-Forwarded-Proto`.
- Responses are `application/json` with `X-Content-Type-Options: nosniff`. Bodies are capped (1 MiB default) and flagged when truncated.
- The client IP is read from `X-Forwarded-For` for trusted proxies.
- With `ECHO_KUBERNETES=true` (chart `config.kubernetes`), adds a `kubernetes` block (pod, namespace, IP, node) from the Downward API. Off by default.
- Callers can shape the response (status code, delay, extra headers) and pretty-print it — see below.

## Shaping the response

Beyond reflecting the request, a caller can tell echo how to _respond_, which
turns it into a test target for ingress, proxies, retries, and timeouts. Each
directive is read from an `echo-*` query parameter or the matching `X-Echo-*`
header (the query parameter wins if both are set):

| Directive       | Query / header                              | Effect                                                                      |
| --------------- | ------------------------------------------- | --------------------------------------------------------------------------- |
| Status code     | `echo-code` / `X-Echo-Code`                 | Respond with this status (100–599); invalid values are ignored.             |
| Delay           | `echo-delay` / `X-Echo-Delay`               | Wait this long before responding (Go duration), capped at `ECHO_MAX_DELAY`. |
| Response header | `echo-header` / `X-Echo-Header`             | Add a `Name: Value` response header; repeat for more than one.              |
| Response cookie | `echo-cookie` / `X-Echo-Cookie`             | Set a `name:value` response cookie; repeat for more than one.               |
| Pretty-print    | `echo-pretty-print` / `X-Echo-Pretty-Print` | Indent the JSON response. Always available, even when commands are off.     |

```console
$ curl -s -o /dev/null -w '%{http_code}\n' 'http://localhost:8080/?echo-code=503'
503
$ curl -s 'http://localhost:8080/?echo-delay=2s&echo-header=X-Test:1&echo-pretty-print'
# ...waits 2s, sets X-Test: 1, and returns indented JSON with an "applied" block
```

The response echoes which shaping directives were honored in an `applied` block,
so a directive that was ignored (out of range, or commands disabled) is easy to
spot. Response shaping is gated by `ECHO_COMMANDS_ENABLED` (on by default); set
it to `false` to make echo a pure reflector. Pretty-printing is independent and
always available. Header names use dashes — underscored header names are
silently dropped by some proxies (e.g. ingress-nginx).

`echo-header` can set any header except echo's own `Content-Type`,
`Content-Length`, `X-Content-Type-Options`, and `Cache-Control`, which are
reserved so the JSON response stays coherent and inert. `echo-cookie` sets a
bare `name=value` cookie; for attributes (`Path`, `HttpOnly`, …) use
`echo-header=Set-Cookie:...`. Request cookies are always reflected back in the
`cookies` field. Because `echo-delay` holds the connection open, keep
`ECHO_MAX_DELAY` modest (or set `ECHO_COMMANDS_ENABLED=false`) when echo is
reachable from untrusted networks, and rate-limit at the ingress.

## Configuration

Set via environment variables:

| Variable                    | Default   | Description                                                                |
| --------------------------- | --------- | -------------------------------------------------------------------------- |
| `ECHO_HTTP_PORT`            | `8080`    | HTTP listen port (also serves the `/healthz` probe)                        |
| `ECHO_METRICS_ENABLED`      | `true`    | Expose Prometheus metrics; disabling removes the metrics listener          |
| `ECHO_METRICS_PORT`         | `8081`    | Metrics listen port (`/metrics` only)                                      |
| `ECHO_LOG_LEVEL`            | `info`    | `debug`, `info`, `warn`, or `error`                                        |
| `ECHO_LOG_FORMAT`           | `json`    | `json` or `text`                                                           |
| `ECHO_DISABLE_REQUEST_LOGS` | `false`   | Silence the per-request access log                                         |
| `ECHO_BACK_TO_CLIENT`       | `true`    | Return the JSON body, or `204` when false                                  |
| `ECHO_MAX_BODY_BYTES`       | `1048576` | Max request body bytes read and echoed                                     |
| `ECHO_COMMANDS_ENABLED`     | `true`    | Allow callers to shape the response (`echo-*` query / `X-Echo-*` headers)  |
| `ECHO_MAX_DELAY`            | `10s`     | Cap on the caller-requested `echo-delay`; larger values are clamped        |
| `ECHO_PRETTY_PRINT`         | `false`   | Indent the JSON response by default (overridable with `echo-pretty-print`) |
| `ECHO_WS_ENABLED`           | `true`    | Serve the WebSocket echo at `/ws`                                          |
| `ECHO_WS_ALLOWED_ORIGINS`   | _(empty)_ | Allowed WebSocket Origin host patterns (comma-separated); empty allows any |
| `ECHO_TRUSTED_PROXIES`      | _(empty)_ | Trusted-proxy CIDRs for `X-Forwarded-For` (comma-separated)                |
| `ECHO_KUBERNETES`           | `false`   | Add a `kubernetes` block (pod/node identity via the Downward API)          |
| `ECHO_SHUTDOWN_TIMEOUT`     | `15s`     | Graceful shutdown timeout                                                  |

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
