package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func newTestServer(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return client.New(server.URL, "admin@example.com", "password"), server
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/collections/_superusers/auth-with-password" {
		http.NotFound(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"token": "test-token"})
}

func TestClientAuthenticatesOnFirstRequest(t *testing.T) {
	var authCalls int32

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			atomic.AddInt32(&authCalls, 1)
			authHandler(w, r)
		case "/api/collections/systems/records":
			if r.Header.Get("Authorization") != "test-token" {
				t.Errorf("expected Authorization header to be the auth token, got %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []client.System{}})
		default:
			http.NotFound(w, r)
		}
	}))

	if _, err := c.ListSystems(t.Context(), ""); err != nil {
		t.Fatalf("ListSystems: %v", err)
	}
	if got := atomic.LoadInt32(&authCalls); got != 1 {
		t.Fatalf("expected exactly 1 auth call, got %d", got)
	}
}

func TestClientReAuthenticatesOn401(t *testing.T) {
	var authCalls int32
	var systemsCalls int32

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			atomic.AddInt32(&authCalls, 1)
			authHandler(w, r)
		case "/api/collections/systems/records":
			call := atomic.AddInt32(&systemsCalls, 1)
			if call == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 401, "message": "expired token"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []client.System{}})
		default:
			http.NotFound(w, r)
		}
	}))

	if _, err := c.ListSystems(t.Context(), ""); err != nil {
		t.Fatalf("ListSystems: %v", err)
	}
	if got := atomic.LoadInt32(&authCalls); got != 2 {
		t.Fatalf("expected 2 auth calls (initial + re-auth), got %d", got)
	}
	if got := atomic.LoadInt32(&systemsCalls); got != 2 {
		t.Fatalf("expected 2 systems calls (failed + retried), got %d", got)
	}
}

func TestAPIErrorFromNonEnvelopeBody(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("internal server error"))
		}
	}))

	_, err := c.ListSystems(t.Context(), "")
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*client.APIError)
	if !ok {
		t.Fatalf("expected *client.APIError, got %T: %v", err, err)
	}
	if apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", apiErr.Status)
	}
}
