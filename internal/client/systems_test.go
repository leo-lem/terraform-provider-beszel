package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestListSystems(t *testing.T) {
	want := []client.System{
		{ID: "abc123", Name: "docker-desktop", Host: "172.21.0.1", Status: "up"},
	}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/systems/records":
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": want})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.ListSystems(t.Context(), "")
	if err != nil {
		t.Fatalf("ListSystems: %v", err)
	}
	if len(got) != 1 || got[0].Name != "docker-desktop" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetSystemNotFound(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/systems/records/missing":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "The requested resource wasn't found."})
		default:
			http.NotFound(w, r)
		}
	}))

	_, err := c.GetSystem(t.Context(), "missing")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !client.IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestCreateSystem(t *testing.T) {
	in := client.SystemCreate{Name: "web-1", Host: "10.0.0.1", Users: []string{"user123"}}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/systems/records":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var got client.SystemCreate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Name != in.Name || got.Host != in.Host {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.System{
				ID: "newid1234567890", Name: got.Name, Host: got.Host, Users: got.Users, Status: "pending",
			})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.CreateSystem(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateSystem: %v", err)
	}
	if got.ID == "" || got.Name != in.Name {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestUpdateSystem(t *testing.T) {
	newHost := "10.0.0.2"

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/systems/records/abc123":
			if r.Method != http.MethodPatch {
				t.Errorf("expected PATCH, got %s", r.Method)
			}
			var got client.SystemUpdate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Host == nil || *got.Host != newHost {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.System{ID: "abc123", Host: newHost})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.UpdateSystem(t.Context(), "abc123", client.SystemUpdate{Host: &newHost})
	if err != nil {
		t.Fatalf("UpdateSystem: %v", err)
	}
	if got.Host != newHost {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestDeleteSystem(t *testing.T) {
	var deleted bool

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/systems/records/abc123":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	if err := c.DeleteSystem(t.Context(), "abc123"); err != nil {
		t.Fatalf("DeleteSystem: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE request to be sent")
	}
}
