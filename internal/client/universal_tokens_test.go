package client_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func impersonateHandler(t *testing.T, userID string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		wantPath := "/api/collections/users/impersonate/" + userID
		if r.URL.Path != wantPath {
			t.Errorf("expected path %s, got %s", wantPath, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "impersonated-token"})
	}
}

func TestEnableUniversalToken(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/impersonate/user123":
			impersonateHandler(t, "user123")(w, r)
		case "/api/beszel/universal-token":
			if r.Header.Get("Authorization") != "impersonated-token" {
				t.Errorf("expected impersonated Authorization header, got %q", r.Header.Get("Authorization"))
			}
			if got := r.URL.Query().Get("enable"); got != "1" {
				t.Errorf("expected enable=1, got %q", got)
			}
			if got := r.URL.Query().Get("permanent"); got != "1" {
				t.Errorf("expected permanent=1, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "generated-token", "active": true, "permanent": true})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.EnableUniversalToken(t.Context(), "user123", "", true)
	if err != nil {
		t.Fatalf("EnableUniversalToken: %v", err)
	}
	if got.Token != "generated-token" || !got.Active || !got.Permanent {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestDisableUniversalToken(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/impersonate/user123":
			impersonateHandler(t, "user123")(w, r)
		case "/api/beszel/universal-token":
			if got := r.URL.Query().Get("enable"); got != "0" {
				t.Errorf("expected enable=0, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "", "active": false, "permanent": false})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.DisableUniversalToken(t.Context(), "user123")
	if err != nil {
		t.Fatalf("DisableUniversalToken: %v", err)
	}
	if got.Active || got.Permanent {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetUniversalToken(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/impersonate/user123":
			impersonateHandler(t, "user123")(w, r)
		case "/api/beszel/universal-token":
			if r.URL.Query().Has("enable") {
				t.Errorf("expected no enable param on a read, got %q", r.URL.Query().Get("enable"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"token": "existing-token", "active": true, "permanent": false})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.GetUniversalToken(t.Context(), "user123")
	if err != nil {
		t.Fatalf("GetUniversalToken: %v", err)
	}
	if got.Token != "existing-token" || !got.Active || got.Permanent {
		t.Fatalf("unexpected result: %+v", got)
	}
}
