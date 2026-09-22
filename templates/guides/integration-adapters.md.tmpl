---
page_title: "Integration Adapters"
subcategory: "Cloud Integration"
description: |-
  What a custom Integration Adapter is, how it differs from SAP Business Accelerator Hub
  adapters and capability activation, the Cloud Foundry restriction, and exactly which parts of
  this provider's implementation are confirmed against SAP's own documentation versus inferred
  by analogy.
---

# Integration Adapters

A **custom Integration Adapter** is a `*.esa` archive built with the SAP Adapter SDK and
imported into a Cloud Integration package, so it becomes available for modeling inside
integration flows the same way a prebuilt adapter (SFTP, OData, SOAP, ...) is. This guide
covers what [`sapintegrationsuite_integration_adapter`](../resources/integration_adapter.md),
[`sapintegrationsuite_integration_adapter_deployment`](../resources/integration_adapter_deployment.md),
and [`data.sapintegrationsuite_integration_adapter`](../data-sources/integration_adapter.md) do,
and — more than for most features this provider documents — exactly how much of the underlying
API is independently confirmed versus corroborated by analogy or a secondary source. See
`internal/features/catalog.go`'s `cloud_integration.integration_adapter*` entries for the
canonical, machine-readable version of this same breakdown.

## Three different lifecycles that all say "adapter"

It is easy to conflate these. This provider only manages the first one.

1. **Custom Integration Adapters** (this guide): a tenant administrator's own `*.esa`, built
   with the Adapter SDK, imported into one specific package on one specific tenant. This is what
   `sapintegrationsuite_integration_adapter` manages.
2. **SAP Business Accelerator Hub adapters**: prebuilt adapters SAP (or a partner) publishes on
   the Hub. SAP documents importing one of these as something that happens *inside the
   integration flow editor*, when you assign a sender/receiver adapter that happens to come from
   the Hub rather than being preshipped — SAP's own documentation states this "automatically
   imports the adapter and its package to your design workspace, and automatically deploys the
   adapter to the runtime profile configured for the integration flow." That is a UI-triggered,
   implicit design-time-plus-deploy action with no separate identity a Terraform resource could
   own; this provider does not manage it, and never will unless SAP publishes a genuinely
   separate, deterministic API for it.
3. **Integration Suite capability activation**: enabling Cloud Integration itself, or a
   Discover-section content package, on a subaccount. Unrelated tenant-level provisioning, not a
   design-time artifact at all — see `docs/provisioning-capability-matrix.md`.

## Cloud Foundry only

Every SAP source describing custom Integration Adapters repeats the same note: "This information
is relevant only when you use SAP Cloud Integration in the Cloud Foundry environment." There is
no Neo-environment equivalent. If your tenant runs on Neo, this feature does not apply to you at
all — this provider does not attempt to detect your tenant's environment and will simply surface
whatever error SAP's API returns if you try to use it there.

## What is confirmed, and what is not

This provider's other design-time artifact types (Integration Flow, Value Mapping, Message
Mapping, Script Collection) each have a complete SAP "Example Requests" page showing Create,
Read, Update, Delete, and Deploy. The equivalent page for custom adapters —
"Integration Adapter Example Requests, Cloud Foundry Environment" — shows only two operations:

- `POST /api/v1/DeployIntegrationAdapterDesigntimeArtifact?Id='SubsystemSymbolicName1'` (deploy)
- `DELETE /api/v1/IntegrationAdapterDesigntimeArtifacts(Id='SubsystemSymbolicName1')` (delete)

Both are used exactly as shown. Two structural findings follow directly from that Delete
example, and both differ from every sibling design-time artifact type in this same API:

- The entity is addressed by **`Id` alone**, not the composite `(Id, Version)` key every other
  design-time artifact entity set in this API uses. This provider treats `Id` as the adapter's
  sole key everywhere.
- The deploy action name is **singular** ("...Artifact"), matching every sibling deploy action —
  despite it being tempting to assume "...Artifacts" (plural) by loose analogy with the entity
  set's own plural name, SAP's own example confirms the singular form. It also takes **no
  `Version` query parameter**, unlike every sibling deploy action, which is consistent with `Id`
  being the only confirmed key.

Everything else — the Create request shape, what a Read actually returns, whether metadata can
be edited after import, how a deployed adapter's runtime status is surfaced — is either
corroborated by strong analogy to the sibling artifact types in this exact same API family, by a
secondary technical source, or genuinely unconfirmed:

- **Create**: implemented as `POST IntegrationAdapterDesigntimeArtifacts` with a JSON body of
  `PackageId`, `Id`, `Name`, `Type`, `Application`, and base64-encoded `ArtifactContent` — the
  same shape every sibling design-time artifact type in this API confirms for its own Create,
  and corroborated (not primary-confirmed) by a third-party technical source describing this
  exact request for this exact entity. SAP's own UI documentation confirms `Id` uniqueness is
  tenant-wide and that importing a duplicate `Id` is rejected as an error — useful corroborating
  evidence that Create genuinely is a `POST`-to-create operation, not something else.
- **Read**: `GET IntegrationAdapterDesigntimeArtifacts(Id='...')` — the ordinary OData
  GET-by-key convention every entity set in this API follows, not independently confirmed by an
  adapter-specific example.
