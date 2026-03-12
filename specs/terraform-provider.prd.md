# PRD: Apollo Terraform Provider

## Overview

This PRD proposes a new Terraform provider for Apollo GraphOS. The provider would let teams manage GraphOS control-plane configuration through infrastructure as code, using the Terraform Plugin Framework and the GraphOS Platform API.

GraphOS is already used for schema governance, subgraph composition, persisted query management, access control, and other graph-level workflows. At the same time, enterprise platform teams expect important SaaS control planes to show up in Terraform so they can be provisioned, reviewed, standardized, and audited alongside the rest of their infrastructure.

Apollo already has the raw pieces for this. Studio is the main UI, Rover is the standard CLI, and the GraphOS Platform API exists for automation. A Terraform provider would turn that API surface into a stable IaC integration for the parts of GraphOS that belong in long-lived desired state.

The provider should not try to wrap every Rover workflow. It should focus on durable control-plane objects and settings such as variants, subgraphs, persisted query lists, API keys, and governance configuration. That is a much better fit for Terraform than release-time or workflow-oriented operations.

## Problem Statement

GraphOS configuration is spread across too many surfaces today: manual Studio setup, Rover workflows, ad hoc scripts, and internal automation. That fragmentation creates inconsistent operating patterns across environments and teams.

A platform team may define conventions for variants, subgraph registration, persisted query usage, or key management, but those conventions are hard to enforce without a declarative model that can be reviewed and reused. Apollo exposes API support for many of these areas already, but it does not yet package them in the standard IaC form that enterprise platform teams want.

That turns into a governance problem as customers scale. Larger GraphOS deployments need the same discipline customers already apply to cloud IAM, Kubernetes objects, network policy, and secrets. Without a first-class Terraform provider, customers either keep GraphOS partially manual or build custom wrappers against the Platform API. Both options increase maintenance cost and reduce consistency.

Brownfield adoption is another practical concern. Most serious GraphOS users already have existing graphs, variants, subgraphs, persisted query lists, and API keys. If the provider does not support import and normalization well, adoption will be painful and customers will have to choose between reconstructing state by hand or living with disruptive migration steps.

## Opportunity

The opportunity is to make GraphOS a first-class part of the enterprise platform stack.

A Terraform provider would let GraphOS configuration participate in the same workflows customers already use for planning, peer review, drift detection, policy enforcement, and multi-environment standardization. That makes GraphOS easier to adopt for platform engineering teams, internal developer platforms, and DevOps organizations that already work through reusable Terraform modules and Git-based change management.

The timing also lines up. HashiCorp now positions the Terraform Plugin Framework as the recommended base for new providers, and the framework has better support for sensitive values, richer type handling, and newer provider patterns than SDKv2. On the Apollo side, the GraphOS Platform API is now the public automation surface for graph details, schema operations, proposals, persisted query lists, API keys, session configuration, and insights. The technical foundation is finally strong enough to build a provider with a clear product boundary.

## Vision

The Apollo Terraform Provider should make GraphOS control-plane configuration declarative, reviewable, importable, and repeatable.

Customers should be able to define the structure and governance of their GraphOS footprint in Terraform, apply changes through CI/CD, and rely on Terraform state as the desired-state model for long-lived GraphOS objects. That includes how variants are created, how subgraphs are registered, how persisted query lists are linked, how keys are issued, and how proposal-related governance is configured across environments.

The provider should feel native to both Terraform practitioners and GraphOS users. Terraform users should get predictable resources, import support, sensible normalization, and durable objects instead of transient tasks. GraphOS users should see familiar concepts rather than a new abstraction layer that hides how Studio, Rover, and the Platform API already work.

## Product Strategy

The Apollo Terraform Provider should be positioned as a GraphOS control-plane provider, not as a replacement for Rover.

Apollo already distinguishes between the two surfaces in practice: Rover is the normal path for schema publishing, while the Platform API is the advanced path for automation and CI/CD. That boundary is useful. Terraform should manage long-lived GraphOS state. Rover and pipeline-native tooling should continue to own fast-moving delivery flows and interactive publishing ergonomics.

That means the first release should stay focused on stable, organization-managed objects and settings:

- Graph variants
- Subgraph registration and routing metadata
- Persisted query lists and their graph or variant associations
- Graph and subgraph API key lifecycle management
- Governance settings such as schema proposal configuration and organization session policy

Highly ephemeral workflows such as repeated schema publication or execution-time checks should come later, if they are added at all, and only in a form that does not fight Terraform's steady-state model.

## Goals

