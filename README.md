# Terraform Provider Apollo

This repository is an Apollo GraphOS Terraform provider derived from the [HashiCorp Terraform Plugin Framework scaffolding template](https://github.com/hashicorp/terraform-provider-scaffolding-framework).

The provider currently has a working GraphOS client plus implemented slices for persisted query lists, foundational data sources, brownfield graph variants and subgraphs, and graph/subgraph API keys. Governance remains deferred, along with any lifecycle gap where the GraphOS Platform API does not yet expose a safe Terraform contract.

## Current Status

- Implemented resources
  - `apollo_persisted_query_list`
  - `apollo_persisted_query_list_link`
  - `apollo_graph_api_key`
  - `apollo_subgraph_api_key`
  - `apollo_graph_variant` as brownfield import/read/delete only
  - `apollo_subgraph` as brownfield import/read/delete only
- Implemented data sources
  - `apollo_graph`
  - `apollo_graph_variant`
  - `apollo_subgraph`
  - `apollo_persisted_query_list`
- Deferred resources
  - `apollo_proposal_config`
  - `apollo_session_policy`
- Current boundaries
  - New graph variants are created by first subgraph publish, not by a standalone variant create mutation.
  - New subgraphs are created by schema publish, not by a standalone registration mutation.
  - Persisted query list metadata supports create/read/update/delete, but `description` is not returned by the readable API shape.
  - API key values are only available at create time and are stored in sensitive Terraform state for Terraform-created resources.
- Verification status
  - Hermetic acceptance tests are `TF_ACC`-gated and backed by local `httptest` GraphQL servers.
  - Live GraphOS smoke coverage exists for client operations and the implemented Terraform resources.
  - CI validates lint/build, unit tests, generated docs, hermetic acceptance, and GoReleaser snapshot packaging.

Implementation details and API mappings are tracked in [`docs/capability-map.md`](./docs/capability-map.md).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

## Building The Provider

Build the provider locally with:

```shell
go install .
```

## Using the provider

```terraform
terraform {
  required_providers {
    apollo = {
      source = "apollographql/apollo"
    }
  }
}

provider "apollo" {
  api_key = var.apollo_api_key
}
```

The provider currently supports GraphOS persisted query list CRUD plus variant link and unlink lifecycle, foundational read-only data sources, brownfield graph variant and subgraph management, and graph/subgraph API key lifecycle management. Governance resources remain deferred.

## Developing the Provider

Useful commands:

```shell
make fmt
make lint
make test
make testacc
make generate
```

### Test Modes

Default local test runs are fast unit and non-acceptance runs:

```shell
go test ./...
```

Hermetic acceptance tests are the canonical provider-level acceptance layer. They use local `httptest` GraphQL servers instead of a live GraphOS account and only run when `TF_ACC=1`:

```shell
TF_ACC=1 go test ./internal/provider/...
```

Generated docs and examples should stay in sync with the current provider contracts:

```shell
cd tools && go generate ./...
```

### Live GraphOS Integration

Live GraphOS smoke tests are separate from hermetic acceptance. They are optional, require real credentials, and are intentionally not part of the default PR path.

Read-only smoke test:

```shell
APOLLO_PLATFORM_INTEGRATION=1 \
APOLLO_KEY=service:graphos-key \
APOLLO_TEST_GRAPH_ID=inventory \
go test -run '^TestClientIntegrationReadOnly$' ./internal/provider/...
```

Optional mutating client smoke test:

```shell
APOLLO_PLATFORM_INTEGRATION=1 \
APOLLO_INTEGRATION_ALLOW_MUTATIONS=1 \
APOLLO_KEY=service:graphos-key \
APOLLO_TEST_GRAPH_ID=inventory \
APOLLO_TEST_VARIANT=current \
go test -run '^TestClientIntegrationPersistedQueryListSmoke$' ./internal/provider/...
```

Terraform resource smoke test covering the implemented mutable resources plus the brownfield-only subgraph and graph variant resources through disposable bootstrap fixtures:

```shell
APOLLO_PLATFORM_INTEGRATION=1 \
APOLLO_INTEGRATION_ALLOW_MUTATIONS=1 \
APOLLO_KEY=service:graphos-key \
APOLLO_TEST_GRAPH_ID=inventory \
go test -run '^TestTerraformIntegration' ./internal/provider/...
```

Environment variables:

- `TF_ACC=1`: enable hermetic acceptance tests backed by local `httptest` GraphQL servers.
- `APOLLO_PLATFORM_INTEGRATION=1`: enable live GraphOS smoke tests.
- `APOLLO_INTEGRATION_ALLOW_MUTATIONS=1`: allow disposable create/link/delete live coverage, including Terraform resource smoke tests.
- `APOLLO_KEY`: GraphOS API key for live smoke tests.
- `APOLLO_TEST_GRAPH_ID`: graph ref used by live smoke tests.
- `APOLLO_TEST_VARIANT`: optional existing variant used by targeted client smoke paths.
- `APOLLO_TEST_SUBGRAPH`: optional existing subgraph used by targeted client smoke paths.
- `APOLLO_GRAPHOS_ENDPOINT`: optional Platform API endpoint override.

### CI And Release

PR and push CI are hermetic and blocking. The main workflow runs:

- lint and build
- unit tests
- generated artifact verification
- `TF_ACC=1` hermetic acceptance against the supported Terraform matrix
- GoReleaser snapshot packaging validation

The manual `Live GraphOS Integration` workflow is reserved for real API smoke tests. It runs read-only client checks first, then mutating client smoke, then the live Terraform resource sweep when mutation coverage is enabled. Tagged `v*` releases use GoReleaser to publish signed artifacts for `registry.terraform.io/apollographql/apollo`.
