package client

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// impersonationDuration bounds how long the token minted to act as the
// target user is valid. The universal-token call it's used for completes
// immediately, so this only needs to outlive a single request.
const impersonationDuration = time.Minute

// UniversalToken is a per-user agent registration token. Beszel scopes the
// underlying endpoint to the authenticated caller's own record, and superusers
// are explicitly forbidden from calling it, so every method here impersonates
// the target user (via Client.ImpersonateUser) before calling Beszel's custom
// /api/beszel/universal-token endpoint on their behalf.
type UniversalToken struct {
	User      string
	Token     string
	Active    bool
	Permanent bool
}

type universalTokenResponse struct {
	Token     string `json:"token"`
	Active    bool   `json:"active"`
	Permanent bool   `json:"permanent"`
}

const universalTokenPath = "/api/beszel/universal-token"

// GetUniversalToken returns the current universal token state for userID.
func (c *Client) GetUniversalToken(ctx context.Context, userID string) (*UniversalToken, error) {
	return c.callUniversalToken(ctx, userID, nil)
}

// EnableUniversalToken creates or activates a universal token for userID. If
// permanent is true the token persists across hub restarts; otherwise it
// expires after an hour of hub uptime. If token is non-empty, it is used as
// the token value instead of letting the hub generate one.
func (c *Client) EnableUniversalToken(ctx context.Context, userID, token string, permanent bool) (*UniversalToken, error) {
	query := url.Values{"enable": {"1"}}
	if permanent {
		query.Set("permanent", "1")
	}
	if token != "" {
		query.Set("token", token)
	}
	return c.callUniversalToken(ctx, userID, query)
}

// DisableUniversalToken deactivates and deletes userID's universal token
// (ephemeral or permanent).
func (c *Client) DisableUniversalToken(ctx context.Context, userID string) (*UniversalToken, error) {
	return c.callUniversalToken(ctx, userID, url.Values{"enable": {"0"}})
}

func (c *Client) callUniversalToken(ctx context.Context, userID string, query url.Values) (*UniversalToken, error) {
	token, err := c.ImpersonateUser(ctx, userID, impersonationDuration)
	if err != nil {
		return nil, err
	}

	var resp universalTokenResponse
	if err := c.doAs(ctx, token, http.MethodGet, universalTokenPath, query, nil, &resp); err != nil {
		return nil, err
	}

	return &UniversalToken{
		User:      userID,
		Token:     resp.Token,
		Active:    resp.Active,
		Permanent: resp.Permanent,
	}, nil
}