- **Update**: not implemented. SAP's documented "duplicate ID is rejected as an error" behavior
  is positive evidence against a working reimport-to-update flow, and no PUT/PATCH/reimport
  example was found anywhere. Every attribute on `sapintegrationsuite_integration_adapter` is
  `RequiresReplace` as a result — content changes, and even purely cosmetic metadata changes
  (SAP's UI separately mentions an "Edit via View metadata" action, but no public API contract
  for it was found), all replace the resource.
- **`type` / `application`**: SAP's UI documentation confirms these exist as classification
  dropdowns (example `type` values: Analytics, CRM, ERP, Finance, HCM, Marketing; example
  `application` value: Slack) but never states whether the underlying property is a closed enum
  or free-form text. This provider does not validate either attribute against a fixed value set
  — adding a validator without that evidence risks rejecting a perfectly valid value SAP itself
  would accept.
- **Deployment runtime status / undeploy**: `sapintegrationsuite_integration_adapter_deployment`
  reuses the same shared `IntegrationRuntimeArtifacts` polling and undeploy every other
  `*_deployment` resource in this provider uses, by analogy. SAP's own documentation states the
  generic runtime-artifacts deploy mechanism "can only deploy BUNDLE type integration artifacts
  (integration flows, value mappings, or OData services)" — explicitly excluding adapters from
  that generic *deploy* path, which is why adapters have their own dedicated deploy action. It
  does not say whether a custom adapter, once deployed through its own action, becomes readable
  or undeployable through that same shared runtime-artifacts entity. This is the single largest
  unconfirmed assumption in this feature; if you observe it behaving differently in your tenant,
  please report it.

## Why the design-time and deployment resources are separate

Consistent with every other design-time artifact type this provider manages:
`sapintegrationsuite_integration_adapter` owns design-time content, and
`sapintegrationsuite_integration_adapter_deployment` owns runtime deployment state. Deploying an
adapter is never triggered implicitly by managing the design-time resource.

## No `adapter_version` on the deployment resource

Every other `*_deployment` resource in this provider (`sapintegrationsuite_integration_flow_deployment`,
`sapintegrationsuite_script_collection_deployment`, and so on) has a `*_version` attribute that
selects which design-time version to deploy, because those entities have a real, confirmed
`(Id, Version)` key. `sapintegrationsuite_integration_adapter_deployment` has no such attribute:
the confirmed deploy action takes only `Id`, so there is nothing else to select — redeploying an
`Id` always (re)deploys whatever content that `Id` currently has.

## Ordering against a consuming integration flow

SAP states plainly that a custom adapter must be deployed before an integration flow that
consumes it is deployed. This provider does not parse integration flow content to discover
which adapters it uses and does not infer this dependency automatically — declare it explicitly:

```hcl
resource "sapintegrationsuite_integration_adapter_deployment" "sftp_extension" {
  adapter_id = sapintegrationsuite_integration_adapter.sftp_extension.id
}

resource "sapintegrationsuite_integration_flow_deployment" "orders" {
  # ...
  depends_on = [
    sapintegrationsuite_integration_adapter_deployment.sftp_extension,
  ]
}
```

## Import

`sapintegrationsuite_integration_adapter` uses a composite `<package_id>/<id>` import syntax,
even though `Id` alone is SAP's confirmed key for Delete and Deploy:

```shell
terraform import sapintegrationsuite_integration_adapter.sftp_extension CUSTOM_ADAPTERS/custom-sftp-extension
```

This is a deliberate Terraform-side choice, not a claim about SAP's OData key. `package_id` is
`Required` and `RequiresReplace` in this resource's schema, and this project could not confirm
that a plain `GET` by `Id` reliably returns the owning package as a queryable property.
Importing by `Id` alone would leave `package_id` unrecoverable — and since it is
`RequiresReplace`, the very first plan after import would want to destroy and recreate the
adapter purely because Terraform never learned which package it belongs to. Requiring the
practitioner to supply a value they already know at import time is a far smaller cost than that.

`sapintegrationsuite_integration_adapter_deployment` imports by the bare adapter ID, matching
the confirmed single-key shape of the deploy/runtime lifecycle:

```shell
terraform import sapintegrationsuite_integration_adapter_deployment.sftp_extension custom-sftp-extension
```

As with every other file-backed design-time resource in this provider, `content`/`content_hash`
cannot be populated by import — SAP does not return a local file path for an existing artifact.
Apply a matching configuration afterward to bring content under management; since every content
change is `RequiresReplace`, this means the first `content`/`content_hash` you declare after
import must exactly match what is already deployed, or Terraform will plan a replacement.

## Restart is not modeled

SAP's UI exposes a *Restart* action for a deployed custom adapter. This provider deliberately
does not implement `sapintegrationsuite_integration_adapter_restart` or any equivalent: restart
is an imperative, operational action, not a piece of declarative desired state, the same reason
this provider has no restart resource for any other deployment type.

## Never automatically executed or inspected

This provider treats `*.esa` content as opaque. It never unpacks, executes, loads classes from,
or otherwise inspects the content of an adapter archive beyond the size bound and content hash
check already applied to every file-backed resource in this provider. `content_hash` protects
against silently uploading unintended content; it is not a security boundary against a malicious
archive, and this provider does not attempt to be one.
