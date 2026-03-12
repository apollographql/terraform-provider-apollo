# Apollo GraphOS Platform API Capability Map

This document records the provider's GraphOS Platform API mapping. It combines Apollo's narrative docs with the public Studio Platform API schema where the narrative docs stop short of naming the concrete GraphQL fields.

## Confidence Levels

- `documented`
  The field or operation name is visible in Apollo's public docs.
- `public_schema`
  The field or operation name is visible in Apollo's public Studio Platform API schema.
- `inferred`
  The product capability is documented, but the exact field path needs a follow-up check in the Studio Platform API explorer.

## Shared Rules

- GraphOS Platform API is the control-plane source of truth for this provider slice.
- Brownfield-friendly import IDs stay aligned to GraphOS nouns.
- `apollo_graph_variant` does not get standalone create behavior in the first implementation slice.
- `apollo_subgraph` does not get standalone create behavior in the first implementation slice.
- Persisted query list publication remains out of scope even though Apollo supports it outside this provider slice.

## Capability Matrix

| Domain | Provider Surface | Operation | Platform API Entry Point | Required Inputs | Expected Credential Scope | Terraform Viability | Confidence | Import ID | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Graph metadata | `apollo_graph` data source | Read graph metadata | `Query.graph(id)` | `graphId` | Graph read or broader org access | supported | public_schema | `graph_id` | Data source exposes `id`, `name`, `title`, and `variant_names`. |
| Graph variant metadata | `apollo_graph_variant` resource, `apollo_graph_variant` data source | Read variant metadata | `Query.graph(id).variant(name)` | `graphId`, `variantName` | Graph read or broader org access | supported | public_schema | `graph_id:variant` | Data source exposes the variant `id` plus nested subgraph `name`, `routing_url`, and `revision`. |
| Graph variant lifecycle | `apollo_graph_variant` resource | Create new variant | `Mutation.graph.publishSubgraph` or `Mutation.graph.publishSubgraphs` to a new variant | `graphId`, `variant`, first subgraph inputs, schema SDL, routing URL | Subgraph publish authority on the target graph | deferred | documented | `graph_id:variant` | Public docs describe new variants being created by publishing the first subgraph, not via a standalone variant-create mutation. |
| Graph variant lifecycle | `apollo_graph_variant` resource | Delete variant | `Mutation.graph(id).variant(name).delete` | `graphId`, `variantName` | Graph admin or broader org access | supported | public_schema | `graph_id:variant` | The resource uses brownfield import, read, and delete only. Standalone create remains deferred to first-subgraph publish. |
| Subgraph metadata | `apollo_subgraph` resource, `apollo_subgraph` data source | Read subgraph metadata | `Query.graph(id).variant(name).subgraphs` | `graphId`, `variantName` | Graph read or broader org access | supported | public_schema | `graph_id:variant:subgraph_name` | The resource and data source both use a synthetic `id` of `graph_id:variant:name` because the schema does not expose a dedicated subgraph object ID. |
| Subgraph lifecycle | `apollo_subgraph` resource | Create subgraph or bootstrap variant | `Mutation.graph.publishSubgraph` | `graphId`, `variant`, `name`, `url`, `activePartialSchema` | Subgraph publish authority | deferred | documented | `graph_id:variant:subgraph_name` | Public docs say first publish requires SDL and routing URL, so the brownfield Terraform resource does not expose create yet. |
| Subgraph lifecycle | `apollo_subgraph` resource | Remove subgraph | `Mutation.graph(id).removeImplementingServiceAndTriggerComposition` | `graphId`, `graphVariant`, `name`, `dryRun` | Graph admin or broader org access | supported | public_schema | `graph_id:variant:subgraph_name` | The resource is import/read/delete only. Destroy fails on confirmed composition errors, but tolerates transient GraphOS delete responses when the subgraph is already gone on follow-up read. |
| Persisted query list metadata | `apollo_persisted_query_list` resource, `apollo_persisted_query_list` data source | Read one persisted query list | `Query.graph(id).persistedQueryList(id)` | `graphId`, `id` | Graph or org access for persisted query list management | supported | public_schema | `graph_id:persisted_query_list_id` | The current public schema models PQLs as graph-scoped objects, not operation collections. |
| Persisted query list metadata | `apollo_persisted_query_list_link` resource | Read the variant's linked persisted query list | `Query.graph(id).variant(name).persistedQueryList` | `graphId`, `variantName` | Graph or org access for persisted query list management | supported | public_schema | `persisted_query_list_id:graph_id:variant` | A variant can expose at most one linked PQL through this field. |
| Persisted query list lifecycle | `apollo_persisted_query_list` resource | Create persisted query list metadata | `Mutation.graph(id).createPersistedQueryList` | `graphId`, `name`, optional `description` | Graph or org access for persisted query list management | supported | public_schema | `graph_id:persisted_query_list_id` | The API accepts `description`, but the current readable PQL type does not expose it. |
| Persisted query list lifecycle | foundation client | List persisted query lists for a graph | no public schema field | `graphId` | Graph or org access for persisted query list management | deferred | public_schema | n/a | The current public schema does not expose a plural `persistedQueryLists` field, so the provider currently has no list operation to wire. |
| Persisted query list lifecycle | `apollo_persisted_query_list` resource | Update persisted query list metadata | `Mutation.graph(id).persistedQueryList(id).updateMetadata` | `graphId`, `id`, optional `name`, optional `description` | Graph or org access for persisted query list management | supported | public_schema | `graph_id:persisted_query_list_id` | Terraform stores `description` from config/state because the current schema does not return it on read. |
| Persisted query list lifecycle | `apollo_persisted_query_list` resource | Delete persisted query list | `Mutation.graph(id).persistedQueryList(id).delete` | `graphId`, `id` | Graph or org access for persisted query list management | supported | public_schema | `graph_id:persisted_query_list_id` | Delete can fail with `CannotDeleteLinkedPersistedQueryListError` while variants are still attached. |
| Persisted query list linking | `apollo_persisted_query_list_link` resource | Link persisted query list to variant | `Mutation.graph(id).variant(name).linkPersistedQueryList` | `graphId`, `variantName`, `persistedQueryListId` | Graph admin or broader org access | supported | public_schema | `persisted_query_list_id:graph_id:variant` | The current public schema shows the link on `GraphVariantMutation`, not on operation collections. |
| Persisted query list linking | `apollo_persisted_query_list_link` resource | Unlink persisted query list from variant | `Mutation.graph(id).variant(name).unlinkPersistedQueryList` | `graphId`, `variantName` | Graph admin or broader org access | supported | public_schema | `persisted_query_list_id:graph_id:variant` | The current public schema omits the concrete unlink success type from the union, so the client verifies success with a follow-up read. |
| Graph API key lifecycle | `apollo_graph_api_key` resource | Create graph API key | `Mutation.graph(id).newKey` | `graphId`, `keyName`, `role` | Org Admin or Graph Admin access for graph API key management | supported | public_schema | `graph_id:key_id` | Create returns the one-time-visible `token`. Terraform stores that sensitive value in state for Terraform-created resources only. |
| Graph API key lifecycle | `apollo_graph_api_key` resource | Read graph API key metadata | `Query.graph(id).apiKeys` | `graphId` | Org Admin or Graph Admin access for graph API key management | supported | public_schema | `graph_id:key_id` | The client filters the graph-scoped key list by remote key ID. Read never requests `token`, because GraphOS does not re-expose it after creation. |
| Graph API key lifecycle | `apollo_graph_api_key` resource | Rename graph API key | `Mutation.graph(id).renameKey` | `graphId`, `id`, `newKeyName` | Org Admin or Graph Admin access for graph API key management | supported | public_schema | `graph_id:key_id` | Graph API key roles cannot be changed after creation, so `role` is replacement-only in Terraform. |
| Graph API key lifecycle | `apollo_graph_api_key` resource | Delete graph API key | `Mutation.graph(id).removeKey` | `graphId`, `id` | Org Admin or Graph Admin access for graph API key management | supported | public_schema | `graph_id:key_id` | Delete treats remote-not-found as success after a pre-read. |
| Subgraph API key lifecycle | `apollo_subgraph_api_key` resource | Create subgraph API key | `Mutation.organization(id).createKey` after `Query.graph(id).account.id` | `organizationId`, `name`, `type=SUBGRAPH`, `resources.subgraphs[]` | Org Admin or Graph Admin access for subgraph API key management | supported | public_schema | `graph_id:variant:subgraph_name:key_id` | Create returns the one-time-visible `token`. This slice supports exactly one subgraph target tuple per resource. |
| Subgraph API key lifecycle | `apollo_subgraph_api_key` resource | Read subgraph API key metadata | `Query.organization(id).apiKey(keyId)` after `Query.graph(id).account.id` | `organizationId`, `keyId` | Org Admin or Graph Admin access for subgraph API key management | supported | public_schema | `graph_id:variant:subgraph_name:key_id` | Read never requests `token`, because GraphOS does not re-expose it after creation. Multi-target remote keys are rejected as unsupported in this slice. |
| Subgraph API key lifecycle | `apollo_subgraph_api_key` resource | Rename subgraph API key | `Mutation.organization(id).renameKey` after `Query.graph(id).account.id` | `organizationId`, `keyId`, `name` | Org Admin or Graph Admin access for subgraph API key management | supported | public_schema | `graph_id:variant:subgraph_name:key_id` | The resource maps one Terraform object to exactly one `graph_id:variant:subgraph_name` target tuple. |
| Subgraph API key lifecycle | `apollo_subgraph_api_key` resource | Delete subgraph API key | `Mutation.organization(id).deleteKey` after `Query.graph(id).account.id` | `organizationId`, `keyId` | Org Admin or Graph Admin access for subgraph API key management | supported | public_schema | `graph_id:variant:subgraph_name:key_id` | Delete treats remote-not-found as success after a pre-read. Rotation and expiration remain out of scope. |

