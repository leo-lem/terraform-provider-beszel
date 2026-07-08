package client

import "fmt"

// APIError represents the error envelope PocketBase returns on non-2xx responses:
//
//	{"code": 400, "message": "...", "data": {"field": {"code": "...", "message": "..."}}}
type APIError struct {
	Status  int                      `json:"code"`
	Message string                   `json:"message"`
	Data    map[string]APIFieldError `json:"data"`
}

// APIFieldError is a single field-level validation error inside an APIError's Data map.
type APIFieldError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("beszel: request failed with status %d: %s", e.Status, e.Message)
}

// IsNotFound reports whether err represents a 404 response from the Beszel API.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Status == 404
}
