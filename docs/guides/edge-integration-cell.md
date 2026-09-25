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

Two things, both through the cloud tenant's own APIs:

- **Content and security material on a specific Edge Integration Cell**, selected with
  `runtime_location_id` on the deployment and credential resources. This is experimental; see
  [Targeting an Edge Integration Cell](#targeting-an-edge-integration-cell).
- **Which runtimes an access policy has reached**, read with
  `sapintegrationsuite_access_policy_runtime_assignments`.

Everything else about Edge Integration Cell resolves to either "no public API" or "a public API
exists, but it belongs to Kubernetes/Helm infrastructure or to monitoring, not to this provider".
The sections below and the `edge_integration_cell.*` entries in `docs/feature-support.md` give
the itemized conclusions.

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

## Targeting an Edge Integration Cell

Earlier research looked for a *parameter* on the deploy actions that selects an Edge Integration
Cell, and found none. The answer turned out to be the URL instead. Since mid-2026, SAP Help's
Integration Content, Security Content and Partner Directory pages document a second service root
for the same APIs:

```text
https://<tenant host>/api/v1/<path>                                 cloud runtime
https://<tenant host>/location/<runtime location id>/api/v1/<path>  an Edge Integration Cell
```

Every operation is the same; only the prefix changes. The provider exposes this as an optional
`runtime_location_id` on the resources whose objects exist once per runtime:

```terraform
resource "sapintegrationsuite_user_credential" "erp_on_edge" {
  id                  = "ERP_BASIC"
  runtime_location_id = "plant-a"
  user                = "integration-user"

  password_wo         = var.erp_password
  password_wo_version = "1"
}

resource "sapintegrationsuite_integration_flow_deployment" "orders_on_edge" {
  package_id          = sapintegrationsuite_integration_package.orders.id
  flow_id             = sapintegrationsuite_integration_flow.orders.flow_id
  flow_version        = sapintegrationsuite_integration_flow.orders.version
  runtime_location_id = "plant-a"
}
```

The design-time content itself (`sapintegrationsuite_integration_flow` and the other content
resources) lives in the cloud tenant and has no runtime location. Only where it runs, and the
security material it uses there, is per runtime. To run the same flow in the cloud and on an
Edge Integration Cell, declare two deployment resources, one with and one without
`runtime_location_id`.

**Finding the ID.** In Integration Suite, open *Monitor* > *Integrations and APIs* and choose
the Edge Integration Cell in the *Runtime* selector. The browser URL then contains
`{"edge":{"runtimeLocationId":"plant-a"}}`. The provider accepts letters, digits, `.`, `_` and
`-`, which keeps the ID a single safe URL segment.

**Changing it** replaces the resource: the object is created on the new runtime and removed
from the old one. For deployments that means an undeploy on the old runtime.

**Importing** takes the location as an optional first segment, for example
`plant-a/ERP_BASIC` for a credential or `plant-a/ORDERS/order_flow` for a deployment. Without
the prefix, the cloud runtime is assumed.

**Why experimental.** SAP documents the prefix once for all operations, with no per-operation
examples, and SAP's own CI/CD tooling still described the path as unpublished in May 2026. It
has not yet been verified against a tenant with an Edge Integration Cell. Certificates, key
pairs, keystore reads and Partner Directory resources do not take `runtime_location_id` yet.
`edge_integration_cell.deployment_target` in `docs/feature-support.md` tracks the status.

## Access Policy replication: readable, not writable

In the Access Policies screen an administrator picks the runtimes a policy is created in,
which can include individual Edge Integration Cells, and can change that selection later. Each
runtime then reports a replication status. An Edge Integration Cell that is offline shows
*Pending* until it reconnects and picks the policy up.

The public `AccessPolicies` entity exposes these assignments through its
`AccessPolicyRuntimeAssignments` navigation property. Each assignment names its runtime by
`RuntimeLocationId` and carries `TransferStatus`, `TransferErrors` and `StatusUpdatedAt`. The
`sapintegrationsuite_access_policy_runtime_assignments` data source reads them, which makes an
Edge Integration Cell that has not received a policy visible from Terraform or from monitoring
built on it. Whether assignments can be written through the API is not documented, so
runtime selection itself stays in the UI. The [Access Policies guide](access-policies.md)
describes the practical consequences, and `edge_integration_cell.access_policy_replication` in
`docs/feature-support.md` tracks the write gap.

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
