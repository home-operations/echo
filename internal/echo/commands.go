package echo

import (
	"maps"
	"net/http"
	"net/textproto"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Caller directives are read from two parallel surfaces: an echo-* query
// parameter and the matching X-Echo-* request header. Query parameters use
// dashes (echo-code) and headers follow HTTP convention (X-Echo-Code) — dashes
// throughout, which also avoids ingress-nginx silently dropping underscored
// header names.
const (
	queryPrefix  = "echo-"
	headerPrefix = "X-Echo-"

	directiveCode   = "code"
	directiveDelay  = "delay"
	directiveHeader = "header"
	directiveCookie = "cookie"
	directivePretty = "pretty-print"
)

// CommandOptions controls how response directives are parsed from a request.
type CommandOptions struct {
	// Enabled turns on response shaping (status, delay, headers). Pretty-print
	// is parsed regardless, since it only affects formatting and cannot be
	// abused.
	Enabled bool
	// MaxDelay caps the delay a caller may request; larger values are clamped.
	MaxDelay time.Duration
	// DefaultPretty is the baseline pretty-print setting used when the request
	// does not carry an echo-pretty-print directive.
	DefaultPretty bool
}

// Commands are the response directives echo applies for a single request,
// parsed from echo-* query parameters and X-Echo-* request headers.
type Commands struct {
	// Status overrides the response status code; 0 leaves the default.
	Status int
	// Delay is how long to wait before responding; 0 means no delay.
	Delay time.Duration
	// Headers are extra response headers the caller asked echo to set.
	Headers http.Header
	// Cookies are response cookies the caller asked echo to set.
	Cookies []*http.Cookie
	// Pretty indents the JSON response.
	Pretty bool
}

// Applied is the echo-directives summary added to the response document so a
// caller can confirm which response-shaping directives echo honored — useful
// for spotting directives that were ignored because commands are disabled or a
// value was out of range. It is omitted when no shaping directive took effect.
type Applied struct {
	Status  int      `json:"status,omitempty"`
	Delay   string   `json:"delay,omitempty"`
	Headers []string `json:"headers,omitempty"`
	Cookies []string `json:"cookies,omitempty"`
}

// ParseCommands extracts the caller's response directives from the request.
// Pretty-print is always parsed; the response-shaping directives (status,
// delay, headers) are parsed only when opts.Enabled is set.
func ParseCommands(r *http.Request, opts CommandOptions) Commands {
	q := r.URL.Query()
	c := Commands{Pretty: opts.DefaultPretty}

	if v, ok := boolDirective(q, r.Header, directivePretty); ok {
		c.Pretty = v
	}

	if !opts.Enabled {
		return c
	}

	// 1xx is excluded: WriteHeader(1xx) sends an informational response and the
	// JSON body then triggers an implicit 200, so the client would see two
	// status lines while metrics and the applied summary report the first.
	if raw, ok := directive(q, r.Header, directiveCode); ok {
		if code, err := strconv.Atoi(raw); err == nil && code >= 200 && code <= 599 {
			c.Status = code
		}
	}

	if raw, ok := directive(q, r.Header, directiveDelay); ok {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			// Clamp to the cap. MaxDelay == 0 therefore means "no delay
			// allowed" (clamped to zero), not "unbounded".
			if d > opts.MaxDelay {
				d = opts.MaxDelay
			}
			c.Delay = d
		}
	}

	c.Headers = parseHeaders(q, r.Header)
	c.Cookies = parseCookies(q, r.Header)

	return c
}

// Applied summarizes the response-shaping directives that took effect, for
// inclusion in the response document. It returns nil when none applied.
func (c Commands) Applied() *Applied {
	a := &Applied{Status: c.Status}
	if c.Delay > 0 {
		a.Delay = c.Delay.String()
	}
	if len(c.Headers) > 0 {
		a.Headers = slices.Sorted(maps.Keys(c.Headers))
	}
	if len(c.Cookies) > 0 {
		a.Cookies = make([]string, 0, len(c.Cookies))
		for _, ck := range c.Cookies {
			a.Cookies = append(a.Cookies, ck.Name)
		}
		slices.Sort(a.Cookies)
	}
	if a.Status == 0 && a.Delay == "" && len(a.Headers) == 0 && len(a.Cookies) == 0 {
		return nil
	}
	return a
}

// directive returns the raw value of a single-valued directive. The query
// parameter wins over the header when both are present.
func directive(q url.Values, h http.Header, name string) (string, bool) {
	if q.Has(queryPrefix + name) {
		return strings.TrimSpace(q.Get(queryPrefix + name)), true
	}
	if v, ok := firstHeader(h, name); ok {
		return strings.TrimSpace(v), true
	}
	return "", false
}

