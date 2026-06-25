# Release

Releases are tag-driven. Pushing a `v*` tag starts `.github/workflows/release.yml`, which runs GoReleaser and uploads Terraform Registry assets.

## One-time Setup

The release workflow needs these repository secrets:

| Secret | Purpose |
| --- | --- |
| `GPG_PRIVATE_KEY` | Armored private key used to sign the checksum file. |
| `PASSPHRASE` | Passphrase for the private key. |

Register the matching public key in the Terraform Registry namespace before publishing provider versions.

## Publish

1. Confirm `main` is green.
2. Update `CHANGELOG.md`.
3. Run:

   ```bash
   make lint
   make test
   make generate
   make terraform-test
   go run github.com/goreleaser/goreleaser/v2@latest check
   ```

4. Tag and push:

   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

5. Watch the release workflow:

   ```bash
   gh run list --limit 10
   gh run watch <run-id> --exit-status
   ```

6. Verify the GitHub release has archives, a manifest, `SHA256SUMS`, and `SHA256SUMS.sig`.

GoReleaser changelog generation is disabled, so edit the GitHub release notes from `CHANGELOG.md` after publish when the release body needs detail.
