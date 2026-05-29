package oauth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Loopback server timeouts. The loopback only ever serves one fast GET
// from the local browser, so the limits are deliberately tight: a stalled
// connection (security scanner, slowloris-style local probe, browser
// pre-fetcher that opens but never sends) would otherwise pin the
// listener until the outer 5-minute flow deadline fires, and would block
// Shutdown indefinitely on close.
const (
	loopbackReadHeaderTimeout = 5 * time.Second
	loopbackReadTimeout       = 10 * time.Second
	loopbackWriteTimeout      = 10 * time.Second
	loopbackShutdownTimeout   = 5 * time.Second
)

// callbackResult is the outcome of the OAuth redirect: either a code + state
// or an authorization-server error. State is included so the caller (Client)
// can perform a defense-in-depth check, even though the listener also
// validates it before sending the result.
type callbackResult struct {
	code  string
	state string
	err   error
}

// loopback owns the local HTTP listener that catches the OAuth redirect.
// One loopback corresponds to one login attempt: it accepts a single valid
// callback and is then shut down.
type loopback struct {
	listener      net.Listener
	server        *http.Server
	port          int
	redirectURL   string
	expectedState string

	result chan callbackResult
	once   sync.Once
}

// startLoopback binds 127.0.0.1 on an OS-assigned port and serves the OAuth
// redirect endpoint at /callback. The handler is single-shot: the first
// valid callback (or the first error) is delivered to the result channel,
// and subsequent requests are ignored.
//
// The expectedState parameter is the state value we sent on the authorize
// step; it is validated against the callback's `state` query parameter.
// A mismatch is reported via the result channel as ErrStateMismatch.
func startLoopback(expectedState string) (*loopback, error) {
	// 127.0.0.1 (not 0.0.0.0, not "localhost"). "localhost" can resolve to
	// ::1 on systems that prefer IPv6, which would break the redirect
	// matcher on the server side: the registered URI is the IPv4 literal.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("bind loopback listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	lb := &loopback{
		listener:      listener,
		port:          port,
		redirectURL:   fmt.Sprintf("http://127.0.0.1:%d/callback", port),
		expectedState: expectedState,
		result:        make(chan callbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", lb.handleCallback)
	// Any other path is not part of the OAuth dance. Don't reveal that this
	// is a CLI listener; just 404.
	mux.HandleFunc("/", http.NotFound)

	lb.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: loopbackReadHeaderTimeout,
		ReadTimeout:       loopbackReadTimeout,
		WriteTimeout:      loopbackWriteTimeout,
	}
	go func() {
		// Serve returns http.ErrServerClosed on shutdown, which is the
		// normal exit path. Anything else means the listener died before
		// we got a callback (e.g., FD exhaustion, the listener was
		// closed externally) -- surface that via the result channel so
		// await fails fast instead of timing out at the outer deadline.
		// deliver() is single-shot, so if a real callback was already
		// delivered this is a no-op.
		if err := lb.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lb.deliver(callbackResult{err: fmt.Errorf("oauth: loopback listener died: %w", err)})
		}
	}()

	return lb, nil
}

// await blocks until the callback fires or ctx is cancelled / times out.
func (l *loopback) await(ctx context.Context) (string, string, error) {
	select {
	case r := <-l.result:
		return r.code, r.state, r.err
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
}

// close shuts down the listener. Safe to call multiple times.
func (l *loopback) close() {
	// Bounded shutdown deadline so a hung in-flight handler cannot pin
	// the CLI after the OAuth flow has otherwise completed (e.g. a local
	// connection that never sends bytes -- the request-level read
	// timeouts cover the read side, but Shutdown still waits for those
	// timeouts to fire and we don't want to inherit their full duration
	// when we already have what we need).
	ctx, cancel := context.WithTimeout(context.Background(), loopbackShutdownTimeout)
	defer cancel()
	_ = l.server.Shutdown(ctx)
}

// handleCallback parses the OAuth redirect query, renders a small HTML
// status page so the browser tab is not left blank, and pushes exactly one
// result into the result channel.
func (l *loopback) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	q := r.URL.Query()
	errCode := q.Get("error")
	errDesc := q.Get("error_description")
	code := q.Get("code")
	state := q.Get("state")

	// Validate state before branching on anything else, including the
	// `error` response shape. Otherwise a forged callback (e.g. an
	// <img src> on a page the user is also viewing) can deliver an
	// attacker-chosen ?error=... result to the in-flight login without
	// holding a valid state token -- exactly the attack the state
	// parameter exists to defend against (RFC 6749 §10.12, OAuth 2.0
	// Security BCP §4.7).
	if state != l.expectedState {
		renderError(w, "state_mismatch", "The login flow could not be verified. Please run `dnsimple auth login` again.")
		l.deliver(callbackResult{err: ErrStateMismatch})
		return
	}

	switch {
	case errCode != "":
		renderError(w, errCode, errDesc)
		l.deliver(callbackResult{err: newAuthError(errCode, errDesc)})
	case code == "":
		renderError(w, "invalid_callback", "The authorization response did not include a code.")
		l.deliver(callbackResult{err: fmt.Errorf("oauth: callback missing code")})
	default:
		renderSuccess(w)
		l.deliver(callbackResult{code: code, state: state})
	}
}

// deliver sends a result exactly once, ignoring duplicates if the browser
// fires multiple requests (e.g. through a prefetcher).
func (l *loopback) deliver(r callbackResult) {
	l.once.Do(func() { l.result <- r })
}

// renderSuccess writes a minimal HTML page telling the user the CLI now has
// what it needs and the tab can be closed.
func renderSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, successHTML)
}

// renderError writes a minimal HTML error page including the OAuth error
// code and description from the authorization server.
func renderError(w http.ResponseWriter, code, description string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprintf(w, errorHTMLFmt, htmlEscape(code), htmlEscape(description))
}

// htmlEscape is a tiny inline escaper to avoid a dependency on html/template
// for these two short pages. It covers the characters that matter when the
// content is OAuth error codes and free-form descriptions.
func htmlEscape(s string) string {
	r := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			r = append(r, []byte("&amp;")...)
		case '<':
			r = append(r, []byte("&lt;")...)
		case '>':
			r = append(r, []byte("&gt;")...)
		case '"':
			r = append(r, []byte("&quot;")...)
		case '\'':
			r = append(r, []byte("&#39;")...)
		default:
			r = append(r, s[i])
		}
	}
	return string(r)
}

const successHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>DNSimple CLI: signed in</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif; max-width: 32rem; margin: 4rem auto; padding: 0 1rem; color: #1a1a1a; }
  h1 { font-size: 1.25rem; }
  p { line-height: 1.5; }
</style>
</head>
<body>
<h1>You are signed in to the DNSimple CLI.</h1>
<p>You can close this tab and return to your terminal.</p>
</body>
</html>
`

const errorHTMLFmt = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>DNSimple CLI: login failed</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif; max-width: 32rem; margin: 4rem auto; padding: 0 1rem; color: #1a1a1a; }
  h1 { font-size: 1.25rem; }
  p { line-height: 1.5; }
  code { background: #f4f4f5; padding: 0.1em 0.3em; border-radius: 3px; }
</style>
</head>
<body>
<h1>DNSimple CLI login did not complete.</h1>
<p>Error: <code>%s</code></p>
<p>%s</p>
<p>Return to your terminal for further details and try again.</p>
</body>
</html>
`
