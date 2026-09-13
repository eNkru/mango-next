# Clean up build config (compose alias, Makefile aliases)

## Goal

Remove a redundant compose file and unreferenced Makefile alias targets
(~27 lines). Zero-risk: no active docs or scripts use them.

## Confirmed facts (verified 2026-09-12)

- `docker-compose.go.yml` is self-labeled "Alias of docker-compose.yml"
  with weaker env handling (no `:-defaults`, hard `${PORT}`/paths vs the
  main file's `${PORT:-9000}` / `./data` / `./config`). Referenced only
  in archived Trellis tasks + journal (historical), never in active docs.
  The main docker-compose.yml covers it.
- Makefile `assets-install` / `assets-build` / `assets-check` (lines
  24-26): "Back-compat aliases while docs migrate off the assets:* names."
- Makefile `go-build` / `go-static` / `go-test` / `go-check` / `go-run` /
  `go-all` (lines 53-58): "Aliases kept for existing docs / muscle
  memory." README references neither set — only `make run / test / check
  / build / static / all`. Also trim the 9 alias names from the `.PHONY`
  line (line 6).

## Requirements

1. Delete docker-compose.go.yml.
2. Delete the 9 alias targets + their `.PHONY` entries from the Makefile.
   Keep the comment history out — just remove.

## Acceptance criteria

- [ ] `make run`, `make test`, `make check`, `make build` all still work
  (names unchanged).
- [ ] `docker compose config` succeeds with the remaining main file.
- [ ] `grep -rn "docker-compose.go\|make assets-\|make go-" README.md
  DEPLOY_QNAP.md DOCKER_HUB.md FRONTEND_DEV_GUIDE.md Makefile` returns
  nothing (proving no active doc referenced them).

## Out of scope

- docker-compose.qnap.yml / docker-compose.qnap-prebuilt.yml and
  build-export.sh — live QNAP deploy targets documented in
  DEPLOY_QNAP.md. Keep.
- Main docker-compose.yml, Dockerfile, env.example — live. Keep.