// boolDirective reads a flag directive. A bare query parameter (no value) means
// true, otherwise the value is parsed loosely. The query parameter wins over
// the header when both are present.
func boolDirective(q url.Values, h http.Header, name string) (bool, bool) {
	if q.Has(queryPrefix + name) {
		return parseFlag(q.Get(queryPrefix + name)), true
	}
	if v, ok := firstHeader(h, name); ok {
		return parseFlag(v), true
	}
	return false, false
}

// firstHeader returns the first value of the X-Echo-<name> header, reporting
// whether the header was present (even with an empty value, so presence-only
// flags still register).
func firstHeader(h http.Header, name string) (string, bool) {
	key := textproto.CanonicalMIMEHeaderKey(headerPrefix + name)
	if vs, ok := h[key]; ok && len(vs) > 0 {
		return vs[0], true
	}
	return "", false
}

// rawDirectives collects every value of a repeatable directive from both
// surfaces: the echo-<name> query parameters and the X-Echo-<name> headers.
func rawDirectives(q url.Values, h http.Header, name string) []string {
	return slices.Concat(q[queryPrefix+name], h[textproto.CanonicalMIMEHeaderKey(headerPrefix+name)])
}

// parseFlag interprets a flag value; an empty value means true (so a bare
// ?echo-pretty-print enables it), otherwise it accepts the usual truthy spellings.
func parseFlag(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "1", "t", "true", "yes", "on", "y":
		return true
	default:
		return false
	}
}

// reservedHeaders are response headers a caller may not set via echo-header:
// they carry echo's JSON contract (Content-Type, Content-Length) and its
// browser-inert / no-store guarantees (X-Content-Type-Options, Cache-Control).
// Dropping them keeps the response coherent and the applied summary honest;
// every other header remains settable.
var reservedHeaders = map[string]struct{}{
	"Content-Type":           {},
	"Content-Length":         {},
	"X-Content-Type-Options": {},
	"Cache-Control":          {},
}

// parseHeaders collects custom response headers from every echo-header query
// parameter and X-Echo-Header request header. Each entry is "Name: Value";
// entries that are malformed, carry header-injection bytes, or target a
// reserved header are dropped. Returns nil when none remain.
func parseHeaders(q url.Values, h http.Header) http.Header {
	raw := rawDirectives(q, h, directiveHeader)
	if len(raw) == 0 {
		return nil
	}
	out := http.Header{}
	for _, entry := range raw {
		name, value, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}
		// Canonicalize before validating so the reserved-header lookup matches
		// regardless of the caller's casing; an invalid name is returned
		// unchanged and rejected below.
		name = textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(name))
		value = strings.TrimSpace(value)
		if !validHeaderName(name) || !validHeaderValue(value) {
			continue
		}
		if _, reserved := reservedHeaders[name]; reserved {
			continue
		}
		if name == "Set-Cookie" {
			value, ok = sanitizeSetCookie(value)
			if !ok {
				continue
			}
		}
		out.Add(name, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// sanitizeSetCookie re-renders a caller-supplied Set-Cookie value with the
// Domain attribute removed. A Domain-scoped cookie set from echo's hostname
// would also reach every sibling app under a shared parent domain (session
// fixation); every other attribute only affects echo's own origin. Malformed
// values are rejected.
func sanitizeSetCookie(value string) (string, bool) {
	ck, err := http.ParseSetCookie(value)
	if err != nil {
		return "", false
	}
	ck.Domain = ""
	return ck.String(), true
}

// validHeaderName accepts a non-empty field name free of spaces, control
// characters, and the name/value separator.
func validHeaderName(s string) bool {
	return s != "" && !strings.ContainsFunc(s, func(r rune) bool {
		return r <= ' ' || r == 0x7f || r == ':'
	})
}

// validHeaderValue rejects every control character, including the CR/LF that
// would enable response splitting. net/http does not validate header values
// when writing the response, so this is echo's own guard.
func validHeaderValue(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return r < ' ' || r == 0x7f
	})
}

// parseCookies collects response cookies from every echo-cookie query parameter
// and X-Echo-Cookie request header. Each entry is "name:value" (a bare cookie,
// no attributes — use echo-header=Set-Cookie:... for those); entries that are
// malformed or fail http.Cookie validation are dropped. Returns nil when none
// are valid.
func parseCookies(q url.Values, h http.Header) []*http.Cookie {
	raw := rawDirectives(q, h, directiveCookie)
	if len(raw) == 0 {
		return nil
	}
	var out []*http.Cookie
	for _, entry := range raw {
		name, value, ok := strings.Cut(entry, ":")
		if !ok {
			continue
		}
		ck := &http.Cookie{Name: strings.TrimSpace(name), Value: strings.TrimSpace(value)}
		if ck.Valid() != nil {
			continue
		}
		out = append(out, ck)
	}
	return out
}
