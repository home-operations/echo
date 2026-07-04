# echo

![Version: 0.0.0](https://img.shields.io/badge/Version-0.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.0](https://img.shields.io/badge/AppVersion-0.0.0-informational?style=flat-square)

An HTTP/HTTPS server that echoes each incoming request back to the caller as
JSON — a debugging aid for inspecting exactly what a client (or a proxy/ingress
in front of it) sends. Inspired by
[mendhak/docker-http-https-echo](https://github.com/mendhak/docker-http-https-echo).

**Homepage:** <https://github.com/home-operations/echo>

## Installing

The chart is published as a Cosign-signed OCI artifact:

```bash
helm install echo oci://ghcr.io/home-operations/charts/echo --version <version>
```

Verify the signature (keyless, GitHub Actions OIDC):

```bash
cosign verify ghcr.io/home-operations/charts/echo:<version> \
  --certificate-identity-regexp '^https://github.com/home-operations/echo/\.github/workflows/release\.yaml@.*$' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

## Configuration

echo is configured entirely through environment variables; this chart exposes
them as the structured `config.*` values below, each mapping to an `ECHO_*`
variable the binary reads.

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| home-operations | <contact@home-operations.com> |  |

## Source Code

* <https://github.com/home-operations/echo>

## Requirements

Kubernetes: `>=1.25.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling (templated). |
| autoscaling.behavior | object | `{}` | Scaling behavior (scaleUp/scaleDown policies, templated); empty uses the cluster defaults. |
| autoscaling.enabled | bool | `false` | Create a HorizontalPodAutoscaler (autoscaling/v2) targeting the Deployment. |
| autoscaling.maxReplicas | int | `5` | Maximum replicas. |
| autoscaling.minReplicas | int | `1` | Minimum replicas. |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target average CPU utilization (%); 0 disables the CPU metric. |
| autoscaling.targetMemoryUtilizationPercentage | int | `0` | Target average memory utilization (%); 0 disables the memory metric. |
| config.commandsEnabled | bool | `true` | Let callers shape the response — status code, delay, extra headers — via echo-* query params / X-Echo-* headers (ECHO_COMMANDS_ENABLED). Off makes echo a pure reflector; pretty-printing stays available either way. |
| config.disableRequestLogs | bool | `false` | Silence the per-request access log (ECHO_DISABLE_REQUEST_LOGS); the echo response is unaffected. |
| config.echoBackToClient | bool | `true` | Return the JSON document to the client (ECHO_BACK_TO_CLIENT); false replies 204 No Content. |
| config.httpPort | int | `8080` | HTTP listen port (ECHO_HTTP_PORT). |
| config.kubernetes | bool | `false` | Add a `kubernetes` object (pod/namespace/IP/node) to responses, injected via the Downward API (ECHO_KUBERNETES). Off by default; it reveals cluster topology to callers. |
| config.logFormat | string | `"json"` | Log format (ECHO_LOG_FORMAT): json or text. |
| config.logLevel | string | `"info"` | Log level (ECHO_LOG_LEVEL): debug, info, warn, or error. |
| config.maxBodyBytes | int | `1048576` | Maximum request body bytes read and echoed (ECHO_MAX_BODY_BYTES); larger bodies are flagged truncated. |
| config.maxDelay | string | `"10s"` | Cap on the artificial response delay a caller may request via echo-delay (ECHO_MAX_DELAY); larger values are clamped. Keep it below the server's write timeout (~30s). |
| config.metricsEnabled | bool | `true` | Expose Prometheus metrics at /metrics on metricsPort (ECHO_METRICS_ENABLED). Disabling removes the metrics listener, container port, Service port, and ServiceMonitor entirely; health probes are unaffected (they target the http port). |
| config.metricsPort | int | `8081` | Metrics listen port (ECHO_METRICS_PORT), kept off the public echo port. |
| config.prettyPrint | bool | `false` | Indent the JSON response by default (ECHO_PRETTY_PRINT); callers can still override per request with echo-pretty-print. |
| config.trustedProxies | list | `[]` | CIDRs whose X-Forwarded-For header is trusted for client-IP resolution (ECHO_TRUSTED_PROXIES); comma-joined into the env var. |
| config.wsAllowedOrigins | list | `[]` | Origin host patterns allowed to open a WebSocket (ECHO_WS_ALLOWED_ORIGINS); empty allows any origin (the endpoint only echoes the caller's own frames). |
| config.wsEnabled | bool | `true` | Serve a WebSocket echo at /ws (ECHO_WS_ENABLED); non-upgrade requests to /ws fall through to the HTTP echo. |
| deploymentAnnotations | object | `{}` | Annotations added to the Deployment (workload) metadata, e.g. `reloader.stakater.com/auto: "true"`. |
| env | object | `{}` | Extra environment variables for the echo container, as a map (templated). |
| envFrom | list | `[]` | Sources of environment variables for the echo container (templated), e.g. a Secret with TLS material. |
| extraEnv | list | `[]` | Extra environment variables for the echo container, as a raw list (templated). |
| fullnameOverride | string | `""` | Override the generated name used for every resource's `metadata.name` (the chart "fullname"). |
| httpRoute.annotations | object | `{}` | HTTPRoute annotations. |
| httpRoute.apiVersion | string | `""` | HTTPRoute apiVersion; empty defaults to gateway.networking.k8s.io/v1. |
| httpRoute.enabled | bool | `false` | Expose echo via a Gateway API HTTPRoute (alternative to ingress). |
| httpRoute.hostnames | list | `[]` | Hostnames matched against the Host header (templated). |
| httpRoute.kind | string | `""` | HTTPRoute kind; empty defaults to HTTPRoute. |
| httpRoute.labels | object | `{}` | HTTPRoute labels. |
| httpRoute.matches | list | `[{"path":{"type":"PathPrefix","value":"/"}}]` | Match conditions for the default rule. |
| httpRoute.parentRefs | list | `[]` | Gateways (and listeners) this route attaches to. |
| image.digest | string | `""` | Pin the image by digest (sha256:…); set by the release pipeline. When set, it overrides the tag. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.repository | string | `"ghcr.io/home-operations/echo"` | echo image repository. |
| image.tag | string | `""` | Overrides the image tag; defaults to the chart appVersion. |
| imagePullSecrets | list | `[]` | Image pull secrets for private registries. |
| ingress.annotations | object | `{}` | Ingress annotations. |
| ingress.className | string | `""` | IngressClass name. |
| ingress.enabled | bool | `false` | Expose echo via an Ingress. |
| ingress.hosts | list | `[{"host":"echo.example.com","paths":[{"path":"/","pathType":"Prefix"}]}]` | Ingress hosts and their paths. |
| ingress.tls | list | `[]` | Ingress TLS configuration. |
| livenessProbe | object | `{"httpGet":{"path":"/healthz","port":"http"},"initialDelaySeconds":5,"periodSeconds":20}` | Liveness probe. Targets /healthz on the main http port, so it works regardless of the metrics toggle. |
| monitoring.serviceMonitor.annotations | object | `{}` | ServiceMonitor annotations. |
| monitoring.serviceMonitor.enabled | bool | `false` | Create a Prometheus Operator ServiceMonitor (requires its CRDs and config.metricsEnabled). |
| monitoring.serviceMonitor.interval | string | `"30s"` | Scrape interval. |
| monitoring.serviceMonitor.labels | object | `{}` | ServiceMonitor labels. |
| monitoring.serviceMonitor.metricRelabelings | list | `[]` | Prometheus metric relabelings. |
| monitoring.serviceMonitor.path | string | `"/metrics"` | Metrics path. |
| monitoring.serviceMonitor.relabelings | list | `[]` | Prometheus relabelings. |
| monitoring.serviceMonitor.scrapeTimeout | string | `"10s"` | Scrape timeout. |
| nameOverride | string | `""` | Override the app name used in the `app.kubernetes.io/name` label. |
| nodeSelector | object | `{}` | Node selector for pod scheduling (templated). |
| podAnnotations | object | `{}` | Annotations added to the pod. |
| podDisruptionBudget.enabled | bool | `false` | Create a PodDisruptionBudget. Off by default; useful when running multiple replicas. |
| podDisruptionBudget.maxUnavailable | string | `""` | Maximum pods that may be unavailable, as a count or percentage; takes precedence over `minAvailable` when set. @schema type: [integer, string] @schema |
| podDisruptionBudget.minAvailable | int | `1` | Minimum pods that must stay available, as a count or percentage. Used unless `maxUnavailable` is set. @schema type: [integer, string] @schema |
| podLabels | object | `{}` | Labels added to the pod. |
| podSecurityContext | object | `{"fsGroup":65532,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level securityContext (runs as non-root uid/gid 65532 with the default seccomp profile). |
| priorityClassName | string | `""` | PriorityClass for the pod (templated); empty uses the cluster default. |
| readinessProbe | object | `{"httpGet":{"path":"/healthz","port":"http"},"initialDelaySeconds":2,"periodSeconds":10}` | Readiness probe. Targets /healthz on the main http port, so it works regardless of the metrics toggle. |
| replicaCount | int | `1` | Number of echo replicas (echo is stateless, so it scales horizontally). Ignored when autoscaling.enabled. |
| resources | object | `{}` | echo container resource requests/limits. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | echo container securityContext (no privilege escalation, read-only root filesystem, drops ALL capabilities). |
| service.annotations | object | `{}` | Service annotations. |
| service.externalTrafficPolicy | string | `""` | Service externalTrafficPolicy (`Local` preserves the client source IP; only applies to NodePort/LoadBalancer). Empty uses the cluster default. |
| service.port | int | `80` | HTTP service port. |
| service.type | string | `"ClusterIP"` | Service type. |
| serviceAccount.annotations | object | `{}` | ServiceAccount annotations. |
| serviceAccount.automount | bool | `false` | Mount the API token. echo never calls the Kubernetes API, so this is off by default. |
| serviceAccount.create | bool | `true` | Create a ServiceAccount. |
| serviceAccount.name | string | `""` | ServiceAccount name; empty uses the chart fullname. |
| startupProbe | object | `{"failureThreshold":30,"httpGet":{"path":"/healthz","port":"http"},"periodSeconds":5}` | Startup probe; gates liveness/readiness until echo is up, so a slow start can't trigger a premature restart. Targets the http port. Set to {} to disable. |
| terminationGracePeriodSeconds | int | `30` | Grace period for a clean shutdown. |
| tests.image.pullPolicy | string | `"IfNotPresent"` | `helm test` image pull policy. |
| tests.image.repository | string | `"mirror.gcr.io/curlimages/curl"` | `helm test` connection-pod image; a gcr-mirrored curl, so the test never pulls from Docker Hub. |
| tests.image.tag | string | `"8.21.0@sha256:7c12af72ceb38b7432ab85e1a265cff6ae58e06f95539d539b654f2cfa64bb13"` | `helm test` image, pinned as `tag@sha256:digest` so Renovate bumps the tag and its digest together. |
| tolerations | list | `[]` | Tolerations for pod scheduling (templated). |
| topologySpreadConstraints | list | `[]` | Topology spread constraints for the pods (templated); relevant at replicaCount > 1, e.g. to spread across zones. |
| volumeMounts | list | `[]` | Additional volume mounts on the echo container (templated). |
| volumes | list | `[]` | Additional volumes on the Deployment pod (templated). |

---

_This README is generated by [helm-docs](https://github.com/norwoodj/helm-docs) from `Chart.yaml` and `values.yaml`. Edit those (or `README.md.gotmpl`) and run `mise run generate`._