## Current Provider Defaults

- `apollo_graph_variant`
  - `create`: deferred
  - `read`: implemented through `graph(id).variant(name)`
  - `update`: deferred until the product model is explicit
  - `delete`: implemented through `graph(id).variant(name).delete`
  - `id`: stores the remote GraphOS `GraphVariant.id`, while import uses `graph_id:variant`
- `apollo_subgraph`
  - `create`: deferred because bootstrap requires schema publish inputs
  - `read`: implemented through `graph(id).variant(name).subgraphs`
  - `update`: deferred until a schema-backed update model is approved
  - `delete`: implemented through `graph(id).removeImplementingServiceAndTriggerComposition`
  - `routing_url` and `revision`: computed outputs only in the brownfield resource slice
- `apollo_persisted_query_list`
  - `create`, `read`, `update`, `delete`: implemented through the graph-scoped persisted query list fields above
  - `description`: write-only in the current schema, so import/read cannot hydrate it from GraphOS
  - operation publication: out of scope
- `apollo_persisted_query_list_link`
  - standalone lifecycle: implemented
  - read path: verifies the link through `graph(id).variant(name).persistedQueryList`
  - graph-level linking: out of scope because the current schema models the attachment on variants
- `apollo_graph_api_key`
  - `create`, `read`, `update`, `delete`: implemented through the graph-scoped key fields above
  - `key`: sensitive create-time output only; refresh and import cannot rehydrate it from GraphOS
  - `role`: validated against the local `UserPermission` enum and replacement-only after create
- `apollo_subgraph_api_key`
  - `create`, `read`, `update`, `delete`: implemented through the org-scoped key fields above after resolving the graph owner organization
  - `key`: sensitive create-time output only; refresh and import cannot rehydrate it from GraphOS
  - multi-target keys: out of scope in this slice and rejected with a deliberate diagnostic
  - expiration and rotation: deferred to a future slice

## Sources

- [GraphOS Platform API overview](https://www.apollographql.com/docs/graphos/platform/platform-api)
- [Publish schemas with the Platform API](https://www.apollographql.com/docs/graphos/platform/schema-management/delivery/publishing/platform-api)
- [Add and manage variants](https://www.apollographql.com/docs/graphos/platform/graph-management/variants)
- [Persisted queries and PQL management](https://www.apollographql.com/docs/graphos/platform/security/persisted-queries)
- [API keys and roles](https://www.apollographql.com/docs/graphos/platform/access-management/api-keys)
- [Apollo Platform API public schema in Studio](https://studio.apollographql.com/public/apollo-platform/home?variant=main)
