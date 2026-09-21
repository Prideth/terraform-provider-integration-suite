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
- Provider scope, boundary, architecture, and API discovery documentation.
