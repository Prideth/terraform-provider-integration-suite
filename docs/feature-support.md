# Feature Support

Generated from `internal/features/catalog.go` by `go run ./cmd/gendocs`. Do not edit by hand — regenerate it instead, and see `CONTRIBUTING.md` for when this is required.

This is what this provider version implements, derived from the same catalog the provider binary itself exposes through `sapintegrationsuite_provider_features` (whose `provider_version` attribute reports the exact version at apply time). It is queryable directly from Terraform, with no SAP tenant connection required:

```hcl
data "sapintegrationsuite_provider_features" "all" {}

output "supported_features" {
  value = [
    for f in data.sapintegrationsuite_provider_features.all.features :
    f.key
    if f.support_status == "supported"
  ]
}

data "sapintegrationsuite_provider_feature" "one" {
  key = "cloud_integration.value_mapping"
}
```

## Relationship to the other capability documents

This document, `docs/api-capability-matrix.md`, `docs/provisioning-capability-matrix.md`, and a possible future tenant-capability data source answer three related but distinct questions:

| Document | Answers |
|---|---|
| `docs/api-capability-matrix.md` / `docs/provisioning-capability-matrix.md` | What SAP's public APIs expose, independent of this provider |
| `docs/feature-support.md` (this document) / `sapintegrationsuite_provider_features` | What *this provider version* implements — static provider metadata, no SAP tenant required |
| A possible future `sapintegrationsuite_tenant_capabilities` data source (not implemented) | Which Integration Suite capabilities are *active in a specific SAP tenant* — would require SAP credentials and a reliable public discovery API, neither of which this provider assumes here |

Do not conflate these: a feature can be fully supported by this provider and still be unusable in a given tenant because the underlying SAP capability was never activated there, and vice versa a capability can be active in every tenant while this provider still does not implement a resource for it.

## All features

