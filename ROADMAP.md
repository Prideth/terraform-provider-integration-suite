# Roadmap

This roadmap is adjusted after each round of API discovery documented under `docs/`. It
favors a small, high-quality resource set over broad but shallow coverage. Status labels used
below: **Implemented**, **Partial**, **Next**, **Planned**, **Blocked by API**, **Research
required**, **Out of scope**.

For what this provider *version* actually supports today, feature by feature, see
[`docs/feature-support.md`](docs/feature-support.md) and the
`sapintegrationsuite_provider_features` / `sapintegrationsuite_provider_feature` data sources —
this document describes direction and sequencing, not the authoritative current support
matrix.

## Current development priority

Development branches from `dev`, in this order, until superseded by an explicit reprioritization:

1. **Current API Management / API Artifacts / Integration Cell** — **Current.** SAP's current
   API-centric integration model (API Artifacts, Runtime Profiles, Integration Cell, Virtual
   Hosts, Policies, Reusable API Artifacts) — a distinct product model from Classic API
   Management (API Providers/Proxies/Products), never mixed in this provider's terminology or
   code. See `docs/guides/current-api-management.md` for the full public-API-boundary research
   and every resulting Terraform decision.
2. **Additional current API Management subfeatures** — whatever this phase's research
   identifies as publicly automatable but out of scope for the first pass (for example Edge
   Integration Cell-specific Virtual Host/deployment behavior, or a confirmed Policy API found
   too late to implement this phase).
3. **Classic API Management** — Planned, explicitly *after* the current API Artifact model,
   reflecting where SAP's own product direction now sits. Moved behind current API Management
   deliberately; still a real, tracked priority, not dropped.
4. **Edge Integration Cell** — Blocked by API for SAP control-plane/runtime-specific
   configuration; never a Kubernetes/Helm replacement. Deliberately deferred from this phase even
   though it shares infrastructure with Integration Cell (Virtual Hosts, Runtime Profiles) — see
   `docs/guides/current-api-management.md` for what, if anything, this phase found reusable for
   it later.
5. **Additional Integration Suite capabilities** — Research required: Integration Advisor,
   Trading Partner Management, Integration Assessment, Migration Assessment, and others, only
   where public APIs justify Terraform management.

Access Policies completion (role/identity semantics, artifact reference lifecycle audit,
runtime reconciliation research, minimal-PATCH update, data sources, import/drift) is done —
see "Implemented" below and `docs/guides/access-policies.md`. Partner Directory is also done —
see "Implemented" below and `docs/guides/partner-directory.md`. Cloud Integration Service
Endpoints discovery is also done — see "Implemented" below and `docs/guides/service-endpoints.md`.
Security Content (User Credentials, OAuth2 Client Credentials, Keystore Entry discovery,
Certificates, and SAP-generated Key Pairs), custom Integration Adapter design-time/deployment
support, and Custom Tag Configuration management are also done, each at a `partial` or better
support level — see "Implemented" below, `docs/guides/security-content.md`,
`docs/guides/integration-adapters.md`, and `docs/guides/custom-tag-configurations.md`. Number
Ranges / Variables / Data Stores is also done — a narrow write-only-lifecycle Number Range
resource (SAP documents no `GET`/`DELETE` for it at all), with Variables, Data Stores, and Data
Store Entries deliberately left unsupported/out of scope — see "Implemented" below and
`docs/guides/runtime-stores-and-number-ranges.md`.

This order describes what to work on **next**; it does not retroactively unimplement anything
already shipped (see "Implemented" below).

## Implemented (v0.1.x — Provider Foundation)

- Provider skeleton on the Terraform Plugin Framework, named `sapintegrationsuite`
- OAuth 2.0 client credentials authentication with a thread-safe, context-aware token cache
- Shared HTTP client: retry with backoff/jitter, `Retry-After` handling, bounded response
  sizes, custom User-Agent
- OData V2 request/query/pagination/error-parsing foundation
- Cloud Integration fundamentals:
  - `sapintegrationsuite_integration_package`
  - `sapintegrationsuite_integration_flow`
  - `sapintegrationsuite_integration_flow_deployment`
