# Implementation Plan

## Execution checklist

- [ ] Create `.github/workflows/docker-release.yml`
  - Trigger: `release: [published]`
  - Steps:
    1. Checkout repo
    2. Set up Docker Buildx (QEMU for multi-arch)
    3. Log in to GHCR using `GITHUB_TOKEN`
    4. Extract semver tag from `github.event.release.tag_name`
    5. Build & push: `ghcr.io/enkru/mango-next:<semver>`, `ghcr.io/enkru/mango-next:latest`
    6. Platforms: `linux/amd64,linux/arm64`

## Validation

- Verify the workflow file parses correctly with `act` or GitHub web UI
- Manual dry-run: inspect the rendered YAML has no syntax errors

## Rollback

- Delete or rename `.github/workflows/docker-release.yml` to disable the workflow
