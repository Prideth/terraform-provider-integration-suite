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

See README.md's "Feature Support" section for a compact, high-level dashboard generated from this same catalog (`go run ./cmd/gendocs -readme`); this document is the detailed per-operation matrix.

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
| `cloud_integration.custom_tag_configuration` | cloud_integration | partial (unsafe_terraform_lifecycle) | Yes | Yes | Yes | Yes | — | Yes | — | Resource + Data Source |
| `cloud_integration.data_store` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.data_store_entry` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.integration_adapter` | cloud_integration | partial (public_api_incomplete) | Yes | Yes | Yes | — | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.integration_adapter_deployment` | cloud_integration | partial (public_api_incomplete) | Yes | Yes | Yes | — | Yes | — | Yes | Resource |
| `cloud_integration.integration_flow` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource |
| `cloud_integration.integration_flow_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.integration_package` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.message_mapping` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.message_mapping_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.message_processing_logs` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.message_stores` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.number_range` | cloud_integration | partial (unsafe_terraform_lifecycle) | Yes | Yes | — | Yes | — | — | — | Resource |
| `cloud_integration.script_collection` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.script_collection_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.service_endpoints` | cloud_integration | read_only (unsafe_terraform_lifecycle) | Yes | — | Yes | — | — | — | — | Data Source |
| `cloud_integration.value_mapping` | cloud_integration | partial (unsafe_terraform_lifecycle) | Yes | Yes | Yes | — | Yes | Yes | — | Resource + Data Source |
| `cloud_integration.value_mapping_deployment` | cloud_integration | supported | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Resource |
| `cloud_integration.value_mapping_entry` | cloud_integration | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `cloud_integration.variable` | cloud_integration | unsupported (out_of_scope) | Yes | — | — | — | — | — | — | — |
| `data_space_integration` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `edge_integration_cell.registration` | edge_integration_cell | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `event_mesh` | other_capability | unsupported (out_of_scope) | No | — | — | — | — | — | — | — |
| `integration_advisor` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `integration_assessment` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `integration_cell.runtime` | integration_cell | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `migration_assessment` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `open_connectors` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `partner_directory.alternative_partner` | partner_directory | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `partner_directory.authorized_user` | partner_directory | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `partner_directory.binary_parameter` | partner_directory | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `partner_directory.partner` | partner_directory | read_only (unsafe_terraform_lifecycle) | Yes | — | Yes | — | — | — | — | Data Source |
| `partner_directory.string_parameter` | partner_directory | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `partner_directory.user_credential_parameter` | partner_directory | partial (unsafe_terraform_lifecycle) | Yes | Yes | Yes | — | Yes | Yes | — | Resource |
| `security.access_policy` | security | supported | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `security.access_policy_reference` | security | supported | Yes | Yes | Yes | — | Yes | Yes | — | Resource + Data Source |
| `security.certificate` | security | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `security.certificate_chain` | security | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `security.certificate_user_mapping` | security | unsupported (no_public_api) | No | — | — | — | — | — | — | — |
| `security.key_pair` | security | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `security.keystore_entry` | security | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `security.known_hosts` | security | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `security.oauth2_client_credential` | security | partial (public_api_incomplete) | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `security.secure_parameter` | security | unsupported (research_required) | No | — | — | — | — | — | — | — |
| `security.ssh_key` | security | unsupported (research_required) | Yes | — | — | — | — | — | — | — |
| `security.user_credential` | security | partial (public_api_incomplete) | Yes | Yes | Yes | Yes | Yes | Yes | — | Resource + Data Source |
| `trading_partner_management` | other_capability | unsupported (research_required) | No | — | — | — | — | — | — | — |

## Unsupported and partially supported features

Grouped by why, not just that. A feature can be `partial` and reachable via one of these reasons too — see docs/feature-support.md's per-feature `Limitations` (the `limitations` attribute in Terraform) for exactly what is and is not covered.

### Public API exists but provider implementation is pending

