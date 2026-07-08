package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// System is a Beszel "systems" collection record.
type System struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Host    string          `json:"host"`
	Port    string          `json:"port"`
	Status  string          `json:"status"`
	Info    json.RawMessage `json:"info"`
	Users   []string        `json:"users"`
	Created string          `json:"created"`
	Updated string          `json:"updated"`
}

// SystemCreate holds the fields accepted when creating a system. Status and
// Info are agent-reported and cannot be set on create.
type SystemCreate struct {
	Name  string   `json:"name"`
	Host  string   `json:"host"`
	Port  string   `json:"port,omitempty"`
	Users []string `json:"users"`
}

// SystemUpdate holds the fields accepted when updating a system. Pointer
// fields are omitted from the request when nil, so callers can update a
// subset of fields without clobbering the rest.
type SystemUpdate struct {
	Name  *string   `json:"name,omitempty"`
	Host  *string   `json:"host,omitempty"`
	Port  *string   `json:"port,omitempty"`
	Users *[]string `json:"users,omitempty"`
}

type systemList struct {
	Items []System `json:"items"`
}

const systemsPath = "/api/collections/systems/records"

// ListSystems returns systems matching the given PocketBase filter expression.
// An empty filter returns all systems.
func (c *Client) ListSystems(ctx context.Context, filter string) ([]System, error) {
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	query.Set("perPage", "500")

	var list systemList
	if err := c.do(ctx, http.MethodGet, systemsPath, query, nil, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetSystem fetches a single system by its PocketBase record ID.
func (c *Client) GetSystem(ctx context.Context, id string) (*System, error) {
	var system System
	path := fmt.Sprintf("%s/%s", systemsPath, id)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &system); err != nil {
		return nil, err
	}
	return &system, nil
}

// CreateSystem creates a new system record.
func (c *Client) CreateSystem(ctx context.Context, in SystemCreate) (*System, error) {
	var system System
	if err := c.do(ctx, http.MethodPost, systemsPath, nil, in, &system); err != nil {
		return nil, err
	}
	return &system, nil
}

// UpdateSystem partially updates an existing system record.
func (c *Client) UpdateSystem(ctx context.Context, id string, in SystemUpdate) (*System, error) {
	var system System
	path := fmt.Sprintf("%s/%s", systemsPath, id)
	if err := c.do(ctx, http.MethodPatch, path, nil, in, &system); err != nil {
		return nil, err
	}
	return &system, nil
}

// DeleteSystem deletes a system record by ID.
func (c *Client) DeleteSystem(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", systemsPath, id)
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}
