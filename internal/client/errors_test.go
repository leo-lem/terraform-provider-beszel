package client_test

import (
	"fmt"
	"testing"

	"github.com/leo-lem/terraform-provider-beszel/internal/client"
)

func TestAPIErrorMessage(t *testing.T) {
	err := &client.APIError{Status: 400, Message: "validation failed"}
	got := err.Error()
	if got != "beszel: request failed with status 400: validation failed" {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestIsNotFoundOnWrappedError(t *testing.T) {
	err := fmt.Errorf("reading system: %w", &client.APIError{Status: 404, Message: "not found"})
	if !client.IsNotFound(err) {
		t.Fatal("expected IsNotFound to unwrap and return true")
	}
}

func TestIsNotFoundFalseForOtherStatus(t *testing.T) {
	err := &client.APIError{Status: 500, Message: "server error"}
	if client.IsNotFound(err) {
		t.Fatal("expected IsNotFound to be false for a 500")
	}
}

func TestIsNotFoundFalseForNil(t *testing.T) {
	if client.IsNotFound(nil) {
		t.Fatal("expected IsNotFound(nil) to be false")
	}
}
