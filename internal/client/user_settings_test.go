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

func TestGetUserSettingsNotFound(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/user_settings/records/missing":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))

	_, err := c.GetUserSettings(t.Context(), "missing")
	if !client.IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestUpdateUserSettings(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/user_settings/records/settingsid123":
			if r.Method != http.MethodPatch {
				t.Errorf("expected PATCH, got %s", r.Method)
			}
			var got client.UserSettingsWrite
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(client.UserSettings{ID: "settingsid123", Settings: got.Settings})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.UpdateUserSettings(t.Context(), "settingsid123", client.UserSettingsWrite{
		Settings: client.UserSettingsData{Emails: []string{"b@example.com"}},
	})
	if err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}
	if len(got.Settings.Emails) != 1 || got.Settings.Emails[0] != "b@example.com" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestDeleteUserSettings(t *testing.T) {
	var deleted bool

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/user_settings/records/settingsid123":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	if err := c.DeleteUserSettings(t.Context(), "settingsid123"); err != nil {
		t.Fatalf("DeleteUserSettings: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE request to be sent")
	}
}

func TestUserSettingsDataPreservesUnknownKeys(t *testing.T) {
	var data client.UserSettingsData
	in := `{"emails":["a@example.com"],"webhooks":["generic://x"],"lang":"en","unitTemp":{"v":1}}`
	if err := json.Unmarshal([]byte(in), &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(data.Webhooks) != 1 || data.Webhooks[0] != "generic://x" {
		t.Fatalf("unexpected webhooks: %+v", data.Webhooks)
	}

	data.Webhooks = []string{"generic://y"}
	out, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decoding output: %v", err)
	}
	if got["lang"] != "en" || got["unitTemp"] == nil {
		t.Fatalf("unknown keys were dropped: %s", out)
	}
	if hooks, _ := got["webhooks"].([]any); len(hooks) != 1 || hooks[0] != "generic://y" {
		t.Fatalf("unexpected webhooks in output: %s", out)
	}
}

func TestUserSettingsDataNilSlicesMarshalAsEmpty(t *testing.T) {
	out, err := json.Marshal(client.UserSettingsData{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decoding output: %v", err)
	}
	if e, ok := got["emails"].([]any); !ok || len(e) != 0 {
		t.Fatalf("emails should be []: %s", out)
	}
	if w, ok := got["webhooks"].([]any); !ok || len(w) != 0 {
		t.Fatalf("webhooks should be []: %s", out)
	}
}
