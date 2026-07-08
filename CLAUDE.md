# terraform-provider-beszel

Unofficial Terraform provider for Beszel (self-hosted server monitoring on PocketBase). Uses the
Terraform Plugin Framework, with CRUD reconciliation, acceptance testing, and release tooling. See
README.md for the full data-model mapping and build order.

## Writing style

- Never use emojis.
- Never use double dashes (--) or em dashes (—) anywhere in code, comments, commit messages, or
  docs. Use a period, comma, colon, or parentheses instead.
- Do not describe this project as a portfolio project, learning exercise, or similar, anywhere
  (README, docs, comments, commit messages).

## Conventions

- Resources live in `internal/provider/`, one file per resource/data source, following the pattern
  established by `beszel_system` (the first resource: treat it as the reference implementation for
  CRUD/import/error-handling shape).
- Resources never call HTTP directly. Always through `internal/client/`, one file per PocketBase
  collection, typed methods only (no raw `map[string]interface{}` leaking into provider code).
- Collections that are agent-populated telemetry (`system_stats`, `container_stats`, `system_details`,
  `containers`, `systemd_services`, `smart_devices`, `quiet_hours`, `alerts_history`, `fingerprints`)
  are intentionally not modeled as resources. Don't add them without discussing scope first.
- The provider requires superuser (PocketBase admin) auth. Don't design around regular-user auth
  working for `systems`/`users` CRUD, it doesn't, by Beszel's own collection rules.

## Local Beszel instance

`docker-compose.yml` at repo root runs a hub and agent for manual testing and acceptance tests.

- Superuser: create via `docker compose exec beszel /beszel superuser upsert <email> <password>`.
- Universal tokens (used by the agent's `TOKEN`/`KEY` env vars) are user-scoped, not
  superuser-scoped. A superuser calling `/api/beszel/universal-token` gets a 403. Create a regular
  `users` record (role `admin`) via superuser auth first, then authenticate as that user to fetch the
  token.
- Hub's SSH public key (agent's `KEY` env var) comes from `/api/beszel/getkey`, authenticated.
- `beszel_data/`, `beszel_agent_data/`, `beszel_socket/` are gitignored runtime state, safe to
  delete and recreate.

## Testing

- Unit tests (client layer, `httptest`) run with plain `go test ./...`, no live hub needed.
- Acceptance tests (`TF_ACC=1`) need the real Beszel hub from `docker-compose.yml` running and a
  superuser already provisioned as above.

## Reference

Beszel's actual schema and access-rules ground truth is its own source, not just its docs:
`internal/migrations/0_collections_snapshot_*.go` (field definitions) and
`internal/hub/collections.go` (collection rules/permissions) in the henrygd/beszel repo (MIT
licensed). Re-verify against that source if Beszel's API appears to have changed.
