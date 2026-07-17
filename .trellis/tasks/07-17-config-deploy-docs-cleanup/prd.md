# Configuration deployment and documentation cleanup

## Goal

Make configuration and deployment behavior truthful, reproducible, and safe for
the documented Docker, QNAP, and direct-binary workflows.

## Requirements

- Classify every configuration field as implemented, deprecated, or removable;
  wire supported fields into runtime behavior and reject or document invalid
  combinations.
- Align README, QNAP, Docker Hub, frontend, Makefile, environment example, and
  Trellis specifications with the Go-only implementation.
- Restore or intentionally remove the API documentation page; its OpenAPI route,
  specification file, and embedded ReDoc asset must form a complete contract.
- Make the default Compose workflow validate with actionable defaults or fail
  early with clear required-variable messages.
- Remove obsolete Compose schema fields and stale Crystal paths/instructions.
- Decide whether the dead top-level Crystal specs should be archived or removed,
  preserving any compatibility contracts that still matter.
- Validate persistence mounts, container port behavior, multi-architecture
  builds, image provenance, runtime user/volume permissions, and upgrade/rollback
  instructions.
- Improve Docker layer caching and version pinning where it materially improves
  reproducibility without making NAS deployment harder.

## Acceptance Criteria

- [ ] Every documented configuration key has a tested runtime effect or is clearly
      marked deprecated/removed.
- [ ] `docker compose config` succeeds for the default example and all maintained
      QNAP variants without obsolete-schema warnings.
- [ ] A clean direct-binary and container startup use documented persistent paths
      and ports.
- [ ] README and deployment commands reference only current Go paths and behavior.
- [ ] The API documentation link renders a real current specification or is no
      longer presented as a working feature.
- [ ] `make all` behavior matches its documentation and the CI quality gate.
- [ ] Historical Crystal artifacts no longer look like an executable current test
      suite.
- [ ] Build, upgrade, backup, rollback, and first-admin credential handling are
      documented and smoke-tested with disposable data.

## Notes

- Dependency: none. It can start independently, but final CI command references
  should be synchronized with `07-17-test-ci-baseline`.
- Constraint: do not switch the scratch image away from current pure-Go behavior
  or force a non-root runtime until NAS bind-mount permissions are validated.
- Evidence: several loaded config fields are unused; the default Compose example
  fails with empty values; QNAP docs and specs contain stale Crystal guidance.