| Feature | Domain | Status | Public API | Create | Read | Update | Delete | Import | Deploy | Terraform |
|---|---|---|---|---|---|---|---|---|---|---|
| `api_gateway.api_artifact` | api_gateway | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `api_gateway.api_artifact_deployment` | api_gateway | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `api_gateway.api_policy` | api_gateway | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `api_management.classic.api_product` | api_management_classic | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `api_management.classic.api_provider` | api_management_classic | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `api_management.classic.api_proxy` | api_management_classic | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `api_management.classic.key_value_map` | api_management_classic | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `capabilities.api_gateway` | capability_provisioning | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `capabilities.api_management` | capability_provisioning | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `capabilities.cloud_integration` | capability_provisioning | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `capabilities.edge_integration_cell` | capability_provisioning | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `capabilities.integration_cell` | capability_provisioning | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `cloud_integration.integration_flow` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource |
| `cloud_integration.integration_flow_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.integration_package` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.message_mapping` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.message_mapping_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.message_processing_logs` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.message_stores` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.script_collection` | cloud_integration | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.service_endpoints` | cloud_integration | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.value_mapping` | cloud_integration | partial (unsafe_terraform_lifecycle) | Yes | Yes | Yes | — | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.value_mapping_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.value_mapping_entry` | cloud_integration | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `data_space_integration` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `edge_integration_cell.registration` | edge_integration_cell | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `event_mesh` | other_capability | unsupported (out_of_scope) | No | — | — | — | — | — | — | — |
| `integration_advisor` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `integration_assessment` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `integration_cell.runtime` | integration_cell | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `migration_assessment` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `open_connectors` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `partner_directory.entry` | partner_directory | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `security.access_policy` | security | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource |
| `security.access_policy_reference` | security | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource |
| `security.certificate_user_mapping` | security | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `security.keystore_entry` | security | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `security.oauth2_client_credential` | security | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `security.user_credential` | security | unsupported (not_implemented) | Yes | — | — | — | — | — | — | — |
| `trading_partner_management` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |

## Unsupported and partially supported features

Grouped by why, not just that. A feature can be `partial` and reachable via one of these reasons too — see docs/feature-support.md's per-feature `Limitations` (the `limitations` attribute in Terraform) for exactly what is and is not covered.

### Public API exists but provider implementation is pending

- **`api_management.classic.api_product`** — A classic API Management API product bundling one or more API proxies.
- **`api_management.classic.api_provider`** — A classic API Management backend/API provider system definition.
- **`api_management.classic.api_proxy`** — A classic API Management API proxy definition.
- **`api_management.classic.key_value_map`** — A classic API Management key-value map used for runtime configuration lookups.
- **`cloud_integration.script_collection`** — A reusable collection of Groovy/JavaScript scripts shared across multiple integration flows.
  - A full design-time/deployment API mirroring Integration Flow's is documented; this provider has not implemented it yet.
- **`cloud_integration.service_endpoints`** — Read-only lookup of a deployed integration flow's exposed runtime service endpoint URLs.
- **`partner_directory.entry`** — A trading-partner-style directory entry used by B2B-oriented integration flows.
- **`security.certificate_user_mapping`** — A mapping from a client certificate to an inbound user identity.
- **`security.keystore_entry`** — A certificate keystore entry security material artifact.
- **`security.oauth2_client_credential`** — An OAuth2 client credential security material artifact used by integration flow adapters.
- **`security.user_credential`** — A user credential security material artifact used by integration flow adapters.
  - Secret values are never returned by SAP's read API; requires write-only attribute semantics not yet designed.

### Public lifecycle insufficient for safe Terraform management

- **`cloud_integration.value_mapping`** — A value mapping design-time artifact's content, managed as file-based content. (partial support already implemented — see Limitations below)
  - No confirmed in-place update: changing name, content, or content_hash replaces the resource (create a new artifact, then delete the old one) instead of calling an unverified PUT.
  - SAP separately documents a ValueMappingDesigntimeArtifactSaveAsVersion action this provider does not yet use.
  - Whether Delete removes only the active version or every version of the artifact is unconfirmed against a primary source.

### No suitable public SAP API

- **`capabilities.api_gateway`** — Activating the newer API Gateway / API Artifacts capability itself within an Integration Suite tenant.
- **`capabilities.api_management`** — Activating the classic API Management capability itself within an Integration Suite tenant.
  - Documented as a UI step (Integration Suite → Manage Capabilities → activate API Management); no public API found.
- **`capabilities.cloud_integration`** — Activating the Cloud Integration capability itself within an Integration Suite tenant.
  - Activated as part of the Integration Suite subscription/onboarding wizard; stays a manual, one-time bootstrap step.
- **`capabilities.edge_integration_cell`** — Activating the Edge Integration Cell capability itself within an Integration Suite tenant.
- **`capabilities.integration_cell`** — Activating the Integration Cell capability itself within an Integration Suite tenant.
  - SAP's own documentation describes activation as choosing Activate in Integration Suite → Settings → Runtime; no public API was found.
- **`edge_integration_cell.registration`** — SAP-side registration and runtime association of an Edge Integration Cell, never the customer-managed Kubernetes workloads themselves.
  - Registration is a guided UI plus Helm-based bootstrap process; no public capability-activation or registration API was found.
- **`integration_cell.runtime`** — Runtime status and configuration of an already-activated Integration Cell, distinct from activating the capability itself.
  - No public status or configuration API was found for Integration Cell runtime content, only UI-facing operations.

### Further research required

- **`api_gateway.api_artifact`** — An API-artifact-centric design-time object in SAP's newer API Gateway model.
  - Existence confirmed via UI/feature documentation only; a public design-time API has not been confirmed in enough detail for a stable Terraform schema.
- **`api_gateway.api_artifact_deployment`** — The runtime deployment state of an API Gateway API artifact.
- **`api_gateway.api_policy`** — A policy attached to an API Gateway API artifact.
- **`cloud_integration.value_mapping_entry`** — Individual source/target value pairs inside a value mapping scheme, managed through UpsertValMaps, UpdateDefaultValMap, and DeleteValMaps.
  - Exact request/response payload shapes and DeleteValMaps' delete granularity are not confirmed against a reachable primary source.
  - UpsertValMaps requires an already-existing source/target agency-identifier scheme, so entries cannot be managed independently of the artifact's own content.
- **`data_space_integration`** — SAP's data space connectivity capability within Integration Suite.
- **`integration_advisor`** — SAP's collaborative interface-content-design capability.
- **`integration_assessment`** — SAP's integration landscape assessment capability.
- **`migration_assessment`** — SAP's integration migration assessment capability.
- **`open_connectors`** — SAP's third-party SaaS connectivity capability within Integration Suite.
  - Not yet investigated by this project; listed so its absence is visible rather than silently omitted.
- **`trading_partner_management`** — SAP's B2B trading partner management capability.

### Out of provider scope

- **`cloud_integration.message_processing_logs`** — Runtime message processing log records for deployed integration flows.
  - Monitoring/operational data, not infrastructure state this provider manages — see docs/provider-scope.md.
- **`cloud_integration.message_stores`** — Runtime message queue and data store payload content used by deployed integration flows.
  - Payload/queue content is operational data, not desired state this provider manages.
- **`event_mesh`** — SAP's event broker service for asynchronous, event-driven integration.
  - Likely a separate BTP service outside this provider's Integration Suite content/capability boundary rather than a Cloud Integration design-time concern.

