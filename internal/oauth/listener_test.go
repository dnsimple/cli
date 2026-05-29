package oauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoopbackHappyPath(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	go func() {
		resp, err := http.Get(lb.redirectURL + "?code=abc&state=expected-state")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	code, state, err := lb.await(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "abc", code)
	assert.Equal(t, "expected-state", state)
}

func TestLoopbackReturnsAuthErrorWhenErrorParamPresent(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	go func() {
		u := lb.redirectURL + "?" + url.Values{
			"error":             {"access_denied"},
			"error_description": {"user said no"},
			"state":             {"expected-state"},
		}.Encode()
		resp, err := http.Get(u)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err = lb.await(ctx)
	var ae *AuthError
	if !assert.ErrorAs(t, err, &ae) {
		return
	}
	assert.Equal(t, "access_denied", ae.Code)
	assert.Equal(t, "user said no", ae.Description)
}

func TestLoopbackRejectsStateMismatch(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	go func() {
		resp, err := http.Get(lb.redirectURL + "?code=abc&state=wrong-state")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err = lb.await(ctx)
	assert.ErrorIs(t, err, ErrStateMismatch)
}

// TestLoopbackRejectsForgedErrorWithoutState pins the fix for the CSRF
// vulnerability where a forged ?error=... callback (with no state, or a
// wrong state) was accepted because the error branch ran before the
// state check. State must be validated first, even on the error path.
func TestLoopbackRejectsForgedErrorWithoutState(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{name: "no state", query: "?error=access_denied&error_description=spoof"},
		{name: "wrong state", query: "?error=access_denied&error_description=spoof&state=forged"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lb, err := startLoopback("expected-state")
			if !assert.NoError(t, err) {
				return
			}
			defer lb.close()

			go func() {
				resp, err := http.Get(lb.redirectURL + tc.query)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _, err = lb.await(ctx)
			assert.ErrorIs(t, err, ErrStateMismatch)

			// And specifically, the AuthError path must NOT be reached
			// for these forged inputs.
			var ae *AuthError
			assert.False(t, errors.As(err, &ae), "forged error must not surface as AuthError")
		})
	}
}

// TestLoopbackSurfacesServeErrors verifies that a Serve failure (other
// than the normal ErrServerClosed) is delivered through the result
// channel instead of being silently dropped, so callers see the real
// cause rather than timing out at the outer 5-minute deadline.
func TestLoopbackSurfacesServeErrors(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}

	// Close the listener out from under Serve. Serve's Accept loop
	// returns the wrapped "use of closed network connection" error,
	// which is NOT http.ErrServerClosed.
	_ = lb.listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err = lb.await(ctx)
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "loopback listener died")
}

func TestLoopbackTimesOutWhenNoCallback(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, _, err = lb.await(ctx)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

// TestLoopbackConfiguresHTTPServerTimeouts pins the slowloris-defense
// configuration on the embedded http.Server. If a future refactor drops
// any of these timeouts, a stalled local connection can hold the
// listener open and block Shutdown until the outer 5-minute flow
// deadline fires.
func TestLoopbackConfiguresHTTPServerTimeouts(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	assert.Greater(t, lb.server.ReadHeaderTimeout, time.Duration(0), "ReadHeaderTimeout must be set")
	assert.Greater(t, lb.server.ReadTimeout, time.Duration(0), "ReadTimeout must be set")
	assert.Greater(t, lb.server.WriteTimeout, time.Duration(0), "WriteTimeout must be set")
}

func TestLoopbackIgnoresNonCallbackPaths(t *testing.T) {
	lb, err := startLoopback("expected-state")
	if !assert.NoError(t, err) {
		return
	}
	defer lb.close()

	// A probe to /, /favicon.ico, /anything must not consume the single-shot
	// result. The actual callback fired afterwards should be delivered.
	for _, path := range []string{"/", "/favicon.ico", "/probe"} {
		base := strings.TrimSuffix(lb.redirectURL, "/callback")
		resp, err := http.Get(base + path)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	go func() {
		resp, err := http.Get(lb.redirectURL + "?code=abc&state=expected-state")
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	code, _, err := lb.await(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "abc", code)
}