- Let customers manage a meaningful subset of GraphOS configuration entirely through Terraform.
- Support import from day one so existing GraphOS estates can be brought under management without rebuilds.
- Help teams standardize GraphOS configuration across development, staging, and production through reusable modules.
- Strengthen governance by making GraphOS settings easier to review, version, and audit.
- Build on a durable technical foundation using the Terraform Plugin Framework and the GraphOS Platform API.

## Non-Goals

- Replacing Rover as the standard schema delivery interface.
- Exposing every GraphOS capability in the first release.
- Turning the provider into a thin GraphQL pass-through.
- Hiding GraphOS behind a synthetic "environment bundle" resource that obscures the underlying domain model.

## Target Users

The primary users are platform engineers, DevOps engineers, and internal platform teams that already use Terraform to manage cloud infrastructure, Kubernetes, access controls, secrets, and internal platform primitives.

These users are not looking for a developer-centric CLI wrapper. They want GraphOS to behave like the rest of the platform estate: declarative, importable, reviewable, and standardized across environments. Drift detection, module reuse, remote state, and pull-request-based change review matter more to them than interactive publishing ergonomics.

A second user group is GraphOS administrators and graph platform owners who define policy and operational standards even if they are not the people writing Terraform themselves. They care about naming conventions, environment topology, proposal governance, key roles, and how GraphOS changes move safely through the organization.

## User Needs

These users need a reliable way to represent GraphOS control-plane state in code. That starts with graph variants and subgraphs, because those define environment layout and supergraph topology.

They also need clean management of persisted query lists and their links to graph variants. The Platform API already supports CRUD and linkage operations there, which makes PQLs a strong fit for Terraform. Bringing them under IaC lets teams define repeatable client-access patterns instead of relying on manual configuration.

Key management is another major need. Apollo exposes graph API keys, subgraph API keys, and role-sensitive access controls. A provider that can manage key metadata and lifecycle would reduce manual issuance, but it also has to respect the fact that some credentials are only visible at creation time.

Finally, users need import and normalization. Most teams already run production graphs, so the provider has to make brownfield adoption practical. That means import support, stable state representation, and a design that avoids noisy diffs from formatting or API normalization quirks.

## Scope

The first release should focus on control-plane objects and settings that meet four tests:

- They are durable GraphOS nouns rather than transient tasks.
- They map cleanly to CRUD-like Terraform semantics.
- They matter enough to platform teams that IaC management delivers immediate value.
- They are available through the current GraphOS Platform API without relying on unstable internal dependencies.

Within that boundary, the provider should include resources for:

- Graph variants
- Subgraphs
- Persisted query lists
- Persisted query list to graph or variant associations
- Graph and subgraph API keys
- Schema proposal configuration where the API shape is stable enough
- Organization session policy

It should also include data sources for graph metadata, variants, subgraphs, persisted query lists, and selected operational metadata where read-only access is the right abstraction.

## Resource Model

The resource model should follow GraphOS nouns closely.

A variant in Terraform should correspond to a GraphOS variant. A subgraph resource should correspond to a registered subgraph. A persisted query list should correspond to a GraphOS PQL, with a separate resource for linking that PQL to a graph or variant when the relationship needs its own lifecycle.

API key resources should map to graph-scoped and subgraph-scoped keys, exposing roles and metadata where the API allows. This keeps the provider aligned with both the GraphOS domain model and Terraform provider best practices.

The provider should also stay conservative about mixing configuration state with workflow execution. Subgraph registration and routing metadata are durable and make sense as resources. "Publish this SDL every time it changes" behaves more like a delivery operation than a persistent object and should not define the first release.

## Data Source Model

Data sources should complement resources by exposing useful read-only GraphOS information that teams may want to compose into modules or surrounding platform logic.

Logical early data sources include graph metadata, variant information, subgraph metadata, persisted query list metadata, and selected insights. These represent discovered state or analysis, not desired state, which makes them a better fit for data sources than managed resources.

The provider should avoid turning data sources into a Terraform version of Studio dashboards. The goal is to expose high-value metadata and topology, not to mirror every observability surface in GraphOS.

## Detailed Functional Requirements

1. The provider must authenticate to the GraphOS Platform API with a GraphOS API key and support configuration through provider arguments and environment variables.
2. The provider must return clear diagnostics when a credential does not have permission to perform a given operation.
3. The provider must support import for all durable v1 resources. Import IDs should be understandable and aligned with GraphOS identifiers such as graph refs, variant names, and subgraph names.
4. The provider must normalize semantically equivalent values to avoid persistent drift. Custom types, semantic equality, and plan modification should be used where GraphOS returns normalized values that differ from practitioner input.
5. The provider must handle sensitive data carefully. Where possible, it should use the best available Terraform patterns for values that should not be written back to plan or state, and it must document any remaining risk clearly.
6. The provider should expose first-class resources for persisted query list management and PQL linkage, because those are durable objects and relationships already supported by the Platform API.
7. The provider should expose governance-related resources where the API shape is stable enough, prioritizing durable settings such as proposal configuration and session policy ahead of collaborative or workflow-heavy proposal objects.

