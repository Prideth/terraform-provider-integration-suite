# Changelog

All notable changes to this project are documented in this file.

## Unreleased

Initial development toward v0.1.0. See `ROADMAP.md` for what is planned and
`docs/` for the API discovery this release is based on.

### Added

- Provider foundation: OAuth 2.0 client credentials authentication with
  token caching, a retrying HTTP client, and an OData V2 request/pagination/
  error-handling layer.
- `sapintegrationsuite_integration_package` resource and data source.
- `sapintegrationsuite_integration_flow` resource (file-based content).
- `sapintegrationsuite_integration_flow_deployment` resource, with
  context-aware polling instead of fixed sleeps.
- `sapintegrationsuite_access_policy` and
  `sapintegrationsuite_access_policy_reference` resources.
- `sapintegrationsuite_value_mapping` resource and data source (file-based
  content), and `sapintegrationsuite_value_mapping_deployment`, sharing the
  runtime-artifact polling and status model already proven for integration
  flow deployments.
- Provider scope, boundary, architecture, and API discovery documentation.

### Known limitations

- `sapintegrationsuite_value_mapping`'s content update uses `PUT`, by
  analogy with `sapintegrationsuite_integration_flow`; SAP separately
  documents a `ValueMappingDesigntimeArtifactSaveAsVersion` action whose
  relationship to `PUT` has not been confirmed against a live tenant.
- Individual value mapping entries are not yet manageable through this
  provider — see `docs/resource-design.md`.
