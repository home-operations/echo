package echo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	body := []byte("hello")
	req := httptest.NewRequest(http.MethodPost, "https://example.com:8443/foo/bar?a=1&a=2&b=3", nil)
	req.Header.Set("X-Custom", "value")
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})
	req.Header.Set("X-Forwarded-Proto", "https")

	// X-Forwarded-Proto is only honored from a trusted proxy; httptest requests
	// arrive from 192.0.2.1.
	got := Build(req, body, false, Options{
		Now:            time.Unix(0, 0),
		Hostname:       "test-host",
		TrustedProxies: []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")},
	})

	if got.Method != http.MethodPost {
		t.Errorf("Method = %q, want POST", got.Method)
	}
	if got.Protocol != "https" {
		t.Errorf("Protocol = %q, want https", got.Protocol)
	}
	if got.Path != "/foo/bar" {
		t.Errorf("Path = %q, want /foo/bar", got.Path)
	}
	if got.Host != "example.com:8443" {
		t.Errorf("Host = %q, want example.com:8443", got.Host)
	}
	if got.Hostname != "example.com" {
		t.Errorf("Hostname = %q, want example.com", got.Hostname)
	}
	if q := got.Query["a"]; len(q) != 2 || q[0] != "1" || q[1] != "2" {
		t.Errorf("Query[a] = %v, want [1 2]", q)
	}
	if got.Headers["X-Custom"][0] != "value" {
		t.Errorf("Headers[X-Custom] = %v, want [value]", got.Headers["X-Custom"])
	}
	if got.Cookies["session"] != "abc" {
		t.Errorf("Cookies[session] = %q, want abc", got.Cookies["session"])
	}
	if got.Body != "hello" {
		t.Errorf("Body = %q, want hello", got.Body)
	}
	if got.OS.Hostname != "test-host" {
		t.Errorf("OS.Hostname = %q, want test-host", got.OS.Hostname)
	}
}

func TestBuildProtocolUntrustedPeer(t *testing.T) {
	// Without a trusted proxy, X-Forwarded-Proto is attacker-controlled and
	// ignored, matching the X-Forwarded-For trust rule.
	req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	if got := Build(req, nil, false, Options{Now: time.Unix(0, 0)}); got.Protocol != "http" {
		t.Errorf("Protocol = %q, want http (XFP from untrusted peer)", got.Protocol)
	}
}

func TestBuildJSONBody(t *testing.T) {
	body := []byte(`{"x":1}`)
	req := httptest.NewRequest(http.MethodPost, "http://x/", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	got := Build(req, body, false, Options{Now: time.Unix(0, 0)})

	if got.JSON == nil {
		t.Fatal("JSON = nil, want raw message for a JSON body")
	}
	if string(got.JSON) != `{"x":1}` {
		t.Errorf("JSON = %s, want {\"x\":1}", got.JSON)
	}
	if got.Protocol != "http" {
		t.Errorf("Protocol = %q, want http", got.Protocol)
	}
}

func TestBuildNonJSONBodyHasNoJSONField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://x/", nil)
	req.Header.Set("Content-Type", "text/plain")
	got := Build(req, []byte("not json"), false, Options{Now: time.Unix(0, 0)})
	if got.JSON != nil {
		t.Errorf("JSON = %s, want nil for non-JSON body", got.JSON)
	}
}

func TestBuildKubernetes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://x/", nil)

	if got := Build(req, nil, false, Options{Now: time.Unix(0, 0)}); got.Kubernetes != nil {
		t.Errorf("Kubernetes = %+v, want nil when not provided", got.Kubernetes)
	}

	k8s := &KubernetesInfo{PodName: "echo-abc", PodNamespace: "obs", PodIP: "10.0.0.1", NodeName: "node-1"}
	got := Build(req, nil, false, Options{Now: time.Unix(0, 0), Kubernetes: k8s})
	if got.Kubernetes == nil || got.Kubernetes.PodName != "echo-abc" || got.Kubernetes.NodeName != "node-1" {
		t.Errorf("Kubernetes = %+v, want populated", got.Kubernetes)
	}
}

