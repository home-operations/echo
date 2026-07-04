package echo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func parse(t *testing.T, target string, enabled bool, header http.Header) Commands {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if header != nil {
		req.Header = header
	}
	return ParseCommands(req, CommandOptions{Enabled: enabled, MaxDelay: time.Second})
}

func TestParseCommandsStatus(t *testing.T) {
	if c := parse(t, "/?echo-code=503", true, nil); c.Status != 503 {
		t.Errorf("Status = %d, want 503", c.Status)
	}
	// Surrounding whitespace is trimmed (%20 = space).
	if c := parse(t, "/?echo-code=%20503%20", true, nil); c.Status != 503 {
		t.Errorf("padded echo-code: Status = %d, want 503", c.Status)
	}
	// Out of range (including 1xx, which would double-respond) and non-numeric
	// are ignored.
	for _, code := range []string{"99", "100", "199", "600", "abc", "200.5"} {
		if c := parse(t, "/?echo-code="+code, true, nil); c.Status != 0 {
			t.Errorf("echo-code=%s: Status = %d, want 0 (ignored)", code, c.Status)
		}
	}
}

func TestParseCommandsDelay(t *testing.T) {
	if c := parse(t, "/?echo-delay=250ms", true, nil); c.Delay != 250*time.Millisecond {
		t.Errorf("Delay = %v, want 250ms", c.Delay)
	}
	// Clamped to MaxDelay (1s in the helper).
	if c := parse(t, "/?echo-delay=10s", true, nil); c.Delay != time.Second {
		t.Errorf("Delay = %v, want clamped to 1s", c.Delay)
	}
	// Negative and unparseable are ignored, and leave no applied entry.
	for _, d := range []string{"-5s", "soon", "0"} {
		c := parse(t, "/?echo-delay="+d, true, nil)
		if c.Delay != 0 {
			t.Errorf("echo-delay=%s: Delay = %v, want 0", d, c.Delay)
		}
		if c.Applied() != nil {
			t.Errorf("echo-delay=%s: Applied() = %+v, want nil", d, c.Applied())
		}
	}
}

func TestParseCommandsDelayMaxZero(t *testing.T) {
	// MaxDelay == 0 means "no delay allowed": any request is clamped to zero.
	req := httptest.NewRequest(http.MethodGet, "/?echo-delay=5s", nil)
	if c := ParseCommands(req, CommandOptions{Enabled: true, MaxDelay: 0}); c.Delay != 0 {
		t.Errorf("Delay = %v, want 0 when MaxDelay is 0", c.Delay)
	}
}

func TestParseCommandsHeaders(t *testing.T) {
	c := parse(t, "/?echo-header=X-One:1&echo-header=X-Two:two", true, nil)
	if c.Headers.Get("X-One") != "1" || c.Headers.Get("X-Two") != "two" {
		t.Errorf("Headers = %v, want X-One:1 and X-Two:two", c.Headers)
	}

	// Value may itself contain colons; only the first splits name from value.
	if c := parse(t, "/?echo-header=Location:http://x/y", true, nil); c.Headers.Get("Location") != "http://x/y" {
		t.Errorf("Location = %q, want http://x/y", c.Headers.Get("Location"))
	}

	// Malformed (no colon), header-injection (CRLF, NUL, other control bytes)
	// attempts are dropped.
	for _, h := range []string{"NoColon", "X-Bad\r\nInjected:1", "X-Ok:val\r\nSet-Cookie:x", "X-Nul:va\x00lue", "X-Tab:a\tb"} {
		c := parse(t, "/", true, http.Header{"X-Echo-Header": {h}})
		if len(c.Headers) != 0 {
			t.Errorf("echo-header=%q: Headers = %v, want none", h, c.Headers)
		}
	}
}

func TestParseCommandsSetCookieDomainStripped(t *testing.T) {
	// Set-Cookie via echo-header keeps its attributes except Domain, which
	// would scope the cookie to a shared parent domain (session fixation).
	// Exercised via the header surface: Go's url.ParseQuery rejects the
	// semicolons a Set-Cookie value carries.
	c := parse(t, "/", true, http.Header{
		"X-Echo-Header": {"Set-Cookie:sess=x; Domain=example.com; Path=/; HttpOnly"},
	})
	got := c.Headers.Get("Set-Cookie")
	if got == "" {
		t.Fatal("Set-Cookie dropped entirely, want Domain-stripped value")
	}
	if strings.Contains(strings.ToLower(got), "domain=") {
		t.Errorf("Set-Cookie = %q, want Domain attribute stripped", got)
	}
	if !strings.Contains(got, "Path=/") || !strings.Contains(got, "HttpOnly") {
		t.Errorf("Set-Cookie = %q, want other attributes preserved", got)
	}

	// A malformed Set-Cookie value is dropped.
	c = parse(t, "/", true, http.Header{"X-Echo-Header": {"Set-Cookie:"}})
	if got := c.Headers.Get("Set-Cookie"); got != "" {
		t.Errorf("malformed Set-Cookie = %q, want dropped", got)
	}
}

