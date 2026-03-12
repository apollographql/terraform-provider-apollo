# Feature Tracker: Graph Variants

## Support Status

- [ ] Feature complete
- [x] `apollo_graph_variant` resource schema is aligned to the checked-in GraphOS schema
- [x] Generated docs and example configuration exist
- [x] Graph variant API mapping is documented in the capability matrix
- [x] Standalone graph variant creation is explicitly deferred pending a schema-publish-backed model
- [ ] Full graph variant lifecycle support is implemented against the GraphOS Platform API
- [x] Import support for `graph_id:variant` is implemented
- [x] Variant-specific normalization is implemented
- [x] Unit and acceptance coverage exist

## Tasks

- [x] Confirm GraphOS API mapping for graph variants
  - [x] Identify create mutation and required input fields
  - [x] Identify read query and response shape
  - [x] Identify update mutation and mutable fields
  - [x] Identify delete mutation and not-found semantics
  - [x] Document required credential scope for each operation
- [ ] Implement `apollo_graph_variant` resource behavior
  - [x] Implement `Create` with the deferred-create model documented explicitly
  - [x] Implement `Read`
  - [x] Implement `Update` as a defensive replacement-only provider bug path
  - [x] Implement `Delete`
  - [x] Map Terraform state fields to GraphOS API fields
- [ ] Implement graph variant import and normalization
  - [x] Parse import ID format `graph_id:variant`
  - [x] Validate import ID errors cleanly
  - [x] Normalize identifiers returned by GraphOS
  - [x] Prevent drift from semantically equivalent values
  - [x] Remove state cleanly when the remote object no longer exists
- [ ] Add graph variant test coverage
  - [x] Add unit tests for model and helper behavior
  - [x] Add acceptance tests for create and read
  - [x] Add acceptance tests for update
  - [x] Add acceptance tests for import
  - [x] Add acceptance tests for delete and not-found handling
- [ ] Review docs after implementation
  - [x] Add realistic example configurations
  - [x] Add import example
  - [x] Document required permissions

## Remaining Gaps

- [ ] Feature complete
  - [ ] Expand beyond brownfield import/read/delete if Apollo adds a clean standalone variant create model
  - [ ] Decide whether mutable variant settings like URL, README, or federation version belong in this resource or separate resources
