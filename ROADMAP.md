# Roadmap

This roadmap is adjusted after each round of API discovery documented under `docs/`. It
favors a small, high-quality resource set over broad but shallow coverage.

## v0.1.x — Provider Foundation

- Provider skeleton on the Terraform Plugin Framework, named `sapintegrationsuite`
- OAuth 2.0 client credentials authentication with a thread-safe, context-aware token cache
- Shared HTTP client: retry with backoff/jitter, `Retry-After` handling, bounded response
  sizes, custom User-Agent
- OData V2 request/query/pagination/error-parsing foundation
- Cloud Integration fundamentals:
  - `sapintegrationsuite_integration_package`
  - `sapintegrationsuite_integration_flow`
  - `sapintegrationsuite_integration_flow_deployment`
- Access Policies:
  - `sapintegrationsuite_access_policy`
  - `sapintegrationsuite_access_policy_reference`
- Value Mappings:
  - `sapintegrationsuite_value_mapping`
  - `sapintegrationsuite_value_mapping_deployment`
- Import support and drift detection for every resource above
- Unit test suite (`httptest`-based) for auth, HTTP retry, OData v2, and domain mapping
- Acceptance test framework (gated on `TF_ACC=1`)
- Generated provider documentation, greenfield/brownfield examples
- CI (fmt/vet/test/build/lint) and GoReleaser-based release pipeline

## v0.2.x

- Value mapping entry-level management (`UpsertValMaps`, `UpdateDefaultValMap`,
  `DeleteValMaps`), once the exact payload/path shapes and delete granularity are confirmed
  against a reachable primary source or a live tenant — see `docs/resource-design.md`
- Script collections, message mappings (same design-time/runtime split as integration flows
  and value mappings)
- Classic API Management resources, once the required scopes and object model are fully
  mapped
- New API Gateway / API Artifact model, once its public API surface is confirmed in enough
  detail for a stable schema:
  - `sapintegrationsuite_api_artifact`
  - `sapintegrationsuite_api_artifact_deployment`
- API Policies, if a typed, stable policy schema is achievable
- Integration Cell: read/status resources, if SAP publishes a public status API (activation
  itself stays a manual bootstrap step until SAP publishes one)

## v0.3.x

- Edge Integration Cell control-plane configuration (registration, runtime association),
  strictly scoped to SAP-specific concerns — never a Kubernetes/Helm replacement
- Security material resources (user credentials, OAuth2 client credentials, keystore
  entries, certificate-user mappings) with write-only/sensitive-value semantics
- Partner Directory resources
- Additional Integration Suite capabilities, as their public APIs are confirmed

## Later

- Any further Integration-Suite capability for which SAP publishes a stable, documented
  public API, evaluated resource-by-resource against `docs/resource-design.md`'s suitability
  checklist before it is added.

## Explicitly not planned

- Anything listed as "Out of Scope" in `docs/provider-scope.md`
- A generic `sapintegrationsuite_capability` resource, unless SAP publishes a public
  capability-activation API (none exists today — see
  `docs/provisioning-capability-matrix.md`)
- Terraform data sources for logs, metrics, or events (see `docs/architecture.md` and
  `docs/provider-scope.md` §66 — this provider is not a monitoring system)
