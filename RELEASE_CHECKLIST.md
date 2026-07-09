# Release checklist

Temporary tracking file, delete once the provider is published and the checklist below is done.

## One-time steps before the first real release

1. Make the repository public (the Terraform Registry only indexes public repos).
2. Generate a GPG key pair; add the public key to the Terraform Registry account under
   Publishing -> Signing Keys.
3. Add `GPG_PRIVATE_KEY` and `PASSPHRASE` as repository secrets in GitHub Actions.
4. Sign in to registry.terraform.io with GitHub and publish this provider (Publish -> Provider),
   pointing it at this repository.
5. Push a `v0.1.0` tag. `release.yml` picks it up, GoReleaser builds and signs the release, and the
   registry ingests the new version automatically within minutes.

Until step 5, `.github/workflows/release.yml` exists but has nothing to trigger it.

## Build order (for reference, all complete except release/registry publish)

1. Provider skeleton and auth
2. Client layer for the `systems` collection
3. Core resources and data sources: `beszel_system`, `beszel_user`, `beszel_alert`,
   `beszel_user_settings`, and the read-only `beszel_system(s)`/`beszel_user` data sources
4. Docs generation and release tooling (goreleaser, signed releases, registry publish)
5. CI/CD: unit tests on every PR, acceptance tests against a real Beszel hub, automatic release
   to the Terraform Registry on tag push (current state: workflows in place, registry publish and
   first tag pending the one-time setup above)
