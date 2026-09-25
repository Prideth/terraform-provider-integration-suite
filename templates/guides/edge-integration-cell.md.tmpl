---
page_title: "Edge Integration Cell"
subcategory: "API Management"
description: |-
  What Edge Integration Cell is, where it stops being SAP's public API and starts being your own
  Kubernetes cluster, and exactly what evidence this provider's decision to manage none of it
  today rests on.
---

# Edge Integration Cell

Edge Integration Cell is an optional, hybrid deployment option for SAP Integration Suite's
current API Management model: instead of running entirely on SAP's infrastructure, the runtime
is deployed onto a Kubernetes cluster you operate — on-premises or in your own cloud account —
while SAP's cloud tenant remains the design-time and control-plane counterpart. It shares its
conceptual model (API Artifacts, Runtime Profiles, Virtual Hosts, Policies) with Integration
Cell, described in
`docs/guides/current-api-management.md`, but with its own registration, local API surface, and
operational tooling.

If you came here looking for `sapintegrationsuite_edge_node` or similar: **it does not exist,
and this guide explains exactly why**, together with what this provider does — and, just as
importantly, deliberately does not — manage in this area.

## Two boundaries, not one

Evaluating Edge Integration Cell for Terraform suitability means clearing two separate bars, not
one:

1. **Does SAP publish a public API for it at all?** The same bar every other feature in this
   provider is held to.
2. **If SAP does publish one, is the object it exposes SAP-side configuration, or is it really
   Kubernetes/Helm infrastructure wearing an SAP-shaped API?** Edge Integration Cell is unusual
   among this provider's feature areas in that both kinds of gap show up here: some objects have
   no public API at all, and others have a confirmed, reachable public API that still turns out
   to be the wrong shape for a Terraform resource — either because it fronts your own Kubernetes
   cluster's pod configuration, or because it is monitoring/operational data, not desired state.

See `docs/provider-scope.md` for this provider's general "a public API does not automatically
make its objects Terraform resources" principle; this guide is the detailed application of it to
Edge Integration Cell specifically.

## What this provider manages here today

Nothing. Every object family investigated for this phase resolved to either "no public API" or
"a public API exists, but it belongs to Kubernetes/Helm infrastructure or to monitoring, not to
this provider" — see the sections below and the `edge_integration_cell.*` entries in
`docs/feature-support.md` for the itemized conclusions.

## SAP-side bootstrap: registration and activation

**Activating the Edge Integration Cell capability** is a one-time step under *Settings* >
*Runtime* on the SAP Integration Suite home page, gated by the `edge_integration_cell` service
plan being assigned to the subaccount (SAP Note 2903776) and by the tenant's BTP region
supporting it (SAP Note 3379690). No public activation API was found — the same conclusion this
provider already reached for the Cloud Integration, API Management, and Integration Cell
capability activations it tracks.

**Registering an Edge Node** — the Kubernetes cluster an Edge Integration Cell runtime deploys
onto — is a guided procedure through the separate **Edge Lifecycle Management (ELM)** UI, not
Integration Suite's own interface:

1. A technical user is created in SAP Repositories Management and in Edge Lifecycle Management's
   own P-User flow.
2. A `kubeconfig` file for the target cluster is uploaded through the ELM configuration wizard,
   which downloads a bootstrapping `context.cfg` file in return.
3. The standalone **Edge Lifecycle Management Bridge** executable is run locally, using that
   `context.cfg` and the operator's own `kubeconfig` context, to onboard the cluster and deploy
   the required resources into the `edgelm` namespace.
4. Removing an Edge Node is the same pattern in reverse: *Edge Lifecycle Management* > *Edge
   Nodes* > select > *Remove*, which deletes all solution deployments and the associated
   Kubernetes resources automatically.

None of this is an HTTP API this provider could call. It is a UI wizard plus a standalone CLI
tool performing a direct Kubernetes handshake — the same category of bootstrap process this
provider already treats as out of its automation boundary for Cloud Integration and Integration
Cell capability activation, just with an extra binary in the loop. `edge_integration_cell.registration`
in `docs/feature-support.md` records this conclusion.

## Local API access: real, reachable, and out of scope

Once an Edge Node is running and local API access is enabled (its own *Modify Configuration* step
in Edge Lifecycle Management — an API Virtual Host name and a TLS key alias), SAP exposes a
genuinely public, documented REST API directly against the edge runtime itself, without going
through the cloud UI at all:

```text
https://<your-api-virtual-host>/local/api/v1/<endpoint>       # Message Processing Logs, Message Stores
https://<your-api-virtual-host>/local/api/eic/v1/<endpoint>   # Operations Cockpit
```

Authentication is certificate-based or `clientId`/`clientSecret` (the same two mechanisms this
provider already documents for Cloud Integration), and every modifying call (POST/PUT/PATCH/
DELETE) requires first fetching a CSRF token via a `GET` with `X-CSRF-Token: Fetch` — the same
CSRF discipline this provider's shared HTTP client already implements for every other SAP API it
calls. Confirmed, current endpoints (SAP Business Accelerator Hub packages
`sap-int-eic-message-processing-logs-v1`, `sap-int-eic-message-store-v1`, and
`sap-int-eic-eic-operations`):

| Area | Endpoints | Protocol |
|---|---|---|
| Message Processing Logs | `/MessageProcessingLogs`, `/MessageProcessingLogErrorInformations`, `/MessageProcessingLogAttachments`, `/MessageProcessingLogCustomHeaderProperties`, `/DataStores`, `/DataStoreEntries`, `/Variables` | OData V2 |
| Message Stores | `/MessageStoreEntries` (+ properties/attachments), `/JmsBrokers`, `/MessagingQueues`, `/MessagingQueues(<queue>)/MessagingMessages` | OData V2 |
| Operations Cockpit | `/Components`, `/Components/<id>/Pods`, `/Components/<id>/Pods/<pod>/Events`, `/Components/<id>/Pods/<pod>/Metrics`, `/Components/<id>/RuntimeParameters`, `/Jobs`, `/Jobs/<jobId>/Schedule` | OData V4 |

This is exactly the confirmation the reverification this phase called for — and it changes
nothing about this provider's scope decision. Message Processing Logs, Message Store entries,
JMS resources, Data Store entries, and Variable values are runtime business/operational data, the
identical category this provider already excludes for Cloud Integration's own equivalents (see
`docs/provider-scope.md`'s "A public API does not automatically make its objects Terraform
resources" section). A working, documented `GET` does not by itself make an object
infrastructure — Terraform is not this provider's monitoring system, on the cloud side or the
edge side. `edge_integration_cell.local_api` in `docs/feature-support.md` records this.

## Operations Cockpit: Kubernetes configuration wearing an EIC-shaped API

The Operations Cockpit API's `Component`, `Pod`, `RuntimeParameter`, and `Job`/`JobSchedule`
entities are a more interesting case than plain monitoring data, and worth explaining rather than
dismissing in one line. `RuntimeParameter` is explicitly documented as changeable: a practitioner
can override a component's default value and choose *Save* "to run the changes in the back end,"
which typically triggers "a rolling restart of pods in Edge Integration Cell." That is a
genuinely writable-looking configuration object, not read-only telemetry — which is precisely why
it deserves a real suitability check rather than being grouped in with the local monitoring APIs
above.

It still does not clear this provider's bar, for two independent reasons:

- **It configures Kubernetes, just not through the Kubernetes API.** Every documented
  `RuntimeParameter` — `LOG_LEVEL`, `MIN_REPLICAS`, `MAX_REPLICAS`, `CPU_LIMIT`, `MEMORY_LIMIT`,
  `EPHEMERAL_STORAGE_LIMIT` — sets pod resource requests, replica counts, or log verbosity for
  SAP-operator-managed components (Edge Deploy Controller, Edge Event Controller, Edge Security
  Artifact Controller, and others). `docs/provider-scope.md` already excludes Kubernetes objects
  and Helm releases from this provider's boundary; fronting the same underlying concern with an
  EIC-specific OData API instead of `kubectl` does not change what it configures.
- **Component also exposes an imperative restart action.** SAP's own documentation states you
  "can also restart components with this API." This provider's standing policy — already applied
  to Message Processing Log reprocessing and Number Range counters — refuses to turn `restart`,
  `retry`, `cancel`, or `purge` into a Terraform resource, since Terraform's plan/apply model
  represents desired state, not one-shot imperative operations.