- **`api_management.classic.api_product`** — A classic API Management API product bundling one or more API proxies.
- **`api_management.classic.api_provider`** — A classic API Management backend/API provider system definition.
- **`api_management.classic.api_proxy`** — A classic API Management API proxy definition.
- **`api_management.classic.key_value_map`** — A classic API Management key-value map used for runtime configuration lookups.

### Public API details are not fully confirmed

- **`cloud_integration.integration_adapter`** — A custom Integration Adapter design-time artifact (a *.esa archive built with the SAP Adapter SDK), imported into a Cloud Integration package. Cloud Foundry environment only. (partial support already implemented — see Limitations below)
  - This provider's evidence base for this entity is thinner than for the sibling design-time artifact types it manages: SAP's own "Integration Adapter Example Requests, Cloud Foundry Environment" documentation shows only Delete (confirming the entity is keyed by Id alone, not the composite (Id, Version) key every other design-time artifact type in this API uses) and the Deploy action — no Create or Read example was found. Create is implemented by strong analogy to every sibling artifact type's confirmed PackageId/ArtifactContent POST body shape, corroborated by a third-party technical source, not by an SAP-published example request for this specific entity. See docs/guides/integration-adapters.md.
  - No in-place update: SAP documents that importing an ID that already exists on the tenant is rejected as an error, which is positive evidence against a working reimport-to-update flow, so every attribute is RequiresReplace rather than an unverified PUT/PATCH.
  - type and application are not validated against a fixed set of values: SAP's documentation confirms both exist as UI dropdowns with example values (Analytics, CRM, ERP, ... for type; Slack, ... for application) but does not confirm whether the underlying property is a closed enum or free-form text.
  - Distinct from SAP Business Accelerator Hub prebundled adapters (a different import/auto-deploy lifecycle reached from inside the integration flow editor) and from Integration Suite capability activation — see docs/guides/integration-adapters.md for why these are not the same feature.
  - There is no sapintegrationsuite_integration_adapters collection data source: no confirmed list/filter contract for this entity set was found, unlike ServiceEndpoints' documented Name/Protocol filters.
- **`cloud_integration.integration_adapter_deployment`** — The runtime deployment state of a custom Integration Adapter, independent of its design-time content lifecycle. (partial support already implemented — see Limitations below)
  - Deploy is confirmed directly from SAP's own example request: POST DeployIntegrationAdapterDesigntimeArtifact?Id='...', singular "Artifact" (matching every sibling deploy action in this API), with no Version query parameter (unlike every sibling deploy action, consistent with Id being this entity's only confirmed key).
  - Runtime status polling and undeploy reuse the same shared IntegrationRuntimeArtifacts entity every other *_deployment resource in this provider polls/undeploys through, by analogy — this project could not independently confirm that a deployed custom adapter surfaces through that same shared entity as opposed to an adapter-specific status/undeploy mechanism (for example BuildAndDeployStatus). See docs/guides/integration-adapters.md.
  - Whether a fresh deployment's runtime status becomes visible immediately or after a build/deploy delay specific to adapters (as opposed to ordinary content deployment) was not confirmed.
- **`security.oauth2_client_credential`** — An "OAuth2 Client Credentials" security material artifact: the client ID, client secret, and token service URL an integration flow adapter uses for the OAuth2 client credentials grant (RFC 6749) on outbound requests. (partial support already implemented — see Limitations below)
  - The client secret is never returned by SAP's read API; client_secret_wo/client_secret_wo_version are write-only attributes (Terraform CLI 1.11+ required) and drift on the secret value itself cannot be detected.
  - Only name, description, token_service_url, client_id, client_secret, and scope are exposed: SAP's UI additionally documents Grant Type placement, Client Authentication mode (body vs. header), Resource, Audience, and up to 20 custom parameters, but this project could not confirm their OData property names against $metadata or a documented example payload, so they are deliberately unimplemented rather than guessed.
  - Update is implemented as a full PUT redeploy and resends client_secret_wo on every apply that touches this resource, matching SAP's documented requirement to re-enter the client secret on every edit.
  - OAuth2 Authorization Code and OAuth2 SAML Bearer Assertion are separate SAP artifact types this provider does not implement: Authorization Code requires interactive human authorization (see security.oauth2_authorization_code note in docs/guides/security-content.md), and SAML Bearer Assertion's public API contract was not confirmed.
