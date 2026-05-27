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
