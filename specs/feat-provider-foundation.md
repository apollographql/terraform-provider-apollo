# Feature Tracker: Provider Foundation

## Support Status

- [x] Feature complete
- [x] Apollo module path, provider address, and project metadata are scaffolded
- [x] Provider schema includes `api_key` and `endpoint`
- [x] Resource and data source registrations reflect the current PRD scope
- [x] Generated docs and examples match the scaffolded provider surface
- [x] Typed GraphOS Platform API client is implemented
- [x] Shared import parsing and normalization helpers are implemented
- [x] Permission-aware diagnostics are implemented
- [x] Provider configuration and client behavior are fully tested
- [x] Capability-state mappings are checked in for deferred versus supported behavior

## Tasks

- [x] Rebrand the HashiCorp scaffold as the Apollo provider
  - [x] Update `go.mod` module path
  - [x] Update `main.go` provider address
  - [x] Remove template action/function/ephemeral scaffolding
  - [x] Replace template docs and examples with Apollo-specific content
- [x] Add initial provider package structure
  - [x] Register Apollo resources
  - [x] Register Apollo data sources
  - [x] Add a minimal provider metadata test
  - [x] Add a placeholder client type for future API work
- [x] Implement the shared GraphOS client
  - [x] Define request and response models for GraphOS Platform API operations
  - [x] Add GraphQL request execution helpers
  - [x] Attach provider user-agent metadata to outbound requests
  - [x] Support configurable endpoint overrides cleanly
  - [x] Normalize network and API errors into Terraform diagnostics
  - [x] Add retry and timeout behavior for transient failures
- [x] Implement shared provider behavior
  - [x] Centralize import ID parsing helpers reused by resources
  - [x] Centralize semantic normalization helpers reused by resources
  - [x] Define common patterns for not-found handling
  - [x] Define common patterns for immutable identity attributes
  - [x] Define common permission error messaging by credential scope
- [x] Harden provider configuration
  - [x] Add tests for explicit `api_key` configuration
  - [x] Add tests for `APOLLO_KEY` fallback behavior
  - [x] Add tests for explicit `endpoint` configuration
  - [x] Add tests for `APOLLO_GRAPHOS_ENDPOINT` fallback behavior
  - [x] Add tests for unknown or missing provider values
- [x] Document provider-wide implementation rules
  - [x] Document import ID conventions
  - [x] Document sensitive-value handling principles
  - [x] Document credential scope expectations
  - [x] Document normalization expectations
