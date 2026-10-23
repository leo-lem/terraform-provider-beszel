# Changelog

All notable changes to this provider are documented here. Versions follow
[Semantic Versioning](https://semver.org/); this format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Added

- `webhooks` on `beszel_user_settings`: webhook / push notification URLs (Shoutrrr format),
  matching the "Webhook / Push notifications" setting in the Beszel UI. Marked sensitive.
- `beszel_quiet_hours` resource, managing quiet hours windows (`one-time` or `daily`, global or
  scoped to one system) that mute alert notifications for a user. Rejects, at plan time, windows
  Beszel accepts but can never match (one-time with `end` not after `start`, daily with the same
  UTC time of day for `start` and `end`).
- Plan time validation of `webhooks` entries on `beszel_user_settings`: each must look like
  `scheme://...`. Beszel itself does not validate them server side.

### Fixed

- `beszel_user_settings` no longer wipes the other keys Beszel keeps in the same settings blob
  (units, language, layout, ...) when it creates or updates a record.

## v0.3.0 (2026-07-09)

### Added

- `beszel_alert` can now be imported by its natural key (`user/system/name`, Beszel's own unique
  index on the alerts collection) as well as its PocketBase record id, matching what's actually
  visible from the hub UI.

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
