// Package echo turns an HTTP request into the JSON document echo returns to the
// caller.
package echo

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// Options tunes how a Request is built.
type Options struct {
	// Now is the timestamp recorded in the response (injected for testability).
	Now time.Time
	// Hostname is the server's OS hostname, resolved once at startup.
	Hostname string
	// TrustedProxies are CIDRs whose X-Forwarded-For entries are trusted when
	// resolving the client IP.
	TrustedProxies []netip.Prefix
	// Kubernetes, when non-nil, is added to the response as a kubernetes object.
	Kubernetes *KubernetesInfo
}

// Request is the JSON document echoed back to the client. omitempty keeps
// optional sections out of the output when they are unused.
type Request struct {
	Timestamp     string              `json:"timestamp"`
	Protocol      string              `json:"protocol"`
	Method        string              `json:"method"`
	Host          string              `json:"host"`
	Hostname      string              `json:"hostname"`
	Path          string              `json:"path"`
	URL           string              `json:"url"`
	Query         map[string][]string `json:"query"`
	Headers       map[string][]string `json:"headers"`
	Cookies       map[string]string   `json:"cookies,omitempty"`
	Body          string              `json:"body,omitempty"`
	JSON          json.RawMessage     `json:"json,omitempty"`
	BodyTruncated bool                `json:"bodyTruncated,omitempty"`
	RemoteAddr    string              `json:"remoteAddr"`
	IP            string              `json:"ip"`
	IPs           []string            `json:"ips,omitempty"`
	OS            OS                  `json:"os"`
	Kubernetes    *KubernetesInfo     `json:"kubernetes,omitempty"`
	// Applied reports the response-shaping directives (echo-code/delay/header)
	// echo honored for this request; omitted when none applied.
	Applied *Applied `json:"applied,omitempty"`
}

// OS describes the server process's host.
type OS struct {
	Hostname string `json:"hostname"`
}

// KubernetesInfo carries pod/node identity from the Downward API. It is only
// populated (and emitted) when ECHO_KUBERNETES is enabled.
type KubernetesInfo struct {
	PodName      string `json:"podName,omitempty"`
	PodNamespace string `json:"podNamespace,omitempty"`
	PodIP        string `json:"podIP,omitempty"`
	NodeName     string `json:"nodeName,omitempty"`
}

// ReadBody reads up to max bytes of the request body, reporting whether the
// body was truncated (i.e. more bytes remained).
func ReadBody(r *http.Request, max int64) ([]byte, bool, error) {
	if r.Body == nil || max <= 0 {
		return nil, false, nil
	}
	// Read one extra byte so we can tell a full body from a clipped one.
	buf, err := io.ReadAll(io.LimitReader(r.Body, max+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(buf)) > max {
		return buf[:max], true, nil
	}
	return buf, false, nil
}

// Build assembles the echo document. body is the already-read (and possibly
// truncated) request body; Build performs no I/O of its own.
func Build(r *http.Request, body []byte, truncated bool, opts Options) *Request {
	// echo serves plain HTTP; report the client-facing scheme from the
	// TLS-terminating proxy's X-Forwarded-Proto — gated on the peer being a
	// trusted proxy, the same trust rule X-Forwarded-For gets.
	protocol := "http"
	if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" && ipTrusted(peerHost(r.RemoteAddr), opts.TrustedProxies) {
		protocol, _, _ = strings.Cut(xfp, ",")
		protocol = strings.TrimSpace(protocol)
	}

	out := &Request{
		Timestamp:     opts.Now.UTC().Format(time.RFC3339Nano),
		Protocol:      protocol,
		Method:        r.Method,
		Host:          r.Host,
		Hostname:      hostnameOnly(r.Host),
		Path:          r.URL.Path,
		URL:           r.URL.RequestURI(),
		Query:         map[string][]string(r.URL.Query()),
		Headers:       map[string][]string(r.Header),
		BodyTruncated: truncated,
		RemoteAddr:    r.RemoteAddr,
		OS:            OS{Hostname: opts.Hostname},
	}

	if len(body) > 0 {
		out.Body = string(body)
		if isJSON(r.Header.Get("Content-Type")) && json.Valid(body) {
			out.JSON = json.RawMessage(body)
		}
	}

	if c := cookieMap(r); len(c) > 0 {
		out.Cookies = c
	}

	out.IP, out.IPs = clientIP(r, opts.TrustedProxies)
	out.Kubernetes = opts.Kubernetes

	return out
}

func hostnameOnly(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func isJSON(contentType string) bool {
	ct, _, _ := strings.Cut(strings.ToLower(contentType), ";")
	ct = strings.TrimSpace(ct)
	return ct == "application/json" || strings.HasSuffix(ct, "+json")
}

func cookieMap(r *http.Request) map[string]string {
	cs := r.Cookies()
	if len(cs) == 0 {
		return nil
	}
	out := make(map[string]string, len(cs))
	for _, c := range cs {
		// First occurrence wins on duplicate names, matching how clients and
		// most servers resolve repeated cookies.
		if _, ok := out[c.Name]; !ok {
			out[c.Name] = c.Value
		}
	}
	return out
}

// peerHost strips the port from a RemoteAddr-style host:port.
func peerHost(remoteAddr string) string {
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return h
	}
	return remoteAddr
}

// clientIP returns the best-guess client IP and the X-Forwarded-For chain. The
// immediate peer is the client unless it is a trusted proxy, in which case the
// right-most untrusted address in X-Forwarded-For is used.
func clientIP(r *http.Request, trusted []netip.Prefix) (string, []string) {
	peer := peerHost(r.RemoteAddr)

	var chain []string
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for p := range strings.SplitSeq(xff, ",") {
			if p = strings.TrimSpace(p); p != "" {
				chain = append(chain, p)
			}
		}
	}

	if len(trusted) == 0 || !ipTrusted(peer, trusted) {
		return peer, chain
	}
	for i := len(chain) - 1; i >= 0; i-- {
		if !ipTrusted(chain[i], trusted) {
			return chain[i], chain
		}
	}
	return peer, chain
}

func ipTrusted(s string, trusted []netip.Prefix) bool {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return false
	}
	for _, p := range trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
