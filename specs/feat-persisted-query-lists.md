# Feature Tracker: Persisted Query Lists

## Support Status

- [ ] Feature complete
- [x] `apollo_persisted_query_list` resource is implemented
- [x] `apollo_persisted_query_list_link` resource is implemented
- [x] Generated docs and example configurations exist
- [x] Persisted query list and linkage API mapping is documented in the capability matrix
- [x] Shared client methods exist for PQL CRUD and variant linkage flows
- [x] Persisted query list CRUD is implemented
- [x] Standalone persisted query list linkage CRUD is implemented
- [x] Import support is implemented for `apollo_persisted_query_list`
- [x] Import support is implemented for `apollo_persisted_query_list_link`
- [x] Unit and Terraform resource coverage exist for the implemented PQL slice
- [x] Public API revalidation uses the current public Platform API schema as the source of truth

## Tasks

- [x] Confirm GraphOS API mapping for persisted query lists
  - [x] Identify graph-scoped create mutation and required fields
  - [x] Identify graph-scoped read query and response shape
  - [x] Identify update mutation and mutable fields
  - [x] Identify delete mutation semantics
  - [x] Record that the current public Platform API schema does not expose a plural PQL list field
  - [x] Document required credential scope for each operation
- [x] Implement `apollo_persisted_query_list` resource behavior
  - [x] Implement `Create`
  - [x] Implement `Read`
  - [x] Implement `Update`
  - [x] Implement `Delete`
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Require `graph_id` in the resource identity
  - [x] Preserve `description` from config/state because the current readable schema does not expose it
- [x] Implement persisted query list import and normalization
  - [x] Parse import ID format `graph_id:persisted_query_list_id`
  - [x] Validate import ID errors cleanly
  - [x] Normalize identifiers and descriptive fields
  - [x] Remove state cleanly when the remote object no longer exists
- [x] Confirm GraphOS API mapping for persisted query list linkage
  - [x] Identify the current public Platform API schema's variant link mutation surface
  - [x] Identify the current public Platform API schema's read query and response shape
  - [x] Identify the current public Platform API schema's unlink/delete surface
  - [x] Confirm that graph-level linking is not the current product model
  - [x] Document required credential scope for each operation
- [x] Implement `apollo_persisted_query_list_link` lifecycle
  - [x] Implement `Create`
  - [x] Implement `Read`
  - [x] Guard `Update` as replacement-only provider-bug behavior
  - [x] Implement `Delete`
  - [x] Implement `ImportState`
  - [x] Require `variant`
  - [x] Use `persisted_query_list_id:graph_id:variant` as the link identity
- [x] Add persisted query list test coverage for the implemented slice
  - [x] Add unit tests for shared PQL client helpers
  - [x] Add Terraform resource tests for PQL create and read
  - [x] Add Terraform resource tests for import and delete behavior
  - [x] Add Terraform resource tests for variant link create, replace, and import behavior
- [x] Review docs after implementation
  - [x] Add realistic example configuration for `apollo_persisted_query_list`
  - [x] Add a real link example
  - [x] Document the graph-scoped PQL identity and write-only description behavior
