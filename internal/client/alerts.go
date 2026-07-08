package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Alert is a Beszel "alerts" collection record. There is a unique index on
// (user, system, name), so a given owner/system/kind combination can only
// have one alert record.
type Alert struct {
	ID        string   `json:"id"`
	User      string   `json:"user"`
	System    string   `json:"system"`
	Name      string   `json:"name"`
	Value     *float64 `json:"value"`
	Min       *int     `json:"min"`
	Triggered bool     `json:"triggered"`
	Created   string   `json:"created"`
	Updated   string   `json:"updated"`
}

// AlertCreate holds the fields accepted when creating an alert.
type AlertCreate struct {
	User   string   `json:"user"`
	System string   `json:"system"`
	Name   string   `json:"name"`
	Value  *float64 `json:"value,omitempty"`
	Min    *int     `json:"min,omitempty"`
}

// AlertUpdate holds the fields accepted when updating an alert.
type AlertUpdate struct {
	Value *float64 `json:"value,omitempty"`
	Min   *int     `json:"min,omitempty"`
}

type alertList struct {
	Items []Alert `json:"items"`
}

const alertsPath = "/api/collections/alerts/records"

// ListAlerts returns alerts matching the given PocketBase filter expression.
func (c *Client) ListAlerts(ctx context.Context, filter string) ([]Alert, error) {
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	query.Set("perPage", "500")

	var list alertList
	if err := c.do(ctx, http.MethodGet, alertsPath, query, nil, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetAlert fetches a single alert by its PocketBase record ID.
func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	var alert Alert
	path := fmt.Sprintf("%s/%s", alertsPath, id)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// CreateAlert creates a new alert record.
func (c *Client) CreateAlert(ctx context.Context, in AlertCreate) (*Alert, error) {
	var alert Alert
	if err := c.do(ctx, http.MethodPost, alertsPath, nil, in, &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// UpdateAlert partially updates an existing alert record.
func (c *Client) UpdateAlert(ctx context.Context, id string, in AlertUpdate) (*Alert, error) {
	var alert Alert
	path := fmt.Sprintf("%s/%s", alertsPath, id)
	if err := c.do(ctx, http.MethodPatch, path, nil, in, &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// DeleteAlert deletes an alert record by ID.
func (c *Client) DeleteAlert(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", alertsPath, id)
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}
