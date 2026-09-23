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

1. **Edge Integration Cell** — SAP-side registration, runtime association, local API access, and
   deployment-targeting research, strictly scoped to control-plane concerns SAP itself exposes
   through a public API — never a Kubernetes/Helm replacement. See
   `docs/guides/edge-integration-cell.md`.
2. **Classic API Management** — API Providers, API Proxies, API Products, Key Value Maps, and the
   classic proxy policy model, through the API Portal's own `Management.svc` OData contract and
   its own service-key credentials — a distinct product model from the current, API-artifact
   centric API Management below, never mixed in this provider's terminology or code. Expected to
   be the largest remaining implementation phase. See `docs/guides/classic-api-management.md`.
3. **Integration Assessment** — Domain/Style/Use-Case/Integration Pattern and related objects,
   through Integration Assessment's own documented API package, evaluated per-object for
   Terraform suitability rather than implemented wholesale. See
   `docs/guides/integration-assessment.md`.
4. **Trading Partner Management** — company/subsidiary/trading-partner profiles, communication
   partner profiles, and agreement templates, kept conceptually and architecturally distinct from
   the already-implemented Partner Directory runtime configuration it can generate. See
   `docs/guides/trading-partner-management.md`.
5. **Integration Advisor** — Message Implementation Guidelines, Mapping Guidelines, type systems,
   B2B standards, and codelists, implemented only where a public API exposes persistent,
   independently identified design-time artifacts.
6. **Migration Assessment** — source-system, rule, effort, and assessment-result objects,
   expected to be mostly workflow/analysis-request data rather than desired-state configuration.
7. **Remaining Integration Suite capability audit** — a full sweep of every capability area this
   provider has not yet formally classified (Open Connectors, Event Mesh, Data Space Integration,
   and others), each resolved to a concrete provider-boundary decision.
8. **Provider completion and hardening** — a catalog-wide accuracy audit, README/documentation
   regeneration, and repository-wide quality/security review once the phases above have
   established this provider's practical ceiling of currently reachable public APIs.

Current API Management / API Artifacts / Integration Cell (API Artifacts, Runtime Profiles,
Integration Cell, Virtual Hosts, Policies, Reusable API Artifacts) research is done — see
`docs/guides/current-api-management.md`: SAP currently exposes no public API for any object in
this family, so this provider implements none of it; the finding is treated as settled rather
than repeatedly reopened, and is revisited only if SAP publishes something new (see the capability
audit phase above).

Developer Hub — SAP's API/Event/MCP Server catalog, publication, and subscription capability — is
not part of this provider's roadmap at all. It has its own API boundary, its own OAuth
credentials, and a consumer/catalog lifecycle unlike anything else this provider manages, and is
planned as a separate, independently versioned Terraform provider (working name
`Prideth/terraform-provider-sap-developer-hub`). See `docs/provider-scope.md` for the boundary
statement and `docs/feature-support.md`'s single `developer_hub` catalog entry.

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
- Edge Integration Cell control-plane configuration (registration, runtime association, local
  API access classification, deployment-targeting research), strictly scoped to SAP-specific
  concerns — never a Kubernetes/Helm replacement (priority 1 above)
- Classic API Management resources (API Providers, API Proxies, API Products, Key Value Maps),
  once the required scopes and object model are fully mapped (priority 2 above)

## v0.3.x

- Integration Assessment, Trading Partner Management, Integration Advisor, and Migration
  Assessment, each evaluated and implemented only where public APIs justify Terraform management
  (priorities 3 through 6 above)
- The remaining Integration Suite capability audit and provider hardening pass (priorities 7 and
  8 above)

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
