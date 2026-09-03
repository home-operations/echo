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
- Plain HTTP by default; terminate TLS at the ingress, and `protocol` is read from `X-Forwarded-Proto` when the peer is a trusted proxy. Set `ECHO_HTTPS_PORT` with a certificate and key to also serve the echo natively over TLS — see below.
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
| Status code     | `echo-code` / `X-Echo-Code`                 | Respond with this status (200–599); invalid values are ignored.             |
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
`echo-header=Set-Cookie:...` — the `Domain` attribute is stripped so a cookie
cannot be scoped to a shared parent domain. Request cookies are always
reflected back in the
`cookies` field. Because `echo-delay` holds the connection open, keep
`ECHO_MAX_DELAY` modest (or set `ECHO_COMMANDS_ENABLED=false`) when echo is
reachable from untrusted networks, and rate-limit at the ingress.

## Native HTTPS

For testing backend TLS (a Gateway API `BackendTLSPolicy`, a service mesh, or a
client's trust store) echo can serve the same echo over TLS on a second port.
Point it at a PEM certificate and key; the listener negotiates TLS 1.2+ and
HTTP/2, reports `"protocol": "https"`, and adds a `tls` block with what the
client negotiated:

```console
$ ECHO_HTTPS_PORT=8443 ECHO_HTTPS_CERT=tls.crt ECHO_HTTPS_KEY=tls.key ./echo &
$ curl -s --cacert ca.crt https://echo.example.com:8443/ | jq .protocol,.tls
"https"
{
  "version": "TLS 1.3",
  "cipherSuite": "TLS_AES_128_GCM_SHA256",
  "serverName": "echo.example.com",
  "negotiatedProtocol": "h2"
}
```

The certificate is loaded once at startup; restart echo to pick up a rotated
one. In the chart, set `config.httpsPort` and mount the TLS Secret at
`/etc/echo/tls` via `volumes`/`volumeMounts` (the default `config.httpsCert`/
`config.httpsKey` paths match a `kubernetes.io/tls` Secret); an `https` port is
then added to the container and the Service.

## Configuration

Set via environment variables:

| Variable                    | Default   | Description                                                                         |
| --------------------------- | --------- | ----------------------------------------------------------------------------------- |
| `ECHO_HTTP_PORT`            | `8080`    | HTTP listen port (also serves the `/healthz` probe)                                 |
| `ECHO_HTTPS_PORT`           | `0`       | HTTPS listen port serving the same echo over TLS; `0` disables                      |
| `ECHO_HTTPS_CERT`           | _(empty)_ | Path to the PEM certificate for `ECHO_HTTPS_PORT` (required with it)                |
| `ECHO_HTTPS_KEY`            | _(empty)_ | Path to the PEM private key for `ECHO_HTTPS_PORT` (required with it)                |
| `ECHO_METRICS_ENABLED`      | `true`    | Expose Prometheus metrics; disabling removes the metrics listener                   |
| `ECHO_METRICS_PORT`         | `8081`    | Metrics listen port (`/metrics` only)                                               |
| `ECHO_LOG_LEVEL`            | `info`    | `debug`, `info`, `warn`, or `error`                                                 |
| `ECHO_LOG_FORMAT`           | `json`    | `json` or `text`                                                                    |
| `ECHO_DISABLE_REQUEST_LOGS` | `false`   | Silence the per-request access log                                                  |
| `ECHO_BACK_TO_CLIENT`       | `true`    | Return the JSON body, or `204` when false                                           |
| `ECHO_MAX_BODY_BYTES`       | `1048576` | Max request body bytes read and echoed                                              |
| `ECHO_COMMANDS_ENABLED`     | `true`    | Allow callers to shape the response (`echo-*` query / `X-Echo-*` headers)           |
| `ECHO_MAX_DELAY`            | `10s`     | Cap on the caller-requested `echo-delay`; itself capped under the 30s write timeout |
| `ECHO_PRETTY_PRINT`         | `false`   | Indent the JSON response by default (overridable with `echo-pretty-print`)          |
| `ECHO_WS_ENABLED`           | `true`    | Serve the WebSocket echo at `/ws`                                                   |
| `ECHO_WS_ALLOWED_ORIGINS`   | _(empty)_ | Allowed WebSocket Origin host patterns (comma-separated); empty allows any          |
| `ECHO_WS_IDLE_TIMEOUT`      | `5m`      | Close a WebSocket that has sent nothing for this long; `0` disables                 |
| `ECHO_TRUSTED_PROXIES`      | _(empty)_ | Trusted-proxy CIDRs for `X-Forwarded-For`/`-Proto` (comma-separated)                |
| `ECHO_KUBERNETES`           | `false`   | Add a `kubernetes` block (pod/node identity via the Downward API)                   |
| `ECHO_SHUTDOWN_TIMEOUT`     | `15s`     | Graceful shutdown timeout                                                           |

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
