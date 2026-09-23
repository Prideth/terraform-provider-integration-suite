---
page_title: "Current API Management and Integration Cell"
subcategory: "API Management"
description: |-
  SAP's current, API-centric integration model — API Artifacts, Runtime Profiles, Integration
  Cell, Virtual Hosts, Policies, Reusable API Artifacts — why this provider currently manages
  none of it, and exactly what evidence that conclusion rests on.
---

# Current API Management and Integration Cell

SAP Integration Suite now has **two** distinct API management product models, and this provider
treats them as two distinct domains, never mixed:

- **Current API Management** — API Artifacts, Runtime Profiles, Integration Cell, Virtual
  Hosts, Policies, Reusable API Artifacts. This is SAP's current product direction, described in
  this guide.
- **Classic API Management** — API Providers, API Proxies, API Products, classic Key Value
  Maps, the API Portal. A separate, later phase for this provider — see `ROADMAP.md`.

If you came here looking for `sapintegrationsuite_api_artifact` or similar: **it does not exist
yet, and this guide explains exactly why**, with the evidence this provider's decision rests on.

## The short version

This provider found **no public API** — REST, OData, or otherwise — for any of the following,
after a thorough documentation research pass:

- API Artifacts (create, read, update, delete, version)
- API Artifact deployment/undeployment
- API Policies
- Reusable API Artifacts
- Runtime Profiles
- Integration Cell activation or runtime status
- Integration Cell Virtual Hosts

