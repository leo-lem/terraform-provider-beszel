resource "beszel_user" "example" {
  email    = "alerts@example.com"
  password = var.beszel_user_password
  role     = "user"
  verified = true
}
