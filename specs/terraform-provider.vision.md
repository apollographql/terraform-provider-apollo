# Apollo Terraform Provider

## Summary

Apollo should offer a Terraform provider for GraphOS so platform teams can manage GraphOS configuration the same way they manage cloud infrastructure, Kubernetes, and other shared platform systems.

Today, GraphOS setup is split across Studio, Rover workflows, and custom automation. A provider would give teams a declarative model for the parts of GraphOS that belong in infrastructure-as-code: defining resources once, reviewing changes in pull requests, and applying them through existing CI/CD pipelines.

The provider should be built with the Terraform Plugin Framework, and it should use the GraphOS Platform API as the system of record. That gives Apollo a provider that matches current Terraform best practices and is built on the right control-plane API.

## Problem

GraphOS configuration lives in too many places today. Teams manage parts of it in Studio, parts of it through Rover or other CLI workflows, and parts of it through custom scripts. That works, but it does not scale cleanly.

The first issue is consistency. Variants, subgraphs, persisted query lists, and governance settings can all be created or updated in different ways across graphs and environments. As usage grows, those differences become harder to track and harder to unwind.

The second issue is workflow fit. Many enterprise platform teams already use Terraform to manage infrastructure, Kubernetes objects, IAM policies, and secrets. They want GraphOS changes to move through the same review, approval, and rollout path, not through a separate set of manual steps.

The third issue is maintainability. Teams can script against Apollo APIs today, but one-off automation is harder to audit, reuse, and standardize than first-class Terraform resources and modules.

Governance is the broader problem underneath all of this. As organizations add more graphs, more environments, and more teams, they need a reliable way to define defaults, security controls, topology, and proposal policies without depending on repeated Studio operations.

## Opportunity

A Terraform provider gives Apollo a much stronger entry point into the platform engineering ecosystem.

It lets Apollo meet customers inside workflows they already trust: Terraform, GitOps, and centralized infrastructure governance. For large organizations, the value is not just automation. It is consistency, auditability, drift detection, and reusable modules that capture how GraphOS should be set up.

It also helps position GraphOS more clearly as enterprise control-plane infrastructure, not just a developer-facing surface.

## Vision

The Apollo Terraform Provider should make GraphOS feel like part of the infrastructure stack.

Platform teams should be able to define GraphOS topology and governance in code, review changes in pull requests, apply them through CI/CD, and import existing GraphOS configuration into Terraform state. The provider should stay focused on durable control-plane objects rather than release-time workflows, because that is where Terraform fits naturally.

Over time, teams should be able to use shared Terraform modules to standardize GraphOS setup across development, staging, and production in the same way they already standardize cloud accounts, Kubernetes clusters, networking, and identity systems.

## Goals

The provider should give customers a reliable way to manage GraphOS configuration as code. The first release should concentrate on stable, high-value GraphOS objects that map cleanly to Terraform resources and support brownfield adoption through import.

It should let teams define GraphOS resources once and reuse those patterns across environments. It should also improve governance by exposing GraphOS settings and relationships in a form that can be reviewed, versioned, and audited.

The technical foundation matters too. Building on the Terraform Plugin Framework and the GraphOS Platform API gives Apollo a provider that follows current best practices on both sides of the integration.

## Non-Goals

The initial provider should not try to replace Rover for every schema delivery workflow. Publishing schemas on every build is usually a better fit for CI-oriented release automation than for Terraform's steady-state resource model.

It also should not attempt to mirror every GraphOS feature in the first release. A focused provider built around the highest-value control-plane objects will be more useful, and much easier to maintain, than a broad but shallow one.

The first version should also stay away from workflow abstractions that depend on human coordination or ephemeral execution semantics. Long-lived configuration and governance are the right starting point.

## Proposed Scope

The initial scope should center on GraphOS capabilities that are durable, organization-managed, and clearly valuable to platform teams.

That likely includes graph variants, subgraph registration and routing metadata, persisted query lists and their links to variants, API key lifecycle management, and governance settings such as schema proposal configuration and organization session policy.

The provider should also include data sources for GraphOS metadata where read-only access is the right abstraction. That gives users a way to pull GraphOS context into larger Terraform workflows without trying to force ephemeral operations into resource state.

## Example Capability Areas

The first release should likely include:

- Graph variant management
- Subgraph registration and routing configuration
- Persisted query list management
- Persisted query list to variant linking
- Graph and subgraph API key management
- Schema proposal configuration
- Organization-level session policy
- Data sources for graphs, variants, subgraphs, and selected GraphOS metadata

These are the areas that fit Terraform best because they map to long-lived objects and policy settings rather than transient execution flows.

## User Value

For platform engineers, the provider creates a standard, reviewable way to manage GraphOS at scale. For DevOps and internal platform teams, it reduces manual setup and makes GraphOS part of the infrastructure workflows they already operate.

For security and governance stakeholders, it improves visibility into who owns GraphOS configuration and how it changes over time. For Apollo, it reduces the need for customers to build their own API wrappers and internal tooling while making GraphOS easier to adopt in enterprise environments.

## Key Product Principles

The provider should stay close to GraphOS concepts. Users moving between Studio, Rover, documentation, and Terraform should not have to learn a second mental model. A variant should still look like a variant, a subgraph should still look like a subgraph, and persisted query list configuration should map cleanly to the underlying GraphOS concepts.

Brownfield adoption needs to be there from the start. Most customers will not begin with an empty GraphOS footprint; they will already have graphs, variants, and subgraphs in production. Import and normalization are essential, not optional.

The provider also needs to handle secrets and other sensitive values carefully. API keys and one-time-visible credentials need deliberate treatment in both schema design and documentation.

## Why Now

Apollo is increasingly being used as part of a broader platform layer inside enterprise environments. As customers mature, they expect GraphOS to plug into the same automation systems they use for the rest of their infrastructure.

The timing also makes sense technically. The GraphOS Platform API is a stronger foundation for external automation than Apollo has had before, and the Terraform ecosystem has moved to the Plugin Framework as the recommended base for new providers. If Apollo is going to offer a provider, this is the right time to do it with the modern stack rather than extend custom scripts or older implementation patterns.

## Success Criteria

A successful first release would let platform teams manage a meaningful subset of GraphOS configuration entirely through Terraform, adopt the provider against existing GraphOS estates, and use modules to standardize GraphOS setup across environments.

It should also feel native to both Terraform users and GraphOS users. If the provider is scoped well and implemented cleanly, it should become the default recommendation for customers who want infrastructure-as-code management of GraphOS control-plane resources.


## Positioning

The Apollo Terraform Provider should be positioned as a GraphOS control-plane provider for infrastructure-as-code, not as a replacement for Rover.

Its job is to help customers standardize, govern, and automate GraphOS configuration with Terraform using modern provider patterns and the GraphOS Platform API. If Apollo keeps the scope tight, this can become an important bridge between GraphOS and the platform engineering workflows that shape enterprise adoption.