- **`security.user_credential`** — A "User Credentials" security material artifact: a username/password credential integration flow adapters use for outbound basic or username-token authentication. (partial support already implemented — see Limitations below)
  - The password is never returned by SAP's read API; password_wo/password_wo_version are write-only attributes (Terraform CLI 1.11+ required) and drift on the password value itself cannot be detected — only readable metadata (user, description, kind, company_id) is compared on Read.
  - Update is implemented as a full PUT redeploy, matching SAP's documented "Edit" action for Credentials artifacts, and resends password_wo on every apply that touches this resource (SAP documents re-entering the secret on every edit for the sibling OAuth2 Client Credentials artifact; this provider assumes the same requirement here since it could not find a documented exception for User Credentials).
  - The Kind and CompanyId field names are corroborated by a documented third-party example payload, not by this project's own inspection of a live tenant's OData $metadata; verify against your tenant before relying on kind="SuccessFactors"/"OpenConnectors" in production. See docs/guides/security-content.md.
  - Deployment status (SAP's UI shows Stored/Deployed/Error) is not exposed: this project could not confirm the OData property name for it, and would rather omit a computed attribute than expose one that is silently always empty.

### Public lifecycle insufficient for safe Terraform management

- **`cloud_integration.custom_tag_configuration`** — The tenant-wide set of custom tags integration package owners are asked, or required, to classify their packages with. (partial support already implemented — see Limitations below)
  - No confirmed delete or clear operation exists for this entity anywhere in SAP's public documentation. Destroying this resource in Terraform returns an explicit error rather than guessing that an empty overwrite means delete, or silently dropping Terraform state while leaving the tenant's configuration untouched — see docs/guides/custom-tag-configurations.md.
  - Create and Update both use the same confirmed POST .../CustomTagConfigurations?Overwrite=true operation (SAP documents no separate plain-POST-without-Overwrite path this provider relies on), sending the complete desired tag list every time. Whether Overwrite=true is a full replace (removing tags not present in the new list) is strongly implied by the word "Overwrite" and by the fact the documented payload is the complete configuration, not a delta, but SAP's documentation never uses the word "replace" explicitly.
  - Whether tag names must be unique, whether permitted values are case-sensitive, and whether SAP preserves submitted ordering are all unconfirmed by SAP's documentation. This provider enforces tag-name uniqueness itself and treats ordering (of both tags and permitted values) as not semantically meaningful, modeling both as unordered Terraform sets so a reordered response never produces a spurious diff.
  - One documented example response shows a single-element permittedValues array containing a comma-separated string ("Mr. Bean, Ms. Bean") rather than two separate array elements; this is treated as a documentation artifact, not a confirmed wire format, since every other array-typed field in SAP's own examples (and everywhere else in this provider) uses one array element per value.
- **`cloud_integration.number_range`** — A Number Ranges object: generates unique interchange numbers for outbound EDI/EDIFACT documents, with a static configuration (min/max/description/rotate/field length) and a live runtime counter (CurrentValue, the UI's "Next Value") that advances as deployed content consumes numbers. (partial support already implemented — see Limitations below)
  - SAP documents no GET operation for this entity anywhere — unlike every sibling entity in the same Message Stores API family (DataStores, DataStoreEntries, Variables all have documented GET examples), NumberRanges has none in SAP's curated "Message Stores Example Requests" index or anywhere else this project found. Without a GET, this resource's Read is a documented no-op that trusts local state rather than verifying anything against the tenant: it cannot detect drift, and 'terraform import' is rejected outright rather than silently leaving most attributes unknown.
  - SAP documents no delete operation for this entity either; the Monitor UI shows an "Undeploy" action with no confirmed REST equivalent. 'terraform destroy' returns an explicit error rather than guessing at an unconfirmed operation — see docs/guides/runtime-stores-and-number-ranges.md.
  - The runtime counter is handled as a write-only, version-gated attribute (current_value_wo / current_value_wo_version), pushed to SAP only when the version marker changes. An ordinary Update that only changes description/min_value/max_value/rotate/field_length omits CurrentValue from the request body entirely, rather than resending a value this provider has no way to confirm is still current — SAP's documentation does not confirm whether an omitted field on this entity's PUT is preserved unchanged or reset, which is an inherent, documented risk of this design, not a guess this provider is hiding.
  - The UI additionally documents a "Runtimes" multi-select deployment field (Cloud Integration plus any active Edge Integration Cell nodes) with no visible counterpart in either of SAP's two documented API examples (Add, Update); this provider's client always targets the default runtime implicitly and does not expose runtime/location selection.
- **`cloud_integration.service_endpoints`** — Read-only discovery of the runtime service endpoints (entry point URLs and API definition links) SAP generates for deployed Cloud Integration content.
  - Discovery only, by design: SAP generates service endpoints from deployed content and there is no create/update/delete API for them, so this provider intentionally has no matching resource type — see docs/guides/service-endpoints.md.
  - No single-endpoint (sapintegrationsuite_service_endpoint) data source exists: this project could not confirm that Name uniquely and stably identifies exactly one service endpoint, so only the collection data source (sapintegrationsuite_service_endpoints, with optional name/protocol filters) is implemented, to avoid a lookup data source that silently returns the wrong result if more than one endpoint ever matches.
  - The EntryPoint/APIDefinition Url property's JSON casing is confirmed from SAP's own open-source Piper library parsing a live response; the ApiDefinitions entity's Url casing specifically is inferred by consistency rather than independently confirmed from an example touching that entity — see docs/sap-api-references.md.
- **`cloud_integration.value_mapping`** — A value mapping design-time artifact's content, managed as file-based content. (partial support already implemented — see Limitations below)
  - No confirmed in-place update: changing name, content, or content_hash replaces the resource (create a new artifact, then delete the old one) instead of calling an unverified PUT.
  - SAP separately documents a ValueMappingDesigntimeArtifactSaveAsVersion action this provider does not yet use.
  - Whether Delete removes only the active version or every version of the artifact is unconfirmed against a primary source.
- **`partner_directory.partner`** — A Partner ID (Pid) known to the tenant's Partner Directory.
  - No resource: SAP documents no confirmed create operation for Partners — a Pid comes into existence implicitly the first time a StringParameter, BinaryParameter, AlternativePartner, AuthorizedUser, or UserCredentialParameter references it.
  - Deleting a Pid is documented as cascading to every entity belonging to it, which is the other reason this stays read-only: a Partner resource's Destroy could erase content owned by an entirely different Terraform module.
- **`partner_directory.user_credential_parameter`** — A communication username/password credential scoped to a Partner ID (Pid). (partial support already implemented — see Limitations below)
  - The password is a write-only attribute (password_wo): Terraform never stores it in plan or state, and this provider never requests or reads a password back from SAP, which does not document returning one. Requires Terraform CLI 1.11 or later.
  - No in-place update: no public API for changing an existing credential's password was confirmed, so rotating it (via the paired password_wo_version attribute) replaces the resource — delete the old credential, then create a new one.
  - UserCredentialParameter cannot be combined with other Partner Directory entity types in a single OData batch (ChangeSet) request; this provider always issues it standalone.
  - Import recovers partner_id, parameter_id, and user, but never the password: a configuration applied right after import must still supply password_wo and a password_wo_version, which plans as a replacement even though nothing server-side has actually changed.

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
- **`security.certificate_user_mapping`** — A mapping from a client certificate to an inbound user identity, used for inbound client certificate authentication.
  - Reverified for this feature family: SAP's own documentation ("Managing Certificate-to-User Mappings", "Client Certificate Authentication and Certificate-to-User Mapping (Inbound)", "Setting Up Inbound HTTP Connections with Certificate-to-User Mapping") exists only under the Neo environment, with no Cloud Foundry equivalent found in SAP's published documentation set. This provider targets the Cloud Foundry environment (its other Security Content resources use the Cloud Foundry "/api/v1" OData host), so this catalog entry is corrected from its previous "not_implemented"/PublicAPI:true state to "no_public_api": the feature cannot be implemented for this provider's target environment, not merely unimplemented yet. If SAP publishes a Cloud Foundry certificate-to-user-mapping API in the future, re-open this entry.

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
- **`security.certificate`** — A standalone X.509 certificate keystore entry (as opposed to a key pair), typically an uploaded root or intermediate CA certificate.
  - SAP's UI documents uploading a certificate to the keystore, but this project could not confirm the OData create/update request shape (entity set name, PEM/DER encoding expectations, alias field) with enough confidence to implement it safely.
- **`security.certificate_chain`** — A certificate chain resource associated with a key pair (CertificateChainResources per the public API catalog).
  - This project could not confirm this entity's identity, upload/download semantics, or relationship to security.key_pair against $metadata or documented examples.
- **`security.key_pair`** — An SAP-generated key pair keystore entry (private key plus X.509 certificate chain), as opposed to one uploaded from outside the tenant.
  - SAP's UI documents key pair generation (KeyPairGenerationRequests / KeyPairResources per the public API catalog), and private keys generated this way are never downloadable, which would make this a safe design (the resource can manage a key pair's existence and metadata without ever handling private key material). This project could not confirm the generation request's exact fields (key algorithm, key size, distinguished name) or its synchronous-vs-asynchronous lifecycle against $metadata or a documented example, so it is not yet implemented.
- **`security.keystore_entry`** — A certificate or key pair entry in the tenant's keystore (KeystoreEntries, Keystores, KeystoreResources, HistoryKeystoreEntries in SAP's Security Content API).
  - SAP's UI documentation (Keystore Monitor) and independent technical sources confirm a public KeystoreEntries OData entity set exists with fields resembling Alias, Type, ValidUntil/ValidNotAfter, SubjectDN, IssuerDN, KeyType, KeySize, SerialNumber, SignatureAlgorithm, and one or more Fingerprints, but this project could not confirm the exact property names and casing against $metadata, so no resource or data source is implemented yet rather than shipping a guessed field mapping. Read-only metadata discovery (a data source, not a resource) is the recommended next step: see docs/guides/security-content.md and ROADMAP.md.
  - HistoryKeystoreEntries is audit/history information, not a mutable artifact, and should only ever become a read-only data source if implemented, never a resource.
