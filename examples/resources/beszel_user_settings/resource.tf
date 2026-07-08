resource "beszel_user_settings" "example" {
  user   = beszel_user.example.id
  emails = ["alerts@example.com"]
}
