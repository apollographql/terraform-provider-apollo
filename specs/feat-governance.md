# Feature Tracker: Governance

## Support Status

- [ ] Feature complete
- [x] `apollo_proposal_config` resource schema is scaffolded
- [x] `apollo_session_policy` resource schema is scaffolded
- [x] Generated docs and example configurations exist
- [ ] Proposal governance behavior is implemented
- [ ] Session policy behavior is implemented
- [ ] Import behavior is defined for supported governance resources
- [ ] Unit and acceptance coverage exist

## Tasks

- [ ] Confirm GraphOS API mapping for proposal governance
  - [ ] Identify the stable API surface for proposal configuration
  - [ ] Confirm whether proposal settings are graph-level or variant-level
  - [ ] Identify mutable versus immutable fields
  - [ ] Document required credential scope for each operation
- [ ] Implement `apollo_proposal_config` resource behavior
  - [ ] Implement `Create` if the API models creation separately
  - [ ] Implement `Read`
  - [ ] Implement `Update`
  - [ ] Implement `Delete` or explicit reset semantics if delete is not a real concept
  - [ ] Map Terraform fields to GraphOS API fields
- [ ] Implement proposal config import and normalization
  - [ ] Define import ID format
  - [ ] Validate import ID errors cleanly
  - [ ] Normalize configuration values returned by GraphOS
  - [ ] Remove state cleanly when the remote object no longer exists or is disabled
- [ ] Confirm GraphOS API mapping for session policy
  - [ ] Identify the stable API surface for organization session settings
  - [ ] Confirm ownership and scope rules
  - [ ] Identify mutable versus immutable fields
  - [ ] Document required credential scope for each operation
- [ ] Implement `apollo_session_policy` resource behavior
  - [ ] Implement `Create` if the API models creation separately
  - [ ] Implement `Read`
  - [ ] Implement `Update`
  - [ ] Implement `Delete` or explicit reset semantics if delete is not a real concept
  - [ ] Map Terraform fields to GraphOS API fields
- [ ] Implement session policy import and normalization
  - [ ] Define import ID format
  - [ ] Validate import ID errors cleanly
  - [ ] Normalize session policy values returned by GraphOS
  - [ ] Remove state cleanly when the remote object no longer exists
- [ ] Add governance test coverage
  - [ ] Add unit tests for proposal config helpers
  - [ ] Add unit tests for session policy helpers
  - [ ] Add acceptance tests for supported governance flows
  - [ ] Add acceptance tests for import semantics
- [ ] Review docs after implementation
  - [ ] Add realistic example configurations
  - [ ] Add import examples
  - [ ] Document required permissions and API stability notes
