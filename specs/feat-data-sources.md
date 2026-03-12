# Feature Tracker: Data Sources

## Support Status

- [x] Feature complete
- [x] `apollo_graph` data source read behavior is implemented
- [x] `apollo_graph_variant` data source read behavior is implemented
- [x] `apollo_subgraph` data source read behavior is implemented
- [x] `apollo_persisted_query_list` data source read behavior is implemented
- [x] Generated docs and example configurations exist
- [x] Read-only GraphOS lookup mappings are documented in the capability matrix
- [x] Shared client methods exist for graph, variant, subgraph, and PQL metadata lookup
- [x] Data source diagnostics are hardened for brownfield discovery use cases
- [x] Unit and acceptance coverage exist

## Tasks

- [x] Confirm GraphOS API mapping for read-only metadata lookups
  - [x] Identify graph metadata query and response shape
  - [x] Identify graph variant metadata query and response shape
  - [x] Identify subgraph metadata query and response shape
  - [x] Identify persisted query list metadata query and response shape
  - [x] Document required credential scope for each lookup
- [x] Implement `apollo_graph` data source behavior
  - [x] Implement `Read`
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Handle not-found errors clearly
- [x] Implement `apollo_graph_variant` data source behavior
  - [x] Implement `Read`
  - [x] Remove the placeholder `description` field that has no schema backing
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Handle not-found errors clearly
- [x] Implement `apollo_subgraph` data source behavior
  - [x] Implement `Read`
  - [x] Expose the synthetic `graph_id:variant:name` identity
  - [x] Map Terraform fields to GraphOS API fields
  - [x] Handle not-found errors clearly
- [x] Implement `apollo_persisted_query_list` data source behavior
  - [x] Implement `Read`
  - [x] Lock the identity to `graph_id` plus `id` based on the current public Platform API schema
  - [x] Expose linked variants as nested objects
  - [x] Handle not-found errors clearly
- [x] Add data source test coverage
  - [x] Add unit tests for shared lookup helper behavior
  - [x] Add acceptance tests for graph lookups
  - [x] Add acceptance tests for graph variant lookups
  - [x] Add acceptance tests for subgraph lookups
  - [x] Add acceptance tests for persisted query list lookups
- [x] Review docs after implementation
  - [x] Add realistic example configurations
  - [x] Document lookup expectations and permissions