Every one of these is real, working SAP functionality — accessible entirely through the SAP
Integration Suite web UI (*Design* > *Integrations and APIs*, *Monitor* > *Integrations and
APIs*, *Settings* > *Runtime*/*Integrations*). This provider does not invent an API SAP has not
documented, so none of it is implemented as a Terraform resource or data source. See
`docs/resource-design.md` and `docs/sap-api-references.md` for the full suitability check and
evidence trail this conclusion is based on.

## How this conclusion was reached

This was not assumed — it was actively ruled in or out, in order:

1. **Is this the existing Integration Content API under a new artifact type?** No. The
   canonical Integration Content OData API overview page's "Resources" table — the same
   authoritative list this provider already relies on for `IntegrationDesigntimeArtifacts`,
   `MessageMappingDesigntimeArtifacts`, `ScriptCollectionDesigntimeArtifacts`,
   `ValueMappingDesigntimeArtifacts`, `IntegrationAdapterDesigntimeArtifacts`, and
   `IntegrationRuntimeArtifacts` — does not include API Artifacts at all. This is the single
   strongest piece of evidence in this research pass: it is an exhaustive, authoritative list,
   and API Artifacts are simply absent from it.
2. **Is there a separate, newer REST/OData API?** Over twenty documentation pages were read in
   full, covering every stage of the API Artifact lifecycle — creation (by target URL/OpenAPI
   specification, by BTP destination, by discoverable system, by referring to an API provider),
   configuration, versioning (including reverting to a previous version), access management,
   deletion, deployment, monitoring, Reusable API Artifacts, Virtual Host management, and
   Runtime Profile configuration. Every single one describes its subject exclusively through UI
   procedures ("Log on to SAP Integration Suite... choose the navigation icon... select..."). Not
   one mentions an OData/REST endpoint, an entity set name, an HTTP method, an example request,
   or a link to SAP Business Accelerator Hub — the pattern every other confirmed public API in
   this provider follows consistently and without exception.
3. **Is UI support being mistaken for API support?** The one page in the entire documentation
   area whose title suggested otherwise — "Accessing API Management APIs Programmatically" —
   turned out, on full reading, to describe exclusively the **Classic** API Management API
   Portal (`apiportal-apiaccess` plan, `Management.svc/APIProxies`), confirmed by the identical
   file also being mirrored under the separate Classic API Management documentation tree. It is
   evidence for Classic API Management's public API (already known, to be handled in that
   later phase), not for the current model.

That leaves Outcome C: real, UI-documented SAP functionality with no public API behind it today.
This provider does not implement Terraform resources against an API it cannot point to
documentation for.

## The conceptual model, for when a public API does appear

Understanding the product model is still useful, both for practitioners working through the UI
today and for evaluating this provider's coverage the next time SAP's public API surface is
reverified.

```text
BTP Subaccount
│
├── API Management capability      [bootstrap — UI only, no public API]
│
├── Integration Cell               [bootstrap/runtime — UI only, no public API]
│   └── Virtual Hosts              [UI only, no public API]
│
└── Integration Package
    └── API Artifact               [UI only, no public API]
        ├── API definition (REST / SOAP / OData / Reusable)
        ├── policies / mediation
        ├── runtime profile (Integration Cell / Edge Integration Cell)
        └── deployment
             └── Integration Cell + Virtual Host
```

Neither the BTP subscription nor the API Management/Integration Cell capability activations are
provider-managed resources — they are bootstrap prerequisites a tenant administrator completes
once, through the UI, before any of the content below them can exist. This provider starts from
the assumption that a tenant already has an active Integration Suite subscription with the API
Management capability and, if applicable, an active Integration Cell — the same starting
assumption this provider makes for Cloud Integration, Security Content, and every other feature
family it manages.

### API Artifact

The core design-time object: a complete API configuration (endpoints, policies, security
configuration, runtime behavior) created under *Design* > *Integrations and APIs* > an
integration package > *Add* > *API*. Supported service types include REST, SOAP, OData, and
Reusable — each determining a different default sender/receiver adapter pair and initial policy
model. Creation methods include a direct target URL, an OpenAPI specification, a BTP
destination, a discoverable system (for example SAP S/4HANA Cloud), or by referring to an
existing API provider.

**API State** (Alpha/Beta/Active) is a classification/demarcation field, not a lifecycle flag —
SAP's own documentation is explicit that it does not govern deployment. If a public API is ever
confirmed, this provider would not map `Active` to a `deployed` boolean; the two concepts stay
separate, the same distinction this provider already draws between an artifact's design-time
state and its runtime deployment status elsewhere in the codebase.

**Versioning**: multiple versions of an API artifact can coexist at both design time and
runtime — unlike an Integration Flow, where a new version typically replaces what is deployed.
SAP documents an explicit revert operation, and states that "reverting to a previous version
does not delete other versions from the version history." Whether SAP's "API Version" and the
design-time artifact revision are the same field or two distinct concepts could not be
determined without a documented API to inspect — flagged here as an open question for whenever
one exists, not answered by guessing.

### Runtime Profile

A runtime profile determines which target platform an API artifact (or integration flow) is
designed and deployed for. SAP's own "Runtime Profiles" documentation page lists Cloud
Integration, Cloud Integration – Starter, SAP Process Orchestration (per release), and Edge
Integration Cell — but, inconsistently with the rest of the API Artifact documentation, does
**not** list a separate "Integration Cell" row, even though Integration Cell is offered as a
distinct runtime profile choice when creating an API artifact elsewhere in the same
documentation set. This provider does not attempt to resolve that inconsistency; it is recorded
here as an open documentation gap, not a fact this provider guessed at.

**Runtime profile is effectively immutable once an API artifact is created**, confirmed verbatim
from SAP's own documentation: "After an API artifact is created on a specific runtime, you
cannot change its runtime profile either by editing the artifact or during deployment. For
example, an API artifact created for the Integration Cell runtime cannot later be deployed to
the Edge Integration Cell runtime, and vice versa." If a public API is ever confirmed and a
`runtime_profile` attribute becomes possible, this provider would model it as `RequiresReplace`
for exactly this reason — never as a field a normal `apply` could silently redeploy across
runtimes.

**The one documented exception is Edge Integration Cell-specific**: "The only exception applies
to API artifacts created with the Edge Integration Cell runtime profile, where you can select or
change the target Edge Integration Cell node during editing or deployment." This provider does
not generalize that exception to Integration Cell; it stays scoped to Edge Integration Cell,
which remains a separate, later phase of its own.

### Integration Cell and Virtual Hosts

Integration Cell is SAP's current, Kubernetes-based, fully managed runtime for API and MCP
Server lifecycle management — a distinct runtime surface from Cloud Integration's own runtime,
with its own section of the Monitor application (*Monitor* > *Integrations and APIs*, with its
own *Runtime* selector for Integration Cell vs. Edge Integration Cell). Activation is a one-time
`Settings` > `Runtime` step; management of already-deployed content happens through `Monitor`.

A **Virtual Host** defines the public-facing host name and base path through which API and MCP
Server artifacts are exposed — managed under *Monitor* > *Manage Virtual Host* by an
administrator holding the `PI_Administrator` role collection. SAP documents a distinction this
provider considers important enough to design around, whenever a public API does exist:

- **Design-time virtual host** — the virtual host configured on the API artifact itself, in the
  Design workspace.
- **Deployment-time virtual host** — the virtual host actually selected when the artifact is
  deployed, which can differ from the design-time value. SAP's own worked example: an artifact
  configured with `api-dev.company.com` at design time, deployed against
  `api-prod.company.com`, keeps showing `api-dev.company.com` in its design-time configuration,
  while the endpoint URL shown under *Monitor* > *Manage Integration Content* reflects
  `api-prod.company.com` — the value actually in effect at runtime.

Were a public API to appear, this provider would model these as two separate fields on two
separate resources (a design-time field on `sapintegrationsuite_api_artifact`, a
deployment-time field on `sapintegrationsuite_api_artifact_deployment`), matching the
design-time/runtime-deployment split this provider already uses consistently for every other
Cloud Integration artifact type. The default virtual host, specifically, would need read-only
treatment rather than an ordinary mutable resource: SAP documents it as having restricted
editability compared to an administrator-created additional virtual host, similar in spirit to
how this provider already treats SAP-owned keystore entries as something to observe, not
destroy.

### Reusable API Artifacts

A modular, internal-only component (policies, mediation logic, schemas, reusable behavior) with
no externally callable HTTP endpoint of its own — invoked only by another API artifact through
the API Direct adapter, and explicitly unable to call other Reusable API Artifacts recursively.
SAP's own documentation treats this as a variant of the general API Artifact concept rather than
a materially different object, and this provider's design intent, if a public API ever appears,
would follow that lead (an `artifact_kind` distinction on the same resource, not a separate one)
unless the confirmed API contract turns out to differ enough to justify otherwise.

### Policies

API Artifacts support policies and mediation steps — authentication, authorization, quota,
spike/surge protection, rate limiting, JSON threat protection, transformations, content
modification. Whether these are persisted as opaque content nested inside the artifact or as
independently addressable entities could not be determined without a public API to inspect
either way. This provider does not create a `sapintegrationsuite_api_policy` resource on the
strength of UI visibility alone — an independent resource needs independently confirmed Create,
Read, Update, Delete, and a stable identity, the same bar this provider applies everywhere else.

### BTP Destinations stay out of this provider

API Artifacts can reference SAP BTP Destinations as a backend connection method. This provider
does not, and will not, create or manage BTP Destinations — that belongs to the official SAP BTP
Terraform provider (`btp_subaccount_destination` and similar). If a public API Artifact API
appears and supports referencing a destination by name, this provider would consume that
reference, never reimplement destination management itself:

```hcl
resource "btp_subaccount_destination" "backend" {
  # managed by the SAP BTP provider, not this one
}

resource "sapintegrationsuite_api_artifact" "orders" {
  # would reference the destination by name, if and when a public API supports it
}
```

SAP separately documents that a destination intended for Integration Cell discovery/use needs
the custom property `IntegrationCell.Include = true` — a prerequisite worth knowing about, not
something this provider mutates on your BTP destinations regardless of which provider manages
them.

### MCP Servers share this infrastructure, and are out of scope for the same reason

SAP's current documentation also describes MCP Server artifacts, built on the same Integration
Cell infrastructure (Runtime Profiles, Virtual Hosts, Policies) as API Artifacts. This provider
does not implement MCP Server support in this phase either, for the identical reason: no public
API was found for it during this research pass.

## What Terraform can manage today

Nothing in this product area, as of this research pass. This provider's role begins once (a) an
Integration Suite subscription exists, (b) the API Management capability is active, and (c), if
targeting Integration Cell, that runtime is active and has at least one virtual host configured
— all three remain manual, UI-driven bootstrap steps this provider does not automate, consistent
with how this provider already treats Cloud Integration and Integration Cell capability
activation elsewhere. Everything below that bootstrap boundary (API Artifacts, their deployment,
policies, Virtual Hosts) also stays outside this provider until SAP publishes a documented public
API for it.

## Revisiting this decision

This is not a permanent judgment about SAP's roadmap — it is the accurate state of SAP's *public,
documented* API surface as of this research pass. If SAP publishes a public API for any of these
objects, re-run the same research this guide describes (checking the Integration Content
resource table first, then a full documentation sweep for a dedicated API reference) before
implementing anything, exactly as this provider did here.
