package server

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "echo_http_requests_total",
		Help: "Total number of HTTP requests handled, by method and status class.",
	}, []string{"method", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "echo_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, by method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

// metricsHandler serves Prometheus metrics and the health/readiness probe
// endpoint on the dedicated monitoring listener, so scraping and probing share
// one port (8081 by default) that is separate from the public echo port.
func metricsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /healthz", handleHealth)
	return mux
}

// statusClass buckets a status code into a low-cardinality label (e.g. "2xx").
func statusClass(code int) string {
	return strconv.Itoa(code/100) + "xx"
}

// knownMethods bounds the method metric label. A client can send an arbitrary
// request method, so unrecognised ones collapse to "other" — otherwise distinct
// method strings would grow the metric's cardinality without bound (a memory
// exhaustion vector).
var knownMethods = map[string]struct{}{
	http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {},
	http.MethodPut: {}, http.MethodPatch: {}, http.MethodDelete: {},
	http.MethodConnect: {}, http.MethodOptions: {}, http.MethodTrace: {},
}

func methodLabel(method string) string {
	if _, ok := knownMethods[method]; ok {
		return method
	}
	return "other"
}
