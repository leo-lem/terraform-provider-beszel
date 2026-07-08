package client_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestListUsers(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/records":
			if r.URL.Query().Get("filter") != `email = "a@example.com"` {
				t.Fatalf("unexpected filter: %q", r.URL.Query().Get("filter"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []client.User{{ID: "userid1234567", Email: "a@example.com"}}})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.ListUsers(t.Context(), `email = "a@example.com"`)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(got) != 1 || got[0].Email != "a@example.com" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGetUserNotFound(t *testing.T) {
	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/records/missing":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))

	_, err := c.GetUser(t.Context(), "missing")
	if !client.IsNotFound(err) {
		t.Fatalf("expected IsNotFound(err) to be true, got %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	in := client.UserCreate{Email: "a@example.com", Password: "supersecret1", PasswordConfirm: "supersecret1", Role: "admin", Verified: true}

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/records":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			var got client.UserCreate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Email != in.Email || got.Role != in.Role {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.User{ID: "userid1234567", Email: got.Email, Role: got.Role, Verified: got.Verified})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.CreateUser(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if got.ID == "" || got.Email != in.Email {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestUpdateUserPartial(t *testing.T) {
	newRole := "readonly"

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/records/userid1234567":
			var got client.UserUpdate
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if got.Password != nil {
				t.Fatalf("expected password to be omitted, got %+v", got.Password)
			}
			if got.Role == nil || *got.Role != newRole {
				t.Fatalf("unexpected request body: %+v", got)
			}
			_ = json.NewEncoder(w).Encode(client.User{ID: "userid1234567", Role: newRole})
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := c.UpdateUser(t.Context(), "userid1234567", client.UserUpdate{Role: &newRole})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if got.Role != newRole {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestDeleteUser(t *testing.T) {
	var deleted bool

	c, _ := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/_superusers/auth-with-password":
			authHandler(w, r)
		case "/api/collections/users/records/userid1234567":
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))

	if err := c.DeleteUser(t.Context(), "userid1234567"); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if !deleted {
		t.Fatal("expected DELETE request to be sent")
	}
}
