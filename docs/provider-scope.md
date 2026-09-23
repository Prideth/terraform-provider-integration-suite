# Provider Scope

This document defines what `Prideth/sap-integration-suite` manages and what it deliberately
does not manage. It exists so that users, reviewers, and contributors can tell at a glance
whether a feature request belongs in this provider or in a different one.

## In Scope

The provider administers the parts of an **already provisioned** SAP Integration Suite tenant
that are reachable through officially documented, SAP-supported public APIs:

- Integration Suite capability discovery and (where a public API exists) capability
  configuration
- Cloud Integration design-time content: integration packages, integration flows, value
  mappings, script collections, message mappings
- Cloud Integration runtime: deployments of the artifacts above, deployment status
- Access Policies and their artifact references
- Partner Directory entries, where a public API exists
- Classic API Management objects (API Providers, API Proxies, API Products, key value maps),
  where a public API exists
- The newer API-artifact-centric API Gateway model (API Artifacts, API Resources, API
  Policies, runtime profiles, deployments), where a public API exists
- Integration Cell: capability status/read, and any configuration that has a public API
- Edge Integration Cell: SAP-side registration, runtime association, and control-plane
  configuration that has a public API (never the Kubernetes workloads themselves)
- Integration-Suite-specific runtime configuration that is not already owned by the SAP BTP
  control plane

Every resource in this list is only implemented once a concrete, currently documented public
API has been identified for it (see [`sap-api-references.md`](sap-api-references.md) and the
capability/API matrices in this directory). A feature being visible in the Integration Suite
UI is not sufficient justification for a resource.

## A public API does not automatically make its objects Terraform resources

This is worth stating as its own principle, not just an implication of the rule above: SAP
publishing a documented, callable API for an object — even one with a full `GET` — is
**necessary but not sufficient** for that object to become a `sapintegrationsuite_*` resource.
Terraform resources model **desired infrastructure state that a practitioner owns and
reconciles**; several categories of public API object fail that test even though the API itself
is real and reachable:

- **Runtime-generated, not practitioner-authored.** An object that only ever comes into
  existence as a side effect of deployed content running (a Data Store, created the first time
  an integration flow's Data Store Write step executes; a Variable, written by a Write Variables
  step) has no Terraform-owned "desired" shape to converge toward. There is nothing to `apply`.
- **Arbitrary runtime business data, not configuration.** A Variable's value, or a Data Store
  Entry's message payload and processing status, is business data flowing through deployed
  integration content — it changes on every message the content processes, for reasons entirely
  outside Terraform's control. Putting it in Terraform state would mean either constant spurious
  diffs or, worse, silently treating live business data as if it were configuration this provider
  reconciles. Message Processing Logs are the original example of this in this provider; Data
  Store Entries and Variable values are the same category.
- **A GET without a safe Create/Update/Delete lifecycle.** `DataStores?overdueonly=true` is a
  real, documented, working GET — and it returns monitoring aggregates (message counts), not
  configuration. A resource needs more than "SAP will tell you something if you ask"; it needs
  something a practitioner can safely bring into existence, change, and tear down again.
- **A Create/Update without a safe Read.** The inverse case: `NumberRanges` has confirmed,
  documented `POST`/`PUT` operations, but no documented `GET` anywhere and no documented
  `DELETE`. This provider still implements `sapintegrationsuite_number_range`, but as an
  explicitly narrower "write-only lifecycle" resource — Read is a documented no-op, and Import
  and Delete both refuse with an explicit error rather than guess at unconfirmed behavior. See
  `docs/guides/runtime-stores-and-number-ranges.md` for the full reasoning. This is the
  conservative middle ground between "skip it entirely" and "pretend the missing operations
  exist" — chosen here specifically because Create/Update alone still give a practitioner a real,
  auditable way to push desired static configuration, unlike the runtime-data cases above.

Concretely, in the Number Ranges / Variables / Data Stores API family: a Number Range's static
configuration (name, min/max, rotate, field length) is potentially suitable for Terraform
management, and its live runtime counter is not — it is exposed as a separate, write-only,
explicitly-gated attribute rather than something Terraform's plan/apply cycle reconciles by
default. A Variable's runtime value is not suitable at all — no resource, no data source. A Data
Store Entry's payload and processing status is operational runtime data, not infrastructure —
unsuitable for the same reason Message Processing Logs are unsuitable. See
`docs/resource-design.md` and `docs/api-capability-matrix.md` for the full per-object
suitability analysis this provider went through before reaching these conclusions.

## Developer Hub is a separate, future provider

Developer Hub — Integration Suite's API/Event/MCP Server catalog, publication, and subscription
capability — is intentionally outside this provider's scope. It authenticates against its own
`/api/1.0` REST API with its own `devportal-apiaccess` OAuth client, entirely separate from every
Cloud Integration and current API Management endpoint this provider talks to, and its object model
(Products, Applications, Subscriptions) is a consumer/catalog lifecycle rather than Integration
Suite design-time content. Rather than stretching this provider's credential and release surface
to cover a genuinely separate API boundary, Developer Hub is planned as its own, independently
versioned Terraform provider (working name `Prideth/terraform-provider-sap-developer-hub`). This
provider carries no Developer Hub configuration, client, resources, or data sources; see the
`developer_hub` entry in `docs/feature-support.md` for the single, high-level catalog statement of
this boundary.

## Out of Scope

The following belong to the SAP BTP control plane, to Kubernetes/Helm, or to other existing
Terraform providers, and are intentionally **not** implemented here:

- BTP Global Accounts, Directories, Subaccounts
- BTP Entitlements and generic BTP Subscriptions (including subscribing to Integration Suite
  itself)
- Generic BTP Service Instances and Service Bindings
- BTP Role Collections and Role Assignments
- Generic BTP Destinations
- Cloud Foundry Organizations and Spaces
- Kyma, Kubernetes objects, Helm releases
- General BTP account/identity management
- Developer Hub (planned as a separate Terraform provider — see above)

For all BTP control-plane concerns, use the official
[`SAP/btp`](https://registry.terraform.io/providers/SAP/btp/latest) Terraform provider. For
the customer-managed Kubernetes cluster that an Edge Integration Cell runs on, use the
Kubernetes and Helm providers directly.

## One-sentence summary

`Prideth/sap-integration-suite` is a Terraform provider for configuring, provisioning, and
administering **content and capabilities inside an already-provisioned SAP Integration
Suite tenant** — it is not, and will not become, a second SAP BTP provider.
