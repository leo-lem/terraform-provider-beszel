package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// QuietHours is a Beszel "quiet_hours" collection record: a window during
// which alert notifications are muted for a user, either globally (empty
// System) or for one system. Start and End are PocketBase date strings
// ("2006-01-02 15:04:05.000Z").
type QuietHours struct {
	ID      string `json:"id"`
	User    string `json:"user"`
	System  string `json:"system"`
	Type    string `json:"type"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// QuietHoursWrite holds the fields accepted when creating or updating a
// quiet hours window. System is always sent: an empty string clears it, which
// makes the window apply to all systems.
type QuietHoursWrite struct {
	User   string `json:"user,omitempty"`
	System string `json:"system"`
	Type   string `json:"type"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

type quietHoursList struct {
	Items []QuietHours `json:"items"`
}

const quietHoursPath = "/api/collections/quiet_hours/records"

// ListQuietHours returns quiet hours records matching the given PocketBase filter expression.
func (c *Client) ListQuietHours(ctx context.Context, filter string) ([]QuietHours, error) {
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	query.Set("perPage", "500")

	var list quietHoursList
	if err := c.do(ctx, http.MethodGet, quietHoursPath, query, nil, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetQuietHours fetches a single quiet hours record by its PocketBase record ID.
func (c *Client) GetQuietHours(ctx context.Context, id string) (*QuietHours, error) {
	var qh QuietHours
	path := fmt.Sprintf("%s/%s", quietHoursPath, id)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &qh); err != nil {
		return nil, err
	}
	return &qh, nil
}

// CreateQuietHours creates a new quiet hours record.
func (c *Client) CreateQuietHours(ctx context.Context, in QuietHoursWrite) (*QuietHours, error) {
	var qh QuietHours
	if err := c.do(ctx, http.MethodPost, quietHoursPath, nil, in, &qh); err != nil {
		return nil, err
	}
	return &qh, nil
}

// UpdateQuietHours updates an existing quiet hours record.
func (c *Client) UpdateQuietHours(ctx context.Context, id string, in QuietHoursWrite) (*QuietHours, error) {
	var qh QuietHours
	path := fmt.Sprintf("%s/%s", quietHoursPath, id)
	if err := c.do(ctx, http.MethodPatch, path, nil, in, &qh); err != nil {
		return nil, err
	}
	return &qh, nil
}

// DeleteQuietHours deletes a quiet hours record by ID.
func (c *Client) DeleteQuietHours(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", quietHoursPath, id)
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}
