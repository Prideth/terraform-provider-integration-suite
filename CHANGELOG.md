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

### Changed

- `sapintegrationsuite_value_mapping` no longer implements Update via
  `PUT`. Re-verifying the Value Mapping API contract found no confirmed
  in-place update path for this entity set (unlike
  `sapintegrationsuite_integration_flow`'s equivalent, which is confirmed);
  `name`, `content`, and `content_hash` are now `RequiresReplace`, so
  changing any of them replaces the resource instead of relying on an
  unverified `PUT`. See `docs/sap-api-references.md` for the full
  reasoning and `docs/resource-design.md` for what was checked.

### Known limitations

- `sapintegrationsuite_value_mapping` has no in-place update (see Changed
  above); SAP separately documents a `ValueMappingDesigntimeArtifactSaveAsVersion`
  action this provider does not yet use, deferred to v0.2.x pending
  confirmation of its exact contract.
- Whether `sapintegrationsuite_value_mapping`'s Delete removes only the
  active version or every version of the artifact has not been confirmed
  against a primary source.
- Individual value mapping entries are not yet manageable through this
  provider — see `docs/resource-design.md`.