func TestParseCommandsReservedHeaders(t *testing.T) {
	// echo's own contract/safety headers can't be set via echo-header, in any casing.
	for _, h := range []string{"Content-Type:text/html", "content-length:0", "X-Content-Type-Options:off", "cache-control:public"} {
		if c := parse(t, "/?echo-header="+h, true, nil); len(c.Headers) != 0 {
			t.Errorf("echo-header=%q: Headers = %v, want none (reserved)", h, c.Headers)
		}
	}
	// A non-reserved header alongside a reserved one still goes through.
	c := parse(t, "/?echo-header=Content-Type:text/html&echo-header=X-Test:1", true, nil)
	if c.Headers.Get("X-Test") != "1" || c.Headers.Get("Content-Type") != "" {
		t.Errorf("Headers = %v, want only X-Test:1", c.Headers)
	}
}

func TestParseCommandsHeadersMergeSources(t *testing.T) {
	c := parse(t, "/?echo-header=X-Query:q", true, http.Header{"X-Echo-Header": {"X-Header:h"}})
	if c.Headers.Get("X-Query") != "q" || c.Headers.Get("X-Header") != "h" {
		t.Errorf("Headers = %v, want both query and header sources merged", c.Headers)
	}
}

func TestParseCommandsCookies(t *testing.T) {
	c := parse(t, "/?echo-cookie=session:abc&echo-cookie=theme:dark", true, nil)
	if len(c.Cookies) != 2 {
		t.Fatalf("Cookies = %v, want 2", c.Cookies)
	}
	if c.Cookies[0].Name != "session" || c.Cookies[0].Value != "abc" {
		t.Errorf("Cookies[0] = %+v, want session=abc", c.Cookies[0])
	}
	// Applied lists the cookie names, sorted.
	if a := c.Applied(); a == nil || len(a.Cookies) != 2 || a.Cookies[0] != "session" || a.Cookies[1] != "theme" {
		t.Errorf("Applied.Cookies = %+v, want [session theme]", a)
	}

	// The header source works too.
	if c := parse(t, "/", true, http.Header{"X-Echo-Cookie": {"id:42"}}); len(c.Cookies) != 1 || c.Cookies[0].Value != "42" {
		t.Errorf("header cookie = %v, want id=42", c.Cookies)
	}

	// Malformed entries and cookies that fail http.Cookie validation are
	// dropped: no colon, an invalid name, control/CRLF or other invalid value
	// bytes (which would enable header injection), or an empty name.
	for _, e := range []string{"NoColon", "bad name:v", "x:has;semi", "x:a\r\nb", "x:nul\x00", ":noname"} {
		if c := parse(t, "/", true, http.Header{"X-Echo-Cookie": {e}}); len(c.Cookies) != 0 {
			t.Errorf("echo-cookie=%q: Cookies = %v, want none", e, c.Cookies)
		}
	}
}

func TestParseCommandsPretty(t *testing.T) {
	// Bare flag (presence) => true.
	if c := parse(t, "/?echo-pretty-print", true, nil); !c.Pretty {
		t.Error("bare echo-pretty-print should enable Pretty")
	}
	if c := parse(t, "/?echo-pretty-print=false", true, nil); c.Pretty {
		t.Error("echo-pretty-print=false should disable Pretty")
	}
	// Header form.
	if c := parse(t, "/", true, http.Header{"X-Echo-Pretty-Print": {"yes"}}); !c.Pretty {
		t.Error("X-Echo-Pretty-Print: yes should enable Pretty")
	}
	// Default carries through when unset.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if c := ParseCommands(req, CommandOptions{Enabled: true, DefaultPretty: true}); !c.Pretty {
		t.Error("DefaultPretty=true should carry through when the request is silent")
	}
}

func TestParseCommandsQueryWinsOverHeader(t *testing.T) {
	c := parse(t, "/?echo-code=503", true, http.Header{"X-Echo-Code": {"418"}})
	if c.Status != 503 {
		t.Errorf("Status = %d, want 503 (query wins over header)", c.Status)
	}
}

func TestParseCommandsDisabled(t *testing.T) {
	// Shaping directives are ignored when disabled...
	c := parse(t, "/?echo-code=503&echo-delay=1s&echo-header=X-T:1&echo-pretty-print", false, nil)
	if c.Status != 0 || c.Delay != 0 || c.Headers != nil {
		t.Errorf("shaping directives honored while disabled: %+v", c)
	}
	// ...but pretty-print is always parsed.
	if !c.Pretty {
		t.Error("Pretty should be honored even when commands are disabled")
	}
}

func TestAppliedSummary(t *testing.T) {
	if a := (Commands{}).Applied(); a != nil {
		t.Errorf("Applied() = %+v, want nil for an empty command set", a)
	}

	c := Commands{
		Status:  503,
		Delay:   2 * time.Second,
		Headers: http.Header{"X-Two": {"2"}, "X-One": {"1"}},
	}
	a := c.Applied()
	if a == nil {
		t.Fatal("Applied() = nil, want populated")
	}
	if a.Status != 503 || a.Delay != "2s" {
		t.Errorf("Applied = %+v, want status 503 / delay 2s", a)
	}
	// Header names are sorted for a stable response.
	if len(a.Headers) != 2 || a.Headers[0] != "X-One" || a.Headers[1] != "X-Two" {
		t.Errorf("Applied.Headers = %v, want sorted [X-One X-Two]", a.Headers)
	}
}
