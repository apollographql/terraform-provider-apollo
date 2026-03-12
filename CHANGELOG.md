## 0.1.0 (Unreleased)

FEATURES:

- Implemented Apollo GraphOS persisted query list CRUD plus variant-scoped link and unlink lifecycle.
- Implemented foundational read-only data sources for graphs, variants, subgraphs, and persisted query lists.
- Implemented brownfield graph variant and subgraph resources with import/read/delete contracts.
- Implemented graph and subgraph API key lifecycle management with one-time-visible key handling.

QUALITY:

- Added `TF_ACC`-gated hermetic acceptance coverage and shared acceptance helpers for the implemented provider surface.
- Added CI jobs for lint/build, unit tests, generate verification, hermetic acceptance, and GoReleaser snapshot validation.
- Added a manual live GraphOS integration workflow for read-only smoke tests and optional disposable PQL mutation coverage.

NOTES:

- Governance resources remain deferred pending a separate API remapping and product-contract pass.
- Standalone variant and subgraph creation remain deferred until a schema-publish-backed Terraform model is implemented.
