# Mute notifications for every system, every night from 22:00 to 06:00 UTC.
# For daily windows only the UTC time of day of start and end matters.
resource "beszel_quiet_hours" "nightly" {
  user  = beszel_user.example.id
  type  = "daily"
  start = "2026-01-01T22:00:00Z"
  end   = "2026-01-02T06:00:00Z"
}

# Mute one system during a planned maintenance window.
resource "beszel_quiet_hours" "maintenance" {
  user   = beszel_user.example.id
  system = beszel_system.example.id
  type   = "one-time"
  start  = "2026-06-01T01:00:00Z"
  end    = "2026-06-01T03:00:00Z"
}
