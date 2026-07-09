# Changelog

All notable changes to this provider are documented here. Versions follow
[Semantic Versioning](https://semver.org/); this format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## v0.2.0 (2026-07-09)

### Added

- `beszel_universal_token` resource, managing a user's universal agent registration token.
  Beszel scopes the underlying endpoint to the caller's own record and forbids superuser callers,
  so the client impersonates the target user (via PocketBase's superuser impersonation route)
  before calling it.
- A full worked example under `examples/guides/getting-started`, rendered as a Terraform Registry
  guide: creates a user, a system, an alert, notification settings, and a universal token in one
  configuration.

### Changed

- CI now runs `golangci-lint` (via `make lint`), not just `gofmt` and `go vet`.

## v0.1.0 (2026-07-09)

Initial release.

### Added

- `beszel_system` resource and `beszel_system`/`beszel_systems` data sources.
- `beszel_user` resource and `beszel_user` data source.
- `beszel_alert` resource.
- `beszel_user_settings` resource.
