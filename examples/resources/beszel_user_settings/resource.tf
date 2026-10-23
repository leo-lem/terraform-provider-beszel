resource "beszel_user_settings" "example" {
  user   = beszel_user.example.id
  emails = ["alerts@example.com"]

  # Webhook / push notifications, as Shoutrrr URLs (https://containrrr.dev/shoutrrr/).
  webhooks = ["generic://example.com/beszel-hook"]
}