- **`security.known_hosts`** — The SSH "known_hosts" file artifact used to validate SFTP server host keys for outbound SFTP connections.
  - SAP's Manage Security Material UI documents uploading and downloading a Known Hosts artifact (file content, not a structured entity with individually settable fields), but this project could not confirm a public OData entity set or REST endpoint for it with enough confidence to implement create/update/delete safely.
- **`security.secure_parameter`** — A "Secure Parameter" security material artifact: an opaque confidential value (for example for a custom adapter) deployed without an associated username.
  - SAP's Manage Security Material UI documents creating and deploying a Secure Parameter artifact, but this project found third-party evidence of at least one practitioner receiving an OData error ("could not find an entity set or function import for SecureParameters") when attempting to call it through the Security Content API, suggesting the entity set name is different from the obvious guess, is not exposed in every API version, or is not publicly documented at all. Marked PublicAPI: false pending confirmation, not because the UI feature doesn't exist, but because a callable public OData contract for it was not confirmed. If a public contract is confirmed, this would be a strong write-only-attribute candidate (value_wo/value_wo_version), the same shape as security.user_credential's password.
- **`security.ssh_key`** — An SAP-generated SSH key pair keystore entry, used for SFTP public-key authentication (SSHKeyGenerationRequests per the public API catalog).
  - Same reasoning as security.key_pair: SAP's UI documents SSH key pair creation and this provider's standing policy prefers modeling a declarative, persistent resource over a one-shot "generate" action, but this project could not confirm the request/response contract well enough to implement it yet.
