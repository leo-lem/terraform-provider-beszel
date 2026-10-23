package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// UserSettingsData is the shape of the "settings" JSON blob on a user_settings
// record. Beszel replaces the whole blob on write, and the web UI stores many
// other preferences in it (units, language, layout), so keys this provider
// does not manage are carried through untouched in extra.
type UserSettingsData struct {
	Emails   []string
	Webhooks []string

	extra map[string]json.RawMessage
}

// UnmarshalJSON decodes the managed fields and keeps every other key.
func (d *UserSettingsData) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	*d = UserSettingsData{}
	if v, ok := raw["emails"]; ok {
		if err := json.Unmarshal(v, &d.Emails); err != nil {
			return err
		}
		delete(raw, "emails")
	}
	if v, ok := raw["webhooks"]; ok {
		if err := json.Unmarshal(v, &d.Webhooks); err != nil {
			return err
		}
		delete(raw, "webhooks")
	}
	d.extra = raw
	return nil
}

// MarshalJSON encodes the managed fields over the preserved extra keys.
func (d UserSettingsData) MarshalJSON() ([]byte, error) {
	out := make(map[string]json.RawMessage, len(d.extra)+2)
	for k, v := range d.extra {
		out[k] = v
	}
	emails, webhooks := d.Emails, d.Webhooks
	if emails == nil {
		emails = []string{}
	}
	if webhooks == nil {
		webhooks = []string{}
	}
	var err error
	if out["emails"], err = json.Marshal(emails); err != nil {
		return nil, err
	}
	if out["webhooks"], err = json.Marshal(webhooks); err != nil {
		return nil, err
	}
	return json.Marshal(out)
}

// UserSettings is a Beszel "user_settings" collection record. There is a
// unique index on user, so it behaves as a singleton per user.
type UserSettings struct {
	ID       string           `json:"id"`
	User     string           `json:"user"`
	Settings UserSettingsData `json:"settings"`
	Created  string           `json:"created"`
	Updated  string           `json:"updated"`
}

// UserSettingsWrite holds the fields accepted when creating or updating user settings.
type UserSettingsWrite struct {
	User     string           `json:"user,omitempty"`
	Settings UserSettingsData `json:"settings"`
}

type userSettingsList struct {
	Items []UserSettings `json:"items"`
}

const userSettingsPath = "/api/collections/user_settings/records"

// ListUserSettings returns user_settings records matching the given PocketBase filter expression.
func (c *Client) ListUserSettings(ctx context.Context, filter string) ([]UserSettings, error) {
	query := url.Values{}
	if filter != "" {
		query.Set("filter", filter)
	}
	query.Set("perPage", "500")

	var list userSettingsList
	if err := c.do(ctx, http.MethodGet, userSettingsPath, query, nil, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetUserSettings fetches a single user_settings record by its PocketBase record ID.
func (c *Client) GetUserSettings(ctx context.Context, id string) (*UserSettings, error) {
	var settings UserSettings
	path := fmt.Sprintf("%s/%s", userSettingsPath, id)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// CreateUserSettings creates the user_settings record for a user. Fails if one already exists.
func (c *Client) CreateUserSettings(ctx context.Context, in UserSettingsWrite) (*UserSettings, error) {
	var settings UserSettings
	if err := c.do(ctx, http.MethodPost, userSettingsPath, nil, in, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateUserSettings replaces the settings blob on an existing user_settings record.
func (c *Client) UpdateUserSettings(ctx context.Context, id string, in UserSettingsWrite) (*UserSettings, error) {
	var settings UserSettings
	path := fmt.Sprintf("%s/%s", userSettingsPath, id)
	if err := c.do(ctx, http.MethodPatch, path, nil, in, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// DeleteUserSettings deletes a user_settings record by ID.
func (c *Client) DeleteUserSettings(ctx context.Context, id string) error {
	path := fmt.Sprintf("%s/%s", userSettingsPath, id)
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}
