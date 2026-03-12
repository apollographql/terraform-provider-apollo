# Feature Tracker: Subgraphs

## Support Status

- [ ] Feature complete
- [x] `apollo_subgraph` resource schema is implemented
- [x] Generated docs and example configuration exist
- [x] Subgraph API mapping is documented in the capability matrix
- [x] Standalone subgraph creation is explicitly deferred pending a schema-publish-backed model
- [x] Brownfield import, read, and delete support is implemented
- [x] Import support for `graph_id:variant:subgraph_name` is implemented
- [x] Subgraph-specific normalization is implemented
- [x] Unit and acceptance coverage exist
- [x] `routing_url` and `revision` are exposed as computed metadata only
- [ ] Schema-publish-backed create and mutable routing/schema workflows remain deferred

## Tasks

- [x] Confirm GraphOS API mapping for subgraphs
  - [x] Identify create/register mutation and required fields
  - [x] Identify read query and response shape
  - [x] Identify that no safe standalone routing-metadata update exists in this slice
  - [x] Identify delete/removal mutation semantics
  - [x] Document the boundary between Terraform-managed subgraph metadata and Rover-managed schema publishing
  - [x] Document required credential scope for each operation
- [x] Implement `apollo_subgraph` resource behavior
  - [x] Implement `Create` with the deferred-create model documented explicitly
  - [x] Implement `Read`
  - [x] Implement `Update` as a defensive provider-bug path
  - [x] Implement `Delete`
  - [x] Map `graph_id`, `variant`, and `name` to GraphOS identity fields
  - [x] Expose `routing_url` and `revision` as computed GraphOS metadata
- [x] Implement subgraph import and normalization
  - [x] Parse import ID format `graph_id:variant:subgraph_name`
  - [x] Validate import ID errors cleanly
  - [x] Normalize routing URLs and identifiers when GraphOS returns canonical forms
  - [x] Remove state cleanly when the remote object no longer exists
- [x] Add subgraph test coverage
  - [x] Add unit tests for model and helper behavior
  - [x] Add acceptance tests for deferred create and read
  - [x] Add acceptance tests for import
  - [x] Add acceptance tests for delete and not-found handling
  - [x] Add acceptance tests for delete authentication, permission, and composition-failure behavior
- [x] Review docs after implementation
  - [x] Add realistic example configurations
  - [x] Add import example
  - [x] Document the Terraform versus Rover boundary explicitly
- [ ] Future subgraph lifecycle work
  - [ ] Design a schema-publish-backed create flow around `publishSubgraph`
  - [ ] Decide whether any mutable routing/schema update contract belongs in Terraform
