# Feature Tracker: Quality And Release

## Support Status

- [x] Feature complete
- [x] Provider metadata, config resolution, client transport, and mapped client methods are unit tested
- [x] `go build ./...` succeeds for the implemented provider
- [x] `go generate ./...` succeeds from `tools/`
- [x] Optional gated integration tests exist for read-only and disposable PQL smoke coverage
- [x] Resource unit coverage exists for implemented behavior
- [x] Acceptance test coverage exists for supported features
- [x] CI expectations are documented and automated
- [x] Release and registry readiness are complete

## Tasks

- [x] Expand automated test coverage
  - [x] Add unit tests for the shared GraphOS client
  - [x] Add unit tests for import ID parsing helpers
  - [x] Add unit tests for normalization helpers
  - [x] Add unit tests for each implemented resource
  - [x] Add unit tests for each implemented data source
- [x] Build acceptance test infrastructure
  - [x] Define required environment variables for acceptance tests
  - [x] Define test fixture strategy for GraphOS resources
  - [x] Add provider factory and shared acceptance helpers
  - [x] Add acceptance coverage for each implemented feature area
  - [x] Document test isolation and cleanup expectations
- [x] Keep generated artifacts in sync
  - [x] Regenerate docs after schema changes
  - [x] Keep examples aligned with implemented behavior
  - [x] Review `docs/` output for accuracy after each feature lands
- [x] Prepare CI workflows
  - [x] Add formatting and lint checks
  - [x] Add unit test execution
  - [x] Add documentation generation verification
  - [x] Add acceptance test workflow strategy
- [x] Prepare release metadata and packaging
  - [x] Review `.goreleaser.yml` for final binary naming and release behavior
  - [x] Review `terraform-registry-manifest.json`
  - [x] Review `META.d/_summary.yaml`
  - [x] Confirm provider address and publishing flow
  - [x] Document versioning and changelog expectations
