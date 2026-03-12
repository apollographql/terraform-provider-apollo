# Apollo Terraform Provider Roadmap

## High-Level Feature Support

- [x] Provider foundation and shared GraphOS client: [feat-provider-foundation.md](./feat-provider-foundation.md)
- [ ] Graph variant management: [feat-graph-variants.md](./feat-graph-variants.md)
- [ ] Subgraph registration and routing metadata: [feat-subgraphs.md](./feat-subgraphs.md)
- [ ] Persisted query list management and linkage constraints: [feat-persisted-query-lists.md](./feat-persisted-query-lists.md)
- [ ] API key lifecycle management: [feat-api-keys.md](./feat-api-keys.md)
- [ ] Governance settings: [feat-governance.md](./feat-governance.md)
- [x] Read-only data sources: [feat-data-sources.md](./feat-data-sources.md)
- [x] Quality, testing, and release readiness: [feat-quality-and-release.md](./feat-quality-and-release.md)

## Current Snapshot

- [x] Terraform Plugin Framework scaffold is in place
- [x] Apollo provider/resource/data source schemas are registered
- [x] Generated provider/resource/data source docs exist
- [x] Example Terraform configurations exist for the scaffolded surface
- [x] Shared GraphOS Platform API client foundation is implemented
- [x] Shared import ID, normalization, capability, and diagnostic helpers are implemented
- [x] Checked-in capability matrix reflects documented versus inferred Platform API coverage
- [x] Unit tests cover provider config resolution, transport behavior, and mapped client methods
- [x] Optional gated integration tests exist for read-only and disposable PQL smoke coverage
- [x] `apollo_persisted_query_list` CRUD and import support are implemented
- [x] `apollo_persisted_query_list_link` CRUD and import support are implemented
- [x] `apollo_graph`, `apollo_graph_variant`, `apollo_subgraph`, and `apollo_persisted_query_list` data source read behavior is implemented
- [x] Data source outputs now match the current public Platform API schema instead of placeholder scaffold fields
- [x] `apollo_graph_variant` import/read/delete support is implemented around the deferred-create model
- [x] `apollo_subgraph` import/read/delete support is implemented around the deferred-create model
- [x] `apollo_graph_api_key` create/read/update/delete and import support are implemented
- [x] `apollo_subgraph_api_key` create/read/update/delete and import support are implemented
- [x] API key resources preserve one-time-visible key material in sensitive Terraform state for Terraform-created resources
- [x] Imported API key resources intentionally leave `key` unset because GraphOS does not re-expose secret values
- [x] Hermetic acceptance coverage exists for the implemented resources and data sources and is explicitly gated behind `TF_ACC=1`
- [x] CI validates lint/build, unit tests, doc generation, hermetic acceptance, and GoReleaser snapshot packaging
- [x] Live GraphOS smoke tests are separated into a manual gated workflow
- [x] Release metadata and registry-facing summaries now describe the Apollo provider instead of the scaffold
- [x] Standalone graph variant creation is explicitly deferred pending a schema-publish-backed model
- [x] Standalone subgraph creation is explicitly deferred pending a schema-publish-backed model
- [x] Persisted query list linkage uses the variant-scoped GraphOS mutations in the current public Platform API schema
- [x] Governance is explicitly deferred from the current MVP slice pending a separate product/API remapping pass

## Suggested Execution Order

- [x] Implement persisted query list resource support around the confirmed public API mutations
- [x] Implement data sources for brownfield discovery and module composition
- [x] Implement graph variant read/import and deletion support around the deferred-create model
- [x] Implement subgraph read/import and deletion support around the deferred-create model
- [x] Implement API key lifecycle handling with explicit sensitive-value behavior
- [ ] Revisit governance resources after API contracts are confirmed stable and the product model is remapped
- [x] Close testing, documentation, and release gaps for an MVP candidate
