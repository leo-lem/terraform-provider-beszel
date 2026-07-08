package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestCreateUserSettings(t *testing.T) {
	in := client.UserSettingsWrite{User: "user123", Settings: client.UserSettingsData{Emails: []string{"a@example.com"}}}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/user_settings/records":
			var got client.UserSettingsWrite
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if len(got.Settings.Emails) != 1 || got.Settings.Emails[0] != "a@example.com" {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.UserSettings{ID: "settingsid123", User: got.User, Settings: got.Settings})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.CreateUserSettings(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateUserSettings: %v", err)
	}
	if got.ID == "" || len(got.Settings.Emails) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestListUserSettingsByUserFilter(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/user_settings/records":
			if r.URL.Query().Get("filter") != `user = "user123"` {
				t.Fatalf("unexpected filter: %q", r.URL.Query().Get("filter"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []client.UserSettings{{ID: "settingsid123", User: "user123"}}})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.ListUserSettings(t.Context(), `user = "user123"`)
	if err != nil {
		t.Fatalf("ListUserSettings: %v", err)
	}
	if len(got) != 1 || got[0].User != "user123" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