There is a third, practical reason this provider would not implement it even if the boundary
question resolved differently: the exact entity keys and PATCH/POST payload shapes for
`RuntimeParameter` and `JobSchedule` are not confirmed from any reachable primary source. The
package's `$metadata`/EDMX document and worked request/response examples live behind SAP Business
Accelerator Hub's authenticated catalog pages (`api.sap.com/package/sap-int-eic-eic-operations`),
which redirect an unauthenticated request straight to a login page — the same access limitation
this project has hit and documented for several other `api.sap.com` packages. `edge_integration_cell.runtime`
in `docs/feature-support.md` records this three-part reasoning.

## Deployment targeting: no confirmed API parameter

SAP's Operations UI shows a "Runtimes" multi-select field on some objects (documented explicitly
for Number Ranges: "one or more runtime nodes to deploy the artifact to... including Cloud
Integration and any active Edge Integration Cell nodes"), suggesting a design-time artifact could
in principle be steered toward a specific Edge node during deployment. No corresponding parameter
appears in any confirmed API this provider calls or has researched: not in the Number Ranges
`Add`/`Update` examples, and not in any of the `Deploy` actions this provider already implements
(`IntegrationDesigntimeArtifacts`, `MessageMappingDesigntimeArtifacts`,
`ScriptCollectionDesigntimeArtifacts`, `ValueMappingDesigntimeArtifacts`). This provider always
targets the implicit default runtime and does not guess at an undocumented query or body
parameter. `edge_integration_cell.deployment_target` in `docs/feature-support.md` records this.

## Access Policy replication: the API exists, its contract does not

In the Access Policies screen an administrator picks the runtimes a policy is created in,
which can include individual Edge Integration Cells, and can change that selection later. Each
runtime then reports a reconciliation status. An Edge Integration Cell that is offline shows
*Pending* until it reconnects and picks the policy up.

On the API side, SAP's own CI/CD tooling shows that the public `AccessPolicies` entity carries
these assignments in a navigation property named `AccessPolicyRuntimeAssignments`. That is as
far as public evidence goes. Neither SAP Help nor any published sample shows what an
assignment contains, how an Edge Integration Cell is identified in it, or whether assignments
can be created through the API. `sapintegrationsuite_access_policy` therefore manages the
policy and leaves runtime selection to the UI. An earlier release exposed a
`reconciliation_status` attribute on the policy; it was removed because the policy entity has
no such property. The [Access Policies guide](access-policies.md) describes what this means in
practice, and `edge_integration_cell.access_policy_replication` in `docs/feature-support.md`
tracks the gap.

## What remains manual

Every one of the following stays a manual, UI- or CLI-driven step, exactly as it does for Cloud
Integration and Integration Cell capability activation elsewhere in this provider:

- Activating the Edge Integration Cell capability
- Registering, upgrading, and removing an Edge Node through Edge Lifecycle Management and the ELM
  Bridge executable
- Deploying the Edge Integration Cell solution onto the Kubernetes cluster (Helm-based; see SAP's
  own `deploy-the-edge-integration-cell-solution` documentation)
- Configuring local API access (API Virtual Host, TLS key alias)
- Adjusting Operations Cockpit runtime parameters, restarting components, or managing scheduled
  jobs
- Creating additional Istio virtual hosts, which SAP documents as raw `kubectl apply` of
  `Gateway`/`VirtualService` Istio custom resources — this provider does not manage Kubernetes
  custom resources under any circumstances, on the edge cluster or anywhere else

## How future API Artifact support could target Edge

`docs/guides/current-api-management.md` already documents the one Edge-specific exception SAP
carves out in its Runtime Profile model: while a Runtime
Profile is otherwise immutable once an API artifact is created, "the only exception applies to
API artifacts created with the Edge Integration Cell runtime profile, where you can select or
change the target Edge Integration Cell node during editing or deployment." If SAP ever publishes
a public API Artifact API, this provider's design intent is to extend `runtime_profile` with that
one documented Edge-specific mutability exception rather than treating Edge Integration Cell as a
wholly separate resource family — the two guides describe one product model split across two
documents purely for research-history and readability reasons, not because this provider expects
them to diverge architecturally.

## Revisiting this decision

Re-run the same two-part check this guide describes — public API existence, then Kubernetes/
monitoring suitability — before implementing anything here, especially for the Operations Cockpit
API, which is the one object family in this list that could plausibly cross into Terraform
suitability if SAP ever exposes its `$metadata` outside an authenticated Business Accelerator Hub
session and this provider can confirm exact field-level behavior.
