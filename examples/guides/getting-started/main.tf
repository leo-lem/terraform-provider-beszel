terraform {
  required_providers {
    beszel = {
      source = "leo-lem/beszel"
    }
  }
}

provider "beszel" {
  endpoint = var.beszel_endpoint
  email    = var.beszel_email
  password = var.beszel_password
}

variable "beszel_endpoint" {
  type = string
}

variable "beszel_email" {
  type = string
}

variable "beszel_password" {
  type      = string
  sensitive = true
}

variable "owner_password" {
  type      = string
  sensitive = true
}

# A regular user who will own a monitored system and receive its alerts.
resource "beszel_user" "owner" {
  email    = "owner@example.com"
  password = var.owner_password
  role     = "user"
  verified = true
}

resource "beszel_user_settings" "owner" {
  user   = beszel_user.owner.id
  emails = ["owner@example.com"]

  # Webhook / push notifications, as Shoutrrr URLs (https://containrrr.dev/shoutrrr/).
  webhooks = ["generic://example.com/beszel-hook"]
}

# A monitored system owned by beszel_user.owner. Until an agent registers
# against it (see beszel_universal_token.owner below), it stays in "pending"
# status: Terraform only manages the declared config, not agent-reported state.
resource "beszel_system" "web" {
  name  = "web-1"
  host  = "10.0.0.1"
  users = [beszel_user.owner.id]
}

# Alert the owner on sustained high CPU.
resource "beszel_alert" "web_cpu" {
  user   = beszel_user.owner.id
  system = beszel_system.web.id
  name   = "CPU"
  value  = 90
  min    = 10
}

# Mute the owner's notifications every night, 22:00 to 06:00 UTC.
resource "beszel_quiet_hours" "owner_nightly" {
  user  = beszel_user.owner.id
  type  = "daily"
  start = "2026-01-01T22:00:00Z"
  end   = "2026-01-02T06:00:00Z"
}

# A permanent token owner's agents can use to register with this hub.
resource "beszel_universal_token" "owner" {
  user      = beszel_user.owner.id
  permanent = true
}

# Look up the system by name once more, and every system on the hub.
data "beszel_system" "web" {
  name = beszel_system.web.name
}

data "beszel_systems" "all" {
  filter = ""
}

output "system_status" {
  value = data.beszel_system.web.status
}
