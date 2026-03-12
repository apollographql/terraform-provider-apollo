# AGENTS.md

## Scope

These instructions apply to the entire `terraform-provider` repository.

## Repo Shape

- Provider implementation lives in `internal/provider/`.
- Feature tracking lives in `specs/`.
- Generated and published docs live in `docs/`.
- Example Terraform configurations live in `examples/`.
- Provider capability notes live in `docs/capability-map.md`.

## Working Rules

- Keep changes aligned with the current public GraphOS Platform API behavior.
- Do not add checked-in schema snapshots to `specs/`; use the public Platform API docs/schema and the existing capability map instead.
- Reuse shared helpers for client access, import ID parsing, normalization, and diagnostics instead of duplicating patterns.
- Keep brownfield behavior explicit when GraphOS does not support a safe standalone create/update contract.

## New Terraform Resource Workflow

- Before implementing any new Terraform resource type, create a dedicated tracker file in `specs/` named `feat-<resource-or-feature>.md`.
- Use the existing tracker files in `specs/feat-*.md` as the template.
- The tracker must use markdown checklists and should follow the existing structure:
  - `## Support Status`
  - `## Tasks`
- Add the new tracker to `specs/roadmap.md` before starting implementation.
- Update the tracker as work lands; do not leave completed work unchecked.
- If the new resource changes API coverage or product boundaries, update `docs/capability-map.md` as part of the same change.

## Resource And Data Source Expectations

- New resources should include:
  - schema
  - client wiring
  - import behavior
  - normalization and diagnostics
  - unit tests
  - hermetic acceptance tests when the behavior is supported
  - docs and example configuration
- New data sources should include:
  - read behavior
  - not-found and permission diagnostics
  - unit tests
  - docs and example configuration

## Docs And Validation

- Keep `README.md`, `docs/`, `examples/`, and `specs/` aligned with the implemented provider behavior.
- Regenerate docs after contract changes with:
  - `cd tools && go generate ./...`
- Validate changes with:
  - `go test ./...`
  - `go build ./...`
- When relevant, also run:
  - `TF_ACC=1 go test ./internal/provider/...`
  - live GraphOS integration commands described in `README.md`

## Change Discipline

- Prefer small, decision-complete feature slices.
- Keep roadmap/tracker status current in the same change that implements the feature.
- Do not mark a feature complete while implementation, docs, or tests are still intentionally missing.
