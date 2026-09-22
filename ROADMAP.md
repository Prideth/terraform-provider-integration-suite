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

1. **Security Content** — **Next.** User Credentials, OAuth Credentials, Certificates, Keystore
   material. Only resources where secrets and read-back semantics are safe for Terraform.
2. **Message Queues / Data Stores / Variables / Number Ranges** — Research required per object;
   do not treat operational/monitoring APIs automatically as Terraform resources.
3. **Classic API Management** — Planned, after Cloud Integration core support is mature.
4. **New API Gateway / API Artifacts** — Blocked by API: existence confirmed via UI/feature
   documentation, but no public design-time API confirmed in enough detail for a stable
   schema. API Artifacts, API Artifact Deployment, API Policies, Runtime Profiles.
5. **Integration Cell** — Blocked by API for SAP-side lifecycle/configuration; no public
   activation or status API found yet. Access policy replication/reconciliation to Integration
   Cell is a documented Integration Suite UI capability, but no public API surface for it was
   confirmed during the access-policy completion pass — see
   `docs/sap-api-references.md`.
6. **Edge Integration Cell** — Blocked by API for SAP control-plane/runtime-specific
   configuration; never a Kubernetes/Helm replacement. Same access-policy-replication caveat
   as Integration Cell above.
7. **Additional Integration Suite capabilities** — Research required: Integration Advisor,
   Trading Partner Management, Integration Assessment, Migration Assessment, and others, only
   where public APIs justify Terraform management.

Access Policies completion (role/identity semantics, artifact reference lifecycle audit,
runtime reconciliation research, minimal-PATCH update, data sources, import/drift) is done —
see "Implemented" below and `docs/guides/access-policies.md`. Partner Directory is also done —
see "Implemented" below and `docs/guides/partner-directory.md`.

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
- Security material resources (user credentials, OAuth2 client credentials, keystore
  entries, certificate-user mappings) with write-only/sensitive-value semantics (**Planned**,
  priority 1 above)
- Classic API Management resources, once the required scopes and object model are fully
  mapped (**Planned**, priority 3 above)
- New API Gateway / API Artifact model, once its public API surface is confirmed in enough
  detail for a stable schema (**Blocked by API**, priority 4 above):
  - `sapintegrationsuite_api_artifact`
  - `sapintegrationsuite_api_artifact_deployment`
- API Policies, if a typed, stable policy schema is achievable (**Blocked by API**)
- Integration Cell: read/status resources, if SAP publishes a public status API (activation
  itself stays a manual bootstrap step until SAP publishes one) (**Blocked by API**)

## v0.3.x

- Edge Integration Cell control-plane configuration (registration, runtime association),
  strictly scoped to SAP-specific concerns — never a Kubernetes/Helm replacement (**Blocked by
  API**)
- Additional Integration Suite capabilities, as their public APIs are confirmed (**Research
  required**)

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
