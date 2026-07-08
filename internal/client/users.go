package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// User is a Beszel "users" collection record. Password is intentionally not
// part of this struct: PocketBase never returns it, so there is nothing to read.
type User struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	EmailVisibility bool   `json:"emailVisibility"`
	Verified        bool   `json:"verified"`
	Username        string `json:"username"`
	Role            string `json:"role"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
}

// UserCreate holds the fields accepted when creating a user.
type UserCreate struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	Role            string `json:"role,omitempty"`
	Verified        bool   `json:"verified"`
}

// UserUpdate holds the fields accepted when updating a user. Password is a
// pointer so callers can leave credentials untouched on unrelated updates.
type UserUpdate struct {
	Email           *string `json:"email,omitempty"`
	Password        *string `json:"password,omitempty"`
	PasswordConfirm *string `json:"passwordConfirm,omitempty"`
	Role            *string `json:"role,omitempty"`
	Verified        *bool   `json:"verified,omitempty"`
}

type userList struct {
	Items []User `json:"items"`
}

const usersPath = "/api/collections/users/records"

// ListUsers returns users matching the given PocketBase filter expression.
func (c *Client) ListUsers(ctx context.Context, filter string) ([]User, error) {
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	query.Set("perPage", "500")

	var list userList
	if err := c.do(ctx, http.MethodGet, usersPath, query, nil, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetUser fetches a single user by its PocketBase record ID.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	path := fmt.Sprintf("%s/%s", usersPath, id)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user record.
func (c *Client) CreateUser(ctx context.Context, in UserCreate) (*User, error) {
	var user User
	if err := c.do(ctx, http.MethodPost, usersPath, nil, in, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser partially updates an existing user record.
func (c *Client) UpdateUser(ctx context.Context, id string, in UserUpdate) (*User, error) {
	var user User
	path := fmt.Sprintf("%s/%s", usersPath, id)
	if err := c.do(ctx, http.MethodPatch, path, nil, in, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser deletes a user record by ID.
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", usersPath, id)
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}