- Access Policies (role/artifact-reference lifecycle audited and completed; minimal-PATCH
  update, data sources, and runtime-reconciliation research documented — see
  `docs/guides/access-policies.md`):
  - `sapintegrationsuite_access_policy`
  - `sapintegrationsuite_access_policy_reference`
  - `data.sapintegrationsuite_access_policy`
  - `data.sapintegrationsuite_access_policy_reference`
- Value Mappings:
  - `sapintegrationsuite_value_mapping`
  - `sapintegrationsuite_value_mapping_deployment`
- Message Mappings (reusable, package-level artifacts — not the inline/local mapping step an
  integration flow can also define directly, see `docs/resource-design.md`):
  - `sapintegrationsuite_message_mapping`
  - `sapintegrationsuite_message_mapping_deployment`
  - `data.sapintegrationsuite_message_mapping`
- Script Collections (reusable Groovy/JavaScript bundles, same design-time/runtime split):
  - `sapintegrationsuite_script_collection`
  - `sapintegrationsuite_script_collection_deployment`
  - `data.sapintegrationsuite_script_collection`
- Partner Directory (string/binary parameters, alternative partners, authorized users, and a
  security-sensitive write-only user credential parameter resource — see
  `docs/guides/partner-directory.md`; no `sapintegrationsuite_partner` resource exists, since
  SAP documents no confirmed create operation for it, only discovery data sources):
  - `sapintegrationsuite_partner_string_parameter`
  - `sapintegrationsuite_partner_binary_parameter`
  - `sapintegrationsuite_alternative_partner`
  - `sapintegrationsuite_partner_authorized_user`
  - `sapintegrationsuite_partner_user_credential_parameter`
  - `data.sapintegrationsuite_partner`
  - `data.sapintegrationsuite_partners`
  - `data.sapintegrationsuite_partner_string_parameter`
  - `data.sapintegrationsuite_partner_string_parameters`
  - `data.sapintegrationsuite_partner_binary_parameter`
  - `data.sapintegrationsuite_alternative_partner`
  - `data.sapintegrationsuite_partner_authorized_user`
- Security Content credentials (write-only secrets, in-place redeploy via `PUT` — see
  `docs/guides/security-content.md`; most other Security Content artifact types remain
  unimplemented, see "Current development priority" above):
  - `sapintegrationsuite_user_credential`
  - `data.sapintegrationsuite_user_credential`
  - `sapintegrationsuite_oauth2_client_credential`
  - `data.sapintegrationsuite_oauth2_client_credential`
- Security Content keystore management (read-only entry discovery, X.509 certificates, and
  SAP-generated key pairs — private key material never enters this provider; certificate drift
  detection uses a locally-computed SHA-256 fingerprint, not raw PEM text; no separate "SSH Key"
  resource, since SAP's own documentation treats it as the same Key Pair mechanism — see
  `docs/guides/security-content.md`):
  - `data.sapintegrationsuite_keystore_entry`
  - `data.sapintegrationsuite_keystore_entries`
  - `sapintegrationsuite_certificate`
  - `sapintegrationsuite_key_pair`
- Cloud Integration Service Endpoints discovery (read-only by design — SAP generates these from
  deployed content, so there is no matching resource; combined `EntryPoints`/`ApiDefinitions`
  expansion, `name`/`protocol` filters, full server-driven pagination, deterministic ordering —
  see `docs/guides/service-endpoints.md`):
  - `data.sapintegrationsuite_service_endpoints`
- Custom Integration Adapter design-time and deployment support (Cloud Foundry only;
  conservative replace-on-any-change model since SAP documents a duplicate-ID import as an
  error and no in-place update was confirmed; no `*_version` attribute on the deployment
  resource, since the confirmed deploy action takes no `Version` parameter — see
  `docs/guides/integration-adapters.md`):
  - `sapintegrationsuite_integration_adapter`
  - `data.sapintegrationsuite_integration_adapter`
  - `sapintegrationsuite_integration_adapter_deployment`
