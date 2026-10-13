package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserSettingsResource(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsResourceConfig(userID, "a@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.#", "1"),
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.0", "a@example.com"),
					resource.TestCheckResourceAttrSet("beszel_user_settings.test", "id"),
				),
			},
			{
				ResourceName:      "beszel_user_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserSettingsResourceConfig(userID, "b@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "emails.0", "b@example.com"),
				),
			},
		},
	})
}

func TestAccUserSettingsResourceWebhooks(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsWebhooksConfig(userID, `["generic://example.com/one"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "webhooks.#", "1"),
					resource.TestCheckResourceAttr("beszel_user_settings.test", "webhooks.0", "generic://example.com/one"),
				),
			},
			{
				ResourceName:      "beszel_user_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserSettingsWebhooksConfig(userID, `["generic://example.com/two", "generic://example.com/three"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "webhooks.#", "2"),
					resource.TestCheckResourceAttr("beszel_user_settings.test", "webhooks.1", "generic://example.com/three"),
				),
			},
			{
				Config: testAccUserSettingsWebhooksConfig(userID, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("beszel_user_settings.test", "webhooks.#", "0"),
				),
			},
		},
	})
}

func TestAccUserSettingsResourceInvalidWebhook(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserSettingsWebhooksConfig(userID, `["not a url"]`),
				ExpectError: regexp.MustCompile(`Invalid webhook URL`),
			},
		},
	})
}

// Beszel replaces the whole settings blob on write; keys the provider does not
// manage (set here out of band, as the web UI would) must survive an apply.
func TestAccUserSettingsResourcePreservesOtherKeys(t *testing.T) {
	userID := os.Getenv("BESZEL_TEST_USER_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsResourceConfig(userID, "a@example.com"),
			},
			{
				PreConfig: func() { testAccSetUserSettingsLang(t, userID, "de") },
				Config:    testAccUserSettingsResourceConfig(userID, "b@example.com"),
				Check: func(*terraform.State) error {
					return testAccCheckUserSettingsLang(userID, "de")
				},
			},
		},
	})
}

func testAccUserSettingsWebhooksConfig(userID, webhooks string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_user_settings" "test" {
  user     = %[1]q
  emails   = ["a@example.com"]
  webhooks = %[2]s
}
`, userID, webhooks)
}

func testAccUserSettingsResourceConfig(userID, email string) string {
	return fmt.Sprintf(`
provider "beszel" {}

resource "beszel_user_settings" "test" {
  user   = %[1]q
  emails = [%[2]q]
}
`, userID, email)
}

func testAccSetUserSettingsLang(t *testing.T, userID, lang string) {
	t.Helper()
	id, settings, err := testAccRawUserSettings(userID)
	if err != nil {
		t.Fatalf("reading user_settings: %v", err)
	}
	settings["lang"] = lang
	if err := testAccPocketBase(http.MethodPatch, "/api/collections/user_settings/records/"+id,
		map[string]any{"settings": settings}, nil); err != nil {
		t.Fatalf("setting lang: %v", err)
	}
}

func testAccCheckUserSettingsLang(userID, want string) error {
	_, settings, err := testAccRawUserSettings(userID)
	if err != nil {
		return err
	}
	if got := settings["lang"]; got != want {
		return fmt.Errorf("settings.lang = %v, want %q", got, want)
	}
	return nil
}

// testAccRawUserSettings returns the raw settings blob for a user, so tests can
// see keys the provider does not model.
func testAccRawUserSettings(userID string) (string, map[string]any, error) {
	var list struct {
		Items []struct {
			ID       string         `json:"id"`
			Settings map[string]any `json:"settings"`
		} `json:"items"`
	}
	q := url.Values{"filter": {fmt.Sprintf("user = %q", userID)}}
	if err := testAccPocketBase(http.MethodGet, "/api/collections/user_settings/records?"+q.Encode(), nil, &list); err != nil {
		return "", nil, err
	}
	if len(list.Items) != 1 {
		return "", nil, fmt.Errorf("expected 1 user_settings record, found %d", len(list.Items))
	}
	return list.Items[0].ID, list.Items[0].Settings, nil
}

// testAccPocketBase makes a superuser-authenticated raw request to the hub.
func testAccPocketBase(method, path string, body, out any) error {
	endpoint := os.Getenv("BESZEL_ENDPOINT")
	creds, _ := json.Marshal(map[string]string{
		"identity": os.Getenv("BESZEL_EMAIL"),
		"password": os.Getenv("BESZEL_PASSWORD"),
	})
	resp, err := http.Post(endpoint+"/api/collections/_superusers/auth-with-password", "application/json", bytes.NewReader(creds))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var auth struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return err
	}

	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, endpoint+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth.Token)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 300 {
		return fmt.Errorf("%s %s: status %d", method, path, res.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}