## Out-of-Scope Functional Requirements

- Replacing Rover-driven schema publishing as the default path for schema delivery.
- Exposing every GraphOS insight, report, or dashboard surface.
- Introducing a monolithic resource that bundles multiple GraphOS primitives into one synthetic abstraction.

## User Experience

The target user experience is straightforward: a platform engineer installs the provider, configures it with a GraphOS API key, and defines GraphOS topology and policy in a module-oriented way.

They should be able to declare a variant, register subgraphs, create a persisted query list, associate it to the right graph or variant, and manage credentials or policy settings without dropping into custom scripts. The configuration should feel like standard Terraform, with predictable CRUD behavior, stable plan output, and import support for brownfield adoption.

The provider also needs to be honest about its boundaries. If a user tries to model a transient workflow that is better handled by Rover or CI/CD automation, the provider and its documentation should say so clearly.

## Technical Architecture

The provider should be implemented with the Terraform Plugin Framework. That gives Apollo the right primitives for GraphOS's mix of durable entities, role-constrained operations, and sensitive values without inheriting the design limits of SDKv2.

Internally, the provider should use a typed GraphOS Platform API client that handles authentication, retries, error normalization, and shared GraphQL query or mutation patterns. The provider layer should then map GraphOS API objects into Terraform schema models with a clear separation between configuration, plan, and state.

Custom types and plan modifiers should be used where they make the provider more predictable. Semantic equality should prevent noisy diffs when GraphOS returns normalized values. Plan modifiers should handle immutable identities, server-generated fields, and invalid in-place transitions cleanly.

## Security and Permissions

Security is central to the provider design because GraphOS exposes multiple scopes of access and because key material is sensitive.

The provider cannot assume that every supported resource can be managed with the same credential type. Some operations may require organization-level authority, while others can be performed with graph-level or subgraph-level credentials. Permission failures need to be detected and explained clearly so users know what class of access is required.

The provider should also minimize exposure of created credentials. Some GraphOS key values are visible only at creation time, so secret handling cannot be treated as an afterthought. Even if the first release cannot avoid every state-handling tradeoff, it should use the safest patterns available and leave room for stronger write-only or ephemeral handling over time.

## API Dependencies and Constraints

The provider depends on the GraphOS Platform API as the authoritative control-plane surface.

That API already exposes enough breadth to support a real provider, including graph details, schema operations, proposals, variant and subgraph removal, API keys, persisted query list CRUD and linking, session-length changes, and selected insights. At the same time, not every available API operation will map cleanly to Terraform.

Before implementation starts, the team should do a capability-by-capability API mapping exercise to confirm which queries and mutations have clean CRUD semantics, which require extra normalization or polling, and which belong outside the provider.

## Success Metrics

The main measure of success is whether the provider becomes a credible, adopted way to manage GraphOS control-plane state in Terraform.

In practice, that means customers can manage a meaningful subset of GraphOS configuration through Terraform, import existing GraphOS estates, and build reusable modules that encode GraphOS standards across environments. Quality of fit matters more than raw resource count.

Operational predictability is the second measure of success. The provider should produce stable plans, minimize spurious drift, and surface clear diagnostics when permissions or API constraints block an operation.

The third measure is product positioning. Apollo teams should feel confident recommending the provider when an enterprise customer asks how to manage GraphOS through infrastructure as code.

## Risks

The biggest risk is scope drift. The Platform API already exposes a lot of functionality, and there will be pressure to include too much in the first release. The provider should stay anchored on durable control-plane state and resist turning v1 into a catch-all automation layer.

Secret handling is the second risk. API keys that are only visible at creation time do not fit neatly into Terraform's normal state model. The provider can reduce that risk through better schema design, strong documentation, and safer handling patterns, but it cannot ignore the tradeoff.

Permission fragmentation is the third risk. Some GraphOS features may need different classes of credentials or authority. If the provider does not communicate those boundaries clearly, users will see confusing failures and lose trust quickly.

## Rollout Plan

The rollout should happen in three phases.

1. API and resource mapping. Inventory Platform API capabilities, classify them into Terraform resources, data sources, future action-style candidates, or out-of-scope domains, and map the required credential type for each.
2. Control-plane MVP. Deliver provider configuration, a stable GraphOS client, import support, and core resources for variants, subgraphs, persisted query lists, PQL links, and API key management, plus a small set of foundational data sources.
3. Governance expansion. Add governance settings such as proposal configuration and organization session policy, then consider carefully chosen workflow-oriented additions only after the control-plane model is stable.
