resource "beszel_universal_token" "example" {
  user      = beszel_user.example.id
  permanent = true
}
