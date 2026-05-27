package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// openerFollowingAuthorize returns a BrowserOpener stub that simulates the
// user reaching the authorize endpoint and consenting: it parses the
// authorize URL, extracts the redirect_uri and state, and fires the
// callback request with the supplied code asynchronously.
//
// recordedAuthURL captures the URL the Client tried to open so tests can
// assert on the query string.
func openerFollowingAuthorize(t *testing.T, code string, recordedAuthURL *string) func(string) error {
	t.Helper()
	return func(authURL string) error {
		if recordedAuthURL != nil {
			*recordedAuthURL = authURL
		}
		u, err := url.Parse(authURL)
		if err != nil {
			return err
		}
		q := u.Query()
		state := q.Get("state")
		redirectURI := q.Get("redirect_uri")

		go func() {
			resp, err := http.Get(redirectURI + "?" + url.Values{
				"code":  {code},
				"state": {state},
			}.Encode())
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
		return nil
	}
}

func TestClientLoginHappyPath(t *testing.T) {
	var receivedBody tokenRequest
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &receivedBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"the-token","token_type":"bearer","scope":null,"account_id":981}`)
	}))
	defer tokenServer.Close()

	var authURL string
	c := &Client{
		ClientID:      "client-abc",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      tokenServer.URL,
		BrowserOpener: openerFollowingAuthorize(t, "fake-code", &authURL),
		Stderr:        io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	token, err := c.Login(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "the-token", token)

	// Authorize URL must carry all PKCE + loopback parameters.
	u, _ := url.Parse(authURL)
	q := u.Query()
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "client-abc", q.Get("client_id"))
	assert.NotEmpty(t, q.Get("state"))
	assert.NotEmpty(t, q.Get("code_challenge"))
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.True(t, strings.HasPrefix(q.Get("redirect_uri"), "http://127.0.0.1:"))

	// Token request body must carry the verifier (not a client_secret) and
	// the same redirect_uri.
	assert.Equal(t, "authorization_code", receivedBody.GrantType)
	assert.Equal(t, "client-abc", receivedBody.ClientID)
	assert.Equal(t, "fake-code", receivedBody.Code)
	assert.NotEmpty(t, receivedBody.CodeVerifier)
	assert.Equal(t, q.Get("redirect_uri"), receivedBody.RedirectURI)
}

func TestClientLoginSurfacesTokenEndpointErrorResponse(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"bad verifier"}`)
	}))
	defer tokenServer.Close()

	c := &Client{
		ClientID:      "client-abc",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      tokenServer.URL,
		BrowserOpener: openerFollowingAuthorize(t, "fake-code", nil),
		Stderr:        io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.Login(ctx)
	var ae *AuthError
	if !assert.ErrorAs(t, err, &ae) {
		return
	}
	assert.Equal(t, "invalid_grant", ae.Code)
	assert.Equal(t, "bad verifier", ae.Description)
}

func TestClientLoginSurfacesAuthorizationDenied(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("token endpoint should not be called when authorize fails")
	}))
	defer tokenServer.Close()

	// Opener that simulates the user clicking Deny.
	opener := func(authURL string) error {
		u, _ := url.Parse(authURL)
		redirect := u.Query().Get("redirect_uri")
		state := u.Query().Get("state")
		go func() {
			resp, err := http.Get(redirect + "?" + url.Values{
				"error":             {"access_denied"},
				"error_description": {"user cancelled"},
				"state":             {state},
			}.Encode())
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
		return nil
	}

	c := &Client{
		ClientID:      "client-abc",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      tokenServer.URL,
		BrowserOpener: opener,
		Stderr:        io.Discard,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.Login(ctx)
	var ae *AuthError
	if !assert.ErrorAs(t, err, &ae) {
		return
	}
	assert.Equal(t, "access_denied", ae.Code)
}

func TestClientLoginRefusesEmptyClientID(t *testing.T) {
	c := &Client{
		ClientID:      "",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      "https://example.test/v2/oauth/access_token",
		BrowserOpener: func(string) error { return nil },
		Stderr:        io.Discard,
	}
	_, err := c.Login(context.Background())
	assert.ErrorIs(t, err, ErrNotProvisioned)
}

func TestClientLoginPrintsURLWhenBrowserCannotBeOpened(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"the-token","token_type":"bearer","scope":null,"account_id":981}`)
	}))
	defer tokenServer.Close()

	// Opener that errors. We still need a callback to fire, otherwise the
	// listener times out. Fire it from inside the opener even on error.
	opener := func(authURL string) error {
		u, _ := url.Parse(authURL)
		redirect := u.Query().Get("redirect_uri")
		state := u.Query().Get("state")
		go func() {
			resp, _ := http.Get(redirect + "?" + url.Values{
				"code":  {"fake-code"},
				"state": {state},
			}.Encode())
			if resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
		return fmt.Errorf("could not launch browser")
	}

	var stderr bytes.Buffer
	c := &Client{
		ClientID:      "client-abc",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      tokenServer.URL,
		BrowserOpener: opener,
		Stderr:        &stderr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	token, err := c.Login(ctx)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "the-token", token)
	assert.Contains(t, stderr.String(), "Could not open a browser")
	assert.Contains(t, stderr.String(), "https://example.test/oauth/authorize")
}

func TestClientLoginTimesOutWhenNoCallback(t *testing.T) {
	c := &Client{
		ClientID:      "client-abc",
		AuthorizeBase: "https://example.test/oauth/authorize",
		TokenURL:      "https://example.test/v2/oauth/access_token",
		// Opener that does nothing: no callback ever fires.
		BrowserOpener: func(string) error { return nil },
		Stderr:        io.Discard,
		Deadline:      150 * time.Millisecond,
	}
	_, err := c.Login(context.Background())
	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "timed out")
}
