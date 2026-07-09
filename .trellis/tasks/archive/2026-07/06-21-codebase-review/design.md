# Codebase Review Design

## Review Boundaries

This task produces review findings and recommendations. It does not change application behavior or refactor production code.

Primary areas:

- Crystal backend: app entry points, routes, handlers, storage, library, queue, plugin, upload, config, logging, utilities, migrations.
- Frontend and views: ECR templates, public JavaScript, LESS styles, static assets, frontend conventions from `.trellis/spec/frontend/`.
- Operations: Dockerfile, Compose files, Makefile, dependency manifests, environment examples, README and deployment docs.
- Tests: Crystal specs, fixtures, and any package scripts.

## Evidence Strategy

Use repository evidence before asking the user questions:

- Inspect config and manifests to identify supported commands.
- Read entry points and high-risk modules first: server setup, auth, upload, storage, plugin download/update paths, file/archive handling, HTTP client/proxy utilities.
- Search for risk patterns: shell execution, path traversal, unsafe file writes, missing auth checks, hardcoded secrets, unvalidated params, broad CORS, direct DOM HTML insertion, disabled TLS, TODO/FIXME markers.
- Run available checks with the repo's tooling where dependencies are present.

## Output Contract

Final output follows code-review format:

- Findings first, ordered by severity.
- Each finding includes a file and line reference.
- Include why it matters and a concrete suggested change.
- Then list open questions or assumptions.
- Then give a short validation summary.

## Compatibility

The review should not alter application files. Trellis task artifacts are the only planned writes.

## Rollback

If planning artifacts need rollback, remove or archive `.trellis/tasks/06-21-codebase-review/`. Do not touch application code.
