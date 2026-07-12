# Codebase Review Execution Plan

## Checklist

- [x] Read local frontend spec index and relevant guideline files.
- [x] Inventory repository structure and identify entry points.
- [x] Inspect project manifests and discover quality commands.
- [x] Review high-risk backend modules.
- [x] Review frontend scripts and ECR templates.
- [x] Review Docker, deployment, environment, and documentation assets.
- [x] Run feasible local checks such as Crystal specs, format checks, dependency audits, or frontend build checks.
- [x] Verify every candidate finding against source code and reduce false positives.
- [x] Produce final findings ordered by severity with file and line references.

## Validation Commands To Consider

- `crystal spec`
- `crystal tool format --check src spec`
- `npm audit --json`
- `npm test` or available package scripts, if present
- `docker compose config`, if Docker validation is needed and available

## Review Hotspots

- `src/server.cr`
- `src/config.cr`
- `src/routes/`
- `src/handlers/`
- `src/storage.cr`
- `src/upload.cr`
- `src/archive.cr`
- `src/plugin/`
- `src/util/`
- `src/views/`
- `public/js/`
- `Dockerfile`
- `docker-compose*.yml`
- `env.example`

## Stop Conditions

- A required dependency is missing and cannot be installed without user approval.
- A check requires network or production credentials.
- The repository state changes unexpectedly in a way that invalidates the review baseline.
