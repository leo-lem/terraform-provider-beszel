package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

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
