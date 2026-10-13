package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestCreateQuietHours(t *testing.T) {
	in := client.QuietHoursWrite{User: "user123", Type: "daily", Start: "2026-10-13T22:00:00Z", End: "2026-10-14T06:00:00Z"}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/quiet_hours/records":
			var got client.QuietHoursWrite
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got != in {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.QuietHours{ID: "qh12345678", User: got.User, Type: got.Type, Start: got.Start, End: got.End})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.CreateQuietHours(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateQuietHours: %v", err)
	}
	if got.ID == "" || got.Type != "daily" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestUpdateQuietHoursSendsEmptySystem(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/quiet_hours/records/qh12345678":
			var got map[string]any
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if sys, ok := got["system"]; !ok || sys != "" {
				t.Fatalf("system should be sent as empty string: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.QuietHours{ID: "qh12345678"})
		default:
			http.NotFound(w, r)
		}
	}))

	if _, err := c.UpdateQuietHours(t.Context(), "qh12345678", client.QuietHoursWrite{Type: "daily"}); err != nil {
		t.Fatalf("UpdateQuietHours: %v", err)
	}
}

func TestGetQuietHoursNotFound(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "not found"})
		}
	}))

	if _, err := c.GetQuietHours(t.Context(), "missing"); !client.IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err), got %v", err)
	}
}
