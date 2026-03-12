# Feature Tracker: API Keys

## Support Status

- [ ] Feature complete
- [x] `apollo_graph_api_key` resource is implemented
- [x] `apollo_subgraph_api_key` resource is implemented
- [x] Generated docs and example configurations exist
- [x] Graph API key lifecycle behavior is implemented
- [x] Subgraph API key lifecycle behavior is implemented
- [x] Sensitive one-time-visible key handling is implemented deliberately
- [x] Import behavior is defined for supported key resources
- [x] Unit and acceptance coverage exist
- [x] Graph API key roles are validated against the checked-in `UserPermission` enum
- [ ] Rotation, expiration management, and multi-target subgraph keys remain deferred

## Tasks

- [x] Confirm GraphOS API mapping for graph API keys
  - [x] Identify create mutation and response shape
  - [x] Identify read query and available metadata fields
  - [x] Identify update semantics for mutable metadata
  - [x] Identify revoke/delete semantics
  - [x] Document required credential scope for each operation
- [x] Implement `apollo_graph_api_key` resource behavior
  - [x] Implement `Create`
  - [x] Implement `Read`
  - [x] Implement `Update`
  - [x] Implement `Delete`
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Make `role` replacement-only
  - [x] Expose `created_at` and `last_used_at`
- [x] Confirm GraphOS API mapping for subgraph API keys
  - [x] Identify create mutation and response shape
  - [x] Identify read query and available metadata fields
  - [x] Identify update semantics for mutable metadata
  - [x] Identify revoke/delete semantics
  - [x] Document required credential scope for each operation
- [x] Implement `apollo_subgraph_api_key` resource behavior
  - [x] Implement `Create`
  - [x] Implement `Read`
  - [x] Implement `Update`
  - [x] Implement `Delete`
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Remove the placeholder `role` argument
  - [x] Enforce exactly one subgraph target tuple per resource
- [x] Implement sensitive-value handling rules
  - [x] Keep `key` as a sensitive create-time output that remains in Terraform state for Terraform-created resources
  - [x] Document the chosen tradeoff clearly
  - [x] Ensure the `key` attribute behavior matches the chosen tradeoff
  - [x] Ensure logs and diagnostics do not leak secret values
- [x] Implement import behavior
  - [x] Define import ID format for graph API keys
  - [x] Define import ID format for subgraph API keys
  - [x] Document that secret material is intentionally unavailable after import
- [x] Add API key test coverage
  - [x] Add unit tests for sensitive-value handling helpers
  - [x] Add client tests for create, read, update, delete, auth, and malformed responses
  - [x] Add Terraform resource tests for create and read
  - [x] Add Terraform resource tests for update
  - [x] Add Terraform resource tests for revoke/delete behavior
  - [x] Add Terraform resource tests for import semantics
- [x] Review docs after implementation
  - [x] Add realistic example configurations
  - [x] Add import examples
  - [x] Document secret-handling behavior and permissions clearly

## Remaining Gaps

- [ ] Feature complete
  - [ ] Add explicit rotation and expiration lifecycle support if that belongs in Terraform
  - [ ] Decide whether multi-target subgraph keys should become a separate resource model instead of being rejected
