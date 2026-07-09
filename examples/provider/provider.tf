terraform {
  required_providers {
    beszel = {
      source = "leo-lem/beszel"
    }
  }
}

provider "beszel" {
  # endpoint, email, and password can also be set via BESZEL_ENDPOINT,
  # BESZEL_EMAIL, and BESZEL_PASSWORD environment variables.
  endpoint = "https://beszel.example.com"
  email    = "admin@example.com"
  password = var.beszel_password
}
