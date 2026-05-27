package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// defaultDeadline is how long the entire login flow has to complete: open
// the browser, wait for the user to consent, exchange the code. Five
// minutes is what gh and similar CLIs use; it covers slow networks and
// users who briefly tab away without leaving stale listeners around.
const defaultDeadline = 5 * time.Minute

// Client runs the full interactive OAuth dance and returns an access token.
// All external dependencies are fields so tests can swap them out without
// real network or browser calls.
type Client struct {
	// ClientID is the OAuth public-client identifier registered with the
	// DNSimple application. Required.
	ClientID string

	// AuthorizeBase is the authorize endpoint URL, e.g.
	// "https://dnsimple.com/oauth/authorize". Required.
	AuthorizeBase string

	// TokenURL is the token endpoint URL, e.g.
	// "https://api.dnsimple.com/v2/oauth/access_token". Required.
	TokenURL string

	// BrowserOpener navigates the user's default browser to the given URL.
	// Production callers pass browser.OpenURL; tests pass a stub that fires
	// the callback synchronously. If nil, Login fails before binding the
	// listener.
	BrowserOpener func(string) error

	// HTTPClient is used for the token-exchange POST. Nil means
	// http.DefaultClient.
	HTTPClient *http.Client

	// Stderr is where user-visible status lines and the fallback authorize
	// URL are printed. Nil means io.Discard, which is appropriate for
	// embedding in tests that do not care about prompts.
	Stderr io.Writer

	// Deadline overrides defaultDeadline when set. Used by tests to fail
	// fast.
	Deadline time.Duration
}

// Login performs the full flow. On success the returned token is a bearer
// access token usable against the DNSimple API the same way as a manually
// pasted API token.
func (c *Client) Login(ctx context.Context) (string, error) {
	if c.ClientID == "" {
		return "", ErrNotProvisioned
	}
	if c.BrowserOpener == nil {
		return "", fmt.Errorf("oauth: BrowserOpener is required")
	}

	deadline := c.Deadline
	if deadline == 0 {
		deadline = defaultDeadline
	}
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	verifier, err := newVerifier()
	if err != nil {
		return "", err
	}
	chal := challenge(verifier)

	state, err := newState()
	if err != nil {
		return "", err
	}

	lb, err := startLoopback(state)
	if err != nil {
		return "", err
	}
	defer lb.close()

	authURL, err := buildAuthorizeURL(c.AuthorizeBase, c.ClientID, state, chal, lb.redirectURL)
	if err != nil {
		return "", err
	}

	stderr := c.stderr()
	if openErr := c.BrowserOpener(authURL); openErr != nil {
		fmt.Fprintln(stderr, "Could not open a browser automatically.")
		fmt.Fprintln(stderr, "Open this URL to continue:")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "  "+authURL)
		fmt.Fprintln(stderr)
	} else {
		fmt.Fprintln(stderr, "Opening your browser to authorize the DNSimple CLI...")
		fmt.Fprintln(stderr, "Waiting for the authorization to complete in the browser.")
	}

	code, _, err := lb.await(ctx)
	if err != nil {
		// Surface a clear deadline message instead of "context deadline
		// exceeded" to the user.
		if ctx.Err() != nil {
			return "", fmt.Errorf("timed out waiting for authorization after %s", deadline)
		}
		return "", err
	}

	token, err := c.exchangeCode(ctx, code, verifier, lb.redirectURL)
	if err != nil {
		return "", err
	}
	return token, nil
}

// buildAuthorizeURL assembles the authorize URL with all required PKCE +
// loopback parameters. Errors only on a malformed AuthorizeBase.
func buildAuthorizeURL(base, clientID, state, challenge, redirectURI string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse authorize base url: %w", err)
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("redirect_uri", redirectURI)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// tokenRequest is the JSON body of the POST to /v2/oauth/access_token for
// a public client. No client_secret; PKCE replaces it.
type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`
}

// tokenResponse is the success-path JSON from /v2/oauth/access_token. The
// CLI only consumes AccessToken; the rest is included so a future change
// (e.g. scoped tokens) can be picked up without re-shaping the struct.
type tokenResponse struct {
	AccessToken string  `json:"access_token"`
	TokenType   string  `json:"token_type"`
	Scope       *string `json:"scope"`
	AccountID   int64   `json:"account_id"`
}

// errorResponse is the RFC 6749 §5.2 error shape returned by the token
// endpoint on a 4xx.
type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// exchangeCode trades the authorization code + PKCE verifier for an access
// token. DNSimple's token endpoint accepts JSON, not form-encoded, which is
// why we hand-roll the request rather than using golang.org/x/oauth2.
func (c *Client) exchangeCode(ctx context.Context, code, verifier, redirectURI string) (string, error) {
	body, err := json.Marshal(tokenRequest{
		GrantType:    "authorization_code",
		ClientID:     c.ClientID,
		Code:         code,
		CodeVerifier: verifier,
		RedirectURI:  redirectURI,
	})
	if err != nil {
		return "", fmt.Errorf("encode token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token endpoint: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var er errorResponse
		if json.Unmarshal(respBody, &er) == nil && er.Error != "" {
			return "", &AuthError{Code: er.Error, Description: er.ErrorDescription}
		}
		return "", fmt.Errorf("token endpoint returned HTTP %d", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.Unmarshal(respBody, &tok); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("token response did not include an access_token")
	}
	return tok.AccessToken, nil
}

func (c *Client) stderr() io.Writer {
	if c.Stderr == nil {
		return io.Discard
	}
	return c.Stderr
}