func TestReadBody(t *testing.T) {
	t.Run("under limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://x/", strings.NewReader("small"))
		body, truncated, err := ReadBody(req, 1024)
		if err != nil {
			t.Fatalf("ReadBody: %v", err)
		}
		if truncated {
			t.Error("truncated = true, want false")
		}
		if string(body) != "small" {
			t.Errorf("body = %q, want small", body)
		}
	})

	t.Run("truncated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://x/", strings.NewReader("0123456789"))
		body, truncated, err := ReadBody(req, 4)
		if err != nil {
			t.Fatalf("ReadBody: %v", err)
		}
		if !truncated {
			t.Error("truncated = false, want true")
		}
		if string(body) != "0123" {
			t.Errorf("body = %q, want 0123", body)
		}
	})

	t.Run("zero max reads nothing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://x/", strings.NewReader("data"))
		body, _, err := ReadBody(req, 0)
		if err != nil {
			t.Fatalf("ReadBody: %v", err)
		}
		if len(body) != 0 {
			t.Errorf("body = %q, want empty", body)
		}
	})
}

func TestClientIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}

	t.Run("no trusted proxies uses peer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
		req.RemoteAddr = "203.0.113.7:5555"
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		ip, chain := clientIP(req, nil)
		if ip != "203.0.113.7" {
			t.Errorf("ip = %q, want 203.0.113.7", ip)
		}
		if len(chain) != 1 || chain[0] != "1.2.3.4" {
			t.Errorf("chain = %v, want [1.2.3.4]", chain)
		}
	})

	t.Run("trusted peer resolves client from XFF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
		req.RemoteAddr = "10.0.0.5:1234"
		req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.9")
		ip, _ := clientIP(req, trusted)
		if ip != "1.2.3.4" {
			t.Errorf("ip = %q, want 1.2.3.4", ip)
		}
	})

	t.Run("untrusted peer ignores XFF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
		req.RemoteAddr = "203.0.113.7:1234"
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		ip, _ := clientIP(req, trusted)
		if ip != "203.0.113.7" {
			t.Errorf("ip = %q, want 203.0.113.7", ip)
		}
	})
}

func TestIsJSON(t *testing.T) {
	tests := map[string]bool{
		"application/json":                true,
		"application/json; charset=utf-8": true,
		"application/vnd.api+json":        true,
		"text/plain":                      false,
		"":                                false,
	}
	for ct, want := range tests {
		if got := isJSON(ct); got != want {
			t.Errorf("isJSON(%q) = %v, want %v", ct, got, want)
		}
	}
}

func TestReadBodyNilBody(t *testing.T) {
	body, truncated, err := ReadBody(&http.Request{}, 16)
	if err != nil {
		t.Fatalf("ReadBody: %v", err)
	}
	if body != nil || truncated {
		t.Errorf("ReadBody(nil body) = (%q, %v), want (nil, false)", body, truncated)
	}
}

func TestClientIPPeerWithoutPort(t *testing.T) {
	// RemoteAddr without a port (no proxy in front) must pass through intact.
	req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	req.RemoteAddr = "203.0.113.7"
	if ip, _ := clientIP(req, nil); ip != "203.0.113.7" {
		t.Errorf("ip = %q, want 203.0.113.7", ip)
	}
}

func TestClientIPNonIPChainEntry(t *testing.T) {
	// A chain entry that isn't an IP can never be trusted, so it resolves as
	// the right-most untrusted hop.
	req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	req.Header.Set("X-Forwarded-For", "garbage")
	ip, _ := clientIP(req, []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})
	if ip != "garbage" {
		t.Errorf("ip = %q, want the untrusted literal entry", ip)
	}
}

// errReader fails immediately, exercising ReadBody's error path.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("broken") }

func TestReadBodyError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://x/", errReader{})
	if _, _, err := ReadBody(req, 16); err == nil {
		t.Error("ReadBody = nil error, want the reader's failure")
	}
}
