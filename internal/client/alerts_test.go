package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestListAlerts(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/alerts/records":
			if r.URL.Query().Get("filter") != `system = "sys123"` {
				t.Fatalf("unexpected filter: %q", r.URL.Query().Get("filter"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []client.Alert{{ID: "alertid123456", System: "sys123", Name: "CPU"}}})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.ListAlerts(t.Context(), `system = "sys123"`)
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(got) != 1 || got[0].Name != "CPU" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetAlertNotFound(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/alerts/records/missing":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))

	_, err := c.GetAlert(t.Context(), "missing")
	if !client.IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestCreateAlert(t *testing.T) {
	min := 5
	in := client.AlertCreate{User: "user123", System: "sys123", Name: "CPU", Min: &min}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/alerts/records":
			var got client.AlertCreate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Name != in.Name || got.Min == nil || *got.Min != min {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.Alert{ID: "alertid123456", User: got.User, System: got.System, Name: got.Name, Min: got.Min})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.CreateAlert(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}
	if got.ID == "" || got.Min == nil || *got.Min != min {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestUpdateAlert(t *testing.T) {
	value := 90.0

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/alerts/records/alertid123456":
			var got client.AlertUpdate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Value == nil || *got.Value != value {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.Alert{ID: "alertid123456", Value: got.Value})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.UpdateAlert(t.Context(), "alertid123456", client.AlertUpdate{Value: &value})
	if err != nil {
		t.Fatalf("UpdateAlert: %v", err)
	}
	if got.Value == nil || *got.Value != value {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestDeleteAlert(t *testing.T) {
	var deleted bool

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/alerts/records/alertid123456":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	if err := c.DeleteAlert(t.Context(), "alertid123456"); err != nil {
		t.Fatalf("DeleteAlert: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE request to be sent")
	}
}
