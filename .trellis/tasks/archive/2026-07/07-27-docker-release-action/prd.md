# GitHub Action: build & publish Docker image on release

## Goal

Automatically build and publish a multi-arch Docker image to GitHub Container Registry (GHCR) when a GitHub Release is published — not on every push or merge to main.

## Requirements

- Trigger only on `release` events (published), never on `push` or PR merge
- Build multi-arch image (`linux/amd64`, `linux/arm64`) using Docker buildx
- Push to GHCR at `ghcr.io/enkru/mango-next`
- Tag the image with:
  - The release semver (e.g. `v1.2.3`)
  - `latest`
- No extra secrets required (use built-in `GITHUB_TOKEN`)
- Use the existing `Dockerfile` at repo root

## Acceptance Criteria

- [x] Workflow file exists at `.github/workflows/docker-release.yml`
- [x] Workflow triggers only on `release: [published]`
- [x] Buildx builds for both `linux/amd64` and `linux/arm64`
- [x] Image is pushed to `ghcr.io/enkru/mango-next:<tag>` and `ghcr.io/enkru/mango-next:latest`
- [x] No other events (push, pull_request, workflow_dispatch) trigger the build
