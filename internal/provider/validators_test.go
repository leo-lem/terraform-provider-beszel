package provider

import "testing"

func TestValidShoutrrrURL(t *testing.T) {
	tests := map[string]bool{
		"generic://example.com/hook":            true,
		"telegram://token@telegram?chats=123":   true,
		"discord://token@id":                    true,
		"slack://token-a/token-b/token-c":       true,
		"slack://hook:T123/B456/secret@channel": true,
		"":                                      false,
		"not a url":                             false,
		"://broken":                             false,
		"javascript:alert(1)":                   false,
		"generic://":                            false,
		"https:/missing-slash":                  false,
	}
	for in, want := range tests {
		if got := validShoutrrrURL(in); got != want {
			t.Errorf("validShoutrrrURL(%q) = %v, want %v", in, got, want)
		}
	}
}
