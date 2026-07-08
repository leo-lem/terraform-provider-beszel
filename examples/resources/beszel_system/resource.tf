resource "beszel_system" "example" {
  name  = "web-1"
  host  = "10.0.0.1"
  users = [beszel_user.example.id]
}