- Custom Tag Configuration management (a tenant-wide singleton; Create/Read/Update confirmed
  verbatim against SAP's own documented example payloads; Delete deliberately returns an
  explicit error rather than a guessed clearing mechanism, since SAP documents no delete or
  clear operation for this entity at all — see `docs/guides/custom-tag-configurations.md`):
  - `sapintegrationsuite_custom_tag_configuration`
  - `data.sapintegrationsuite_custom_tag_configuration`
- Number Range configuration (a write-only-lifecycle resource: SAP documents Create/Update but
  no `GET`/`DELETE` for this entity anywhere, so Read is a documented no-op and Import/Delete
  both refuse explicitly rather than guess; the runtime counter is a version-gated write-only
  attribute, never an ordinary reconciled field — see
  `docs/guides/runtime-stores-and-number-ranges.md`. Variables, Data Stores, and Data Store
  Entries were evaluated and are deliberately unsupported/out of scope — no independent
  creation API, and/or the object is runtime business data, not desired state):
  - `sapintegrationsuite_number_range`
- A shared, CSRF-aware HTTP client: every modifying (POST/PUT/PATCH/DELETE) request now
  transparently fetches and retries with an `X-CSRF-Token` if SAP's API asks for one,
  independently of OAuth authentication — see `docs/sap-api-references.md`
- A machine-readable provider feature support catalog, queryable with no SAP tenant
  credentials — see `docs/feature-support.md`:
  - `data.sapintegrationsuite_provider_features`
  - `data.sapintegrationsuite_provider_feature`
- Import support and drift detection for every resource above
- Unit test suite (`httptest`-based) for auth, HTTP retry, OData v2, and domain mapping
- Acceptance test framework (gated on `TF_ACC=1`)
- Generated provider documentation, greenfield/brownfield examples
- CI (fmt/vet/test/build/lint) and GoReleaser-based release pipeline

## v0.2.x

- Value mapping entry-level management (`UpsertValMaps`, `UpdateDefaultValMap`,
  `DeleteValMaps`), once the exact payload/path shapes and delete granularity are confirmed
  against a reachable primary source or a live tenant — see `docs/resource-design.md`
  (**Research required**)
- Message mapping entry-level or dependent-resource management, if SAP ever exposes one
  independent of the opaque content archive this provider already transports (**Research
  required**)
- Current API Management / API Artifacts / Integration Cell (**Current**, priority 1 above) — see
  `docs/guides/current-api-management.md` for the confirmed public-API boundary and exactly which
  of the following are implemented, read-only, or blocked by a missing public API:
  - `sapintegrationsuite_api_artifact`
  - `sapintegrationsuite_api_artifact_deployment`
  - `data.sapintegrationsuite_runtime_profile(s)`
  - `data.sapintegrationsuite_integration_cell`
  - `sapintegrationsuite_integration_cell_virtual_host`
- Classic API Management resources, once the required scopes and object model are fully
  mapped (**Planned**, priority 3 above — explicitly behind current API Management, reflecting
  SAP's own current product direction)

## v0.3.x

- Edge Integration Cell control-plane configuration (registration, runtime association),
  strictly scoped to SAP-specific concerns — never a Kubernetes/Helm replacement (**Blocked by
  API**, priority 4 above)
- Additional Integration Suite capabilities, as their public APIs are confirmed (**Research
  required**, priority 5 above)

## Later

- Any further Integration-Suite capability for which SAP publishes a stable, documented
  public API, evaluated resource-by-resource against `docs/resource-design.md`'s suitability
  checklist before it is added.

## Explicitly not planned (Out of scope)

- Anything listed as "Out of Scope" in `docs/provider-scope.md`
- A generic `sapintegrationsuite_capability` resource, unless SAP publishes a public
  capability-activation API (none exists today — see
  `docs/provisioning-capability-matrix.md`)
- Terraform data sources for logs, metrics, or events (see `docs/architecture.md` and
  `docs/provider-scope.md` §66 — this provider is not a monitoring system)
