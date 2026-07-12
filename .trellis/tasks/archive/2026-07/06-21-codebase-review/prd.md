# Codebase Review

## Goal

Produce a formal, evidence-backed review of the repository and recommend the highest-value changes. The review should prioritize bugs, security risks, maintainability issues, test gaps, and operational risks over style-only feedback.

## Confirmed Facts

- The repository is a single-repo application with Crystal source under `src/`, migrations under `migration/`, specs under `spec/`, server-rendered ECR views under `src/views/`, and browser assets under `public/`.
- Project tooling includes `shard.yml`, `Makefile`, `Dockerfile`, Docker Compose files, `package.json`, and `package-lock.json`.
- Trellis currently has only a `frontend` spec layer, so review guidance must combine those local frontend rules with direct source inspection for backend and operational code.
- The user requested a formal Trellis task and codebase review with suggested changes.

## Requirements

- Review the codebase without modifying production application code.
- Inspect representative backend, frontend, view, configuration, Docker, migration, and test files.
- Run locally available quality checks when feasible, using existing project commands where possible.
- Report findings in code-review style, ordered by severity and grounded in file and line references.
- Include pragmatic change suggestions and note any test or validation gaps.
- Preserve unrelated uncommitted user changes.

## Acceptance Criteria

- [x] Planning artifacts describe the review scope and execution approach.
- [x] Repository structure, major entry points, and quality tooling are inspected.
- [x] Available tests or checks are run, or blockers are documented.
- [x] Final response lists actionable findings before any summary.
- [x] Each finding includes the affected file and line when applicable.
- [x] Suggestions are scoped to the current repository and do not require unsupported assumptions.

## Out of Scope

- Implementing the suggested code changes in this task.
- Production environment testing or live API calls.
- External paid security scanning or dependency intelligence services.
- Broad product redesign unrelated to code quality.

## Open Questions

- None blocking. Assume the review should cover the whole repository at static-analysis depth plus locally runnable checks.
