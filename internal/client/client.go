// Package client is a hand-written REST client for the Beszel (PocketBase) API.
// Resources in internal/provider never issue HTTP requests directly; they always
// go through the typed methods exposed here, one file per collection.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// defaultTimeout bounds every request so an unreachable or unresponsive hub
// fails a Terraform apply instead of hanging it indefinitely.
const defaultTimeout = 30 * time.Second

// Client is an authenticated client for a Beszel hub. It authenticates as a
// PocketBase superuser, since regular-user auth cannot manage systems/users
// arbitrarily under Beszel's default collection rules.
type Client struct {
	baseURL    string
	email      string
	password   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
}

// New creates a Client for the hub at baseURL, authenticating as the given
// superuser email/password on first request.
func New(baseURL, email, password string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		email:      email,
		password:   password,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

type authResponse struct {
	Token string `json:"token"`
}

// authenticate obtains a fresh superuser auth token.
func (c *Client) authenticate(ctx context.Context) (string, error) {
	body, err := json.Marshal(map[string]string{
		"identity": c.email,
		"password": c.password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/collections/_superusers/auth-with-password",
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("beszel: authentication request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", decodeAPIError(resp.StatusCode, respBody)
	}

	var auth authResponse
	if err := json.Unmarshal(respBody, &auth); err != nil {
		return "", fmt.Errorf("beszel: decoding auth response: %w", err)
	}

	return auth.Token, nil
}

func (c *Client) currentToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" {
		return c.token, nil
	}

	token, err := c.authenticate(ctx)
	if err != nil {
		return "", err
	}
	c.token = token
	return c.token, nil
}

func (c *Client) invalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
}

// do issues an authenticated request using the client's own superuser token.
// If body is non-nil, it is marshaled as JSON. If out is non-nil, the response
// body is unmarshaled into it. A single 401 triggers one re-authentication and
// retry, since PocketBase tokens expire.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(encoded)
	}

	token, err := c.currentToken(ctx)
	if err != nil {
		return err
	}

	resp, respBody, err := c.doOnce(ctx, token, method, path, query, bodyReader)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		c.invalidateToken()
		if bodyReader != nil {
			encoded, err := json.Marshal(body)
			if err != nil {
				return err
			}
			bodyReader = bytes.NewReader(encoded)
		}
		token, err = c.currentToken(ctx)
		if err != nil {
			return err
		}
		resp, respBody, err = c.doOnce(ctx, token, method, path, query, bodyReader)
		if err != nil {
			return err
		}
	}

	return decodeResponse(method, path, resp, respBody, out)
}

// doAs issues a single authenticated request using the given token directly,
// bypassing the client's cached superuser token and 401 retry. Used for
// endpoints that must be called as an impersonated regular user rather than
// the superuser (see ImpersonateUser).
func (c *Client) doAs(ctx context.Context, token, method, path string, query url.Values, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(encoded)
	}

	resp, respBody, err := c.doOnce(ctx, token, method, path, query, bodyReader)
	if err != nil {
		return err
	}

	return decodeResponse(method, path, resp, respBody, out)
}

func decodeResponse(method, path string, resp *http.Response, respBody []byte, out any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("beszel: decoding response from %s %s: %w", method, path, err)
		}
	}

	return nil
}

func (c *Client) doOnce(ctx context.Context, token, method, path string, query url.Values, body io.Reader) (*http.Response, []byte, error) {
	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("beszel: request to %s %s failed: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}

type impersonateResponse struct {
	Token string `json:"token"`
}

// ImpersonateUser obtains an auth token for the given user record via
// PocketBase's superuser impersonation route. Some Beszel endpoints (e.g. the
// universal token endpoint) are scoped to the caller's own record and
// explicitly forbid superuser callers, so the client has to act as the user
// itself for those.
func (c *Client) ImpersonateUser(ctx context.Context, userID string, duration time.Duration) (string, error) {
	var resp impersonateResponse
	path := fmt.Sprintf("/api/collections/users/impersonate/%s", userID)
	body := map[string]any{}
	if duration > 0 {
		body["duration"] = int(duration.Seconds())
	}
	if err := c.do(ctx, http.MethodPost, path, nil, body, &resp); err != nil {
		return "", fmt.Errorf("beszel: impersonating user %s: %w", userID, err)
	}
	return resp.Token, nil
}

func decodeAPIError(status int, body []byte) error {
	apiErr := &APIError{Status: status}
	if len(body) > 0 {
		// Best-effort decode; if the body isn't the expected envelope, fall
		// back to a generic message rather than losing the status code.
		_ = json.Unmarshal(body, apiErr)
	}
	apiErr.Status = status
	if apiErr.Message == "" {
		apiErr.Message = string(body)
	}
	return apiErr
}