- **`trading_partner_management`** — SAP's B2B trading partner management capability.

### Out of provider scope

- **`cloud_integration.data_store`** — A tenant-persisted runtime container of Data Store Entries, created implicitly by an integration flow's Data Store Write step (or an XI adapter's Temporary Storage option) the first time it writes an entry.
  - SAP documents exactly one public operation for this entity: GET .../DataStores?overdueonly=true, an aggregate monitoring endpoint (NumberOfMessages/NumberOfOverdueMessages per store) — the same class of runtime monitoring data as cloud_integration.message_processing_logs, not configuration. There is no independent declarative creation API: a Data Store comes into existence only as a side effect of deployed integration flow content.
  - The DataStores API does not support $filter, $inlinecount, $orderby, $skip, $top, $expand, or $select (confirmed directly from SAP's own documentation).
- **`cloud_integration.data_store_entry`** — A single runtime message (payload and headers) persisted inside a Data Store by an integration flow's Data Store Write step, read back by a Data Store Get or Select step, and deleted only by a Data Store Delete step.
  - SAP documents only GET operations for this entity (a single entry by composite key, all entries for a store, and all entries for a message ID) — every field (Status, MessageId, DueAt, CreatedAt, RetainUntil) is runtime business-message state, not infrastructure desired state. Delete exists only as a design-time integration flow step (entry-by-entry or bulk via an XPath-derived ID list at runtime), never as a REST call this provider could wrap in a Terraform destroy.
  - Deliberately out_of_scope rather than not_implemented: this provider does not manage business message payloads or place them into Terraform state, and 'terraform destroy' semantics are not an excuse to expose operational message deletion as desired infrastructure state — see docs/provider-scope.md.
- **`cloud_integration.message_processing_logs`** — Runtime message processing log records for deployed integration flows.
  - Monitoring/operational data, not infrastructure state this provider manages — see docs/provider-scope.md.
- **`cloud_integration.message_stores`** — Runtime persisted-message-store entries (created by the Persist step) and JMS queue resource metadata used by deployed integration flows. Data Stores, Data Store Entries, Variables, and Number Ranges — all part of the same broader Message Stores API family — each have their own dedicated catalog entry; see cloud_integration.data_store, cloud_integration.data_store_entry, cloud_integration.variable, and cloud_integration.number_range.
  - Payload/queue content is operational data, not desired state this provider manages.
- **`cloud_integration.variable`** — A tenant-persisted runtime value written by an integration flow's "Write Variables" step, shared across steps of the same flow (local) or across every flow deployed on the tenant (global).
  - SAP documents exactly one public operation for this entity: GET .../Variables(...)/$value, which downloads the raw value with no structured metadata (no Visibility/ UpdatedAt/RetainUntil fields are returned by this endpoint). There is no collection GET, no POST, no PUT, and no confirmed DELETE — Variables are created and updated exclusively by deployed integration flow content, an entirely different ownership domain than Terraform-managed infrastructure.
  - A read-only data source was deliberately not implemented: the only confirmed read operation returns nothing but the raw runtime value itself, with no safer metadata-only alternative available, and this provider does not place arbitrary runtime business values into Terraform state merely because an API can return them — see docs/provider-scope.md.
- **`event_mesh`** — SAP's event broker service for asynchronous, event-driven integration.
  - Likely a separate BTP service outside this provider's Integration Suite content/capability boundary rather than a Cloud Integration design-time concern.

