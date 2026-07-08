resource "beszel_alert" "example" {
  user   = beszel_user.example.id
  system = beszel_system.example.id
  name   = "CPU"
  value  = 90
  min    = 10
}
