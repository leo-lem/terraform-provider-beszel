terraform {
  required_providers {
    beszel = {
      source = "leo-lem/beszel"
    }
  }
}

provider "beszel" {
  # endpoint, email, and password can also be set via BESZEL_ENDPOINT,
  # BESZEL_EMAIL, and BESZEL_PASSWORD environment variables, leaving this
  # block empty.
  endpoint = "https://beszel.example.com"
  email    = "admin@example.com"
  password = var.beszel_password
}

variable "beszel_password" {
  type      = string
  sensitive = true
}
