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
  `sapintegrationsuite_access_policy_reference` resources, plus matching
  `data.sapintegrationsuite_access_policy` and
  `data.sapintegrationsuite_access_policy_reference` data sources. See
  `docs/guides/access-policies.md` for role/BTP semantics, supported
  artifact types/attributes/operators, and runtime reconciliation findings.
- `sapintegrationsuite_value_mapping` resource and data source (file-based
  content), and `sapintegrationsuite_value_mapping_deployment`, sharing the
  runtime-artifact polling and status model already proven for integration
  flow deployments.
- `sapintegrationsuite_message_mapping` resource and data source (file-based
  content), and `sapintegrationsuite_message_mapping_deployment`, for the
  reusable, package-level message mapping artifact — not the inline/local
  message mapping step an integration flow can also define directly inside
  its own content. Unlike `sapintegrationsuite_value_mapping`, this
  resource has a confirmed in-place Update via `PUT`, on entity-specific
  evidence documented in `docs/sap-api-references.md`. Reuses the same
  shared runtime-artifact polling and status model, confirmed applicable to
  this entity type rather than assumed.
- `sapintegrationsuite_script_collection` resource and data source
  (file-based content), and `sapintegrationsuite_script_collection_deployment`,
  for reusable Groovy/JavaScript script bundles. Shares its Update model
  with `sapintegrationsuite_message_mapping` (a confirmed in-place `PUT`,
  on the same entity-specific evidence) and reuses the same shared
  runtime-artifact polling and status model as every other `*_deployment`
  resource.
- Partner Directory support: `sapintegrationsuite_partner_string_parameter`,
  `sapintegrationsuite_partner_binary_parameter` (file-based, with SAP's
  documented 260 KB size limit checked before upload),
  `sapintegrationsuite_alternative_partner` (hiding SAP's hex-encoded
  `Hexagency`/`Hexscheme`/`Hexid` entity key behind plain
  agency/scheme/external_id attributes), and
  `sapintegrationsuite_partner_authorized_user` resources, each with a
  matching data source, plus `data.sapintegrationsuite_partner` and
  `data.sapintegrationsuite_partners` for discovery (there is no
  `sapintegrationsuite_partner` resource: SAP documents no confirmed create
  operation for Partners, and deleting one is documented as cascading to
  every entity belonging to it). See `docs/guides/partner-directory.md`.
- `sapintegrationsuite_partner_user_credential_parameter`, a
  security-sensitive Partner Directory resource using a write-only
  `password_wo` attribute (Terraform CLI 1.11+) paired with a
  `password_wo_version` marker, since Terraform never stores the password
  and this provider never reads one back from SAP.
- CSRF token handling in the shared HTTP client for every modifying
  (POST/PUT/PATCH/DELETE) request: SAP's OData V2 services protect writes
  with an `X-CSRF-Token` independently of OAuth, and this had no handling
  anywhere in this provider before now. Fetches and retries transparently
  when SAP asks for a token; a no-op when it does not.
- Generic server-driven paging (`__next` link following) in the OData v2
  client, so a large collection (Partner Directory's String Parameters in
  particular) is read completely rather than silently truncated to its
  first page.
- A machine-readable provider feature support catalog
  (`internal/features`), queryable via `data.sapintegrationsuite_provider_features`
  and `data.sapintegrationsuite_provider_feature` with no SAP host or
  OAuth credentials required — this provider's `Configure` no longer fails
  just because SAP connectivity is unconfigured; every SAP-backed resource
  and data source instead reports a clear, specific error when actually
  used without one. See `docs/feature-support.md`.
- Provider scope, boundary, architecture, and API discovery documentation.
- `sapintegrationsuite_user_credential` and `sapintegrationsuite_oauth2_client_credential`
  resources (plus matching data sources) for SAP's Security Content API, the first
  Security Content credential artifacts this provider manages. Both use write-only
  `password_wo`/`client_secret_wo` attributes paired with a `_wo_version` marker (the
  same pattern as `sapintegrationsuite_partner_user_credential_parameter`), but unlike
  that resource, both have a confirmed in-place Update via `PUT` (SAP documents an
  "Edit and redeploy" action for Credentials artifacts), so rotating a secret redeploys
  the credential rather than replacing the resource. A dedicated
  `internal/client/securitycontent` client package backs both. See
  `docs/guides/security-content.md` for the full security model, which fields could and
  could not be confirmed, and why most other Security Content artifact types (keystore
  entries, certificates, key pairs, SSH keys, certificate chains, secure parameters,
  known hosts) are not implemented yet.
- `data.sapintegrationsuite_service_endpoints`, a read-only discovery data source for
  SAP's `ServiceEndpoints` API: the runtime entry point URLs and API definition links
  SAP generates for deployed Cloud Integration content. Supports the documented `name`/
  `protocol` filters, combines `EntryPoints`/`ApiDefinitions` expansion into a single
  request, fetches every page via server-driven paging, and sorts its result
  deterministically since SAP does not document a guaranteed response order. There is
  deliberately no matching resource (SAP offers no create/update/delete API for these)
  and no singular per-endpoint lookup (uniqueness of `name` is not confirmed) — see
  `docs/guides/service-endpoints.md`.
- A generated "Feature Support" dashboard in `README.md`, between
  `<!-- BEGIN GENERATED FEATURE SUPPORT -->`/`<!-- END GENERATED FEATURE SUPPORT -->`
  markers, produced by `go run ./cmd/gendocs -readme` (also run by `make docs`) from the
  same `internal/features/catalog.go` that already generates `docs/feature-support.md`,
  so the two can never drift apart into independently maintained copies. Grouped by
  domain, using the ✅/⚠️/👁️/🧪/❌ icon legend, and — unlike the README table it
  replaces — shows unsupported and out-of-scope features alongside supported ones, not
  just a curated list of what works.

### Changed

- `sapintegrationsuite_value_mapping` no longer implements Update via
  `PUT`. Re-verifying the Value Mapping API contract found no confirmed
  in-place update path for this entity set (unlike
  `sapintegrationsuite_integration_flow`'s equivalent, which is confirmed);
  `name`, `content`, and `content_hash` are now `RequiresReplace`, so
  changing any of them replaces the resource instead of relying on an
  unverified `PUT`. See `docs/sap-api-references.md` for the full
  reasoning and `docs/resource-design.md` for what was checked.
- `sapintegrationsuite_access_policy`'s Update now sends a PATCH payload
  containing only `Description`, instead of resending the immutable
  `RoleName` unchanged on every description update.
- The `security.certificate_user_mapping` feature catalog entry is corrected from
  `public_api: true` / `not_implemented` to `public_api: false` / `no_public_api`:
  reverifying it found SAP's certificate-to-user mapping documentation exists only for
  the Neo environment, with no Cloud Foundry equivalent, and this provider targets
  Cloud Foundry.

### Known limitations

- `sapintegrationsuite_value_mapping` has no in-place update (see Changed
  above); SAP separately documents a `ValueMappingDesigntimeArtifactSaveAsVersion`
  action this provider does not yet use, deferred to v0.2.x pending
  confirmation of its exact contract.
- Whether `sapintegrationsuite_value_mapping`'s,
  `sapintegrationsuite_message_mapping`'s, or
  `sapintegrationsuite_script_collection`'s Delete removes only the active
  version or every version of the artifact has not been confirmed against
  a primary source.
- Individual value mapping entries are not yet manageable through this
  provider — see `docs/resource-design.md`.
- `sapintegrationsuite_access_policy`'s `reconciliation_status` is
  best-effort and not polled to a terminal state: SAP's documentation
  confirms access policies can be replicated to the Cloud Integration
  runtime, Integration Cell, and Edge Integration Cell with a per-runtime
  `Fail`/`Success`/`Pending` reconciliation status, but this project could
  not confirm that mechanism is exposed through the public `AccessPolicies`
  OData API as opposed to being UI-only. See
  `docs/guides/access-policies.md`.
- The exact wire-format casing SAP's `AccessPolicies` OData API expects for
  `Attribute` (`Name`/`Id`) and `Operator` (`EQUALS`/`MATCHES`) enum values
  has not been confirmed against a live tenant or `$metadata`.
- `sapintegrationsuite_partner_user_credential_parameter` has no in-place
  update and no password read-back — a permanent property of its security
  model, not a gap expected to close later. See
  `docs/guides/partner-directory.md`.
- Whether `sapintegrationsuite_partner_authorized_user`'s `user` value is
  case-normalized by SAP internally has not been confirmed against a
  primary source; this provider does not normalize it.
- There is no `sapintegrationsuite_partner` resource. SAP documents no
  confirmed create operation for `Partners`, and deleting one is
  documented as capable of cascading to every entity that belongs to it —
  see `docs/guides/partner-directory.md`.
- `sapintegrationsuite_user_credential` and `sapintegrationsuite_oauth2_client_credential`
  never read a password/client secret back from SAP — a permanent property of their
  security model. `sapintegrationsuite_oauth2_client_credential` also only exposes name,
  description, token service URL, client ID, client secret, and scope; grant type
  placement, client authentication mode, resource, audience, and custom parameters are
  documented by SAP but not yet implemented. See `docs/guides/security-content.md`.
- `data.sapintegrationsuite_service_endpoints`'s `ApiDefinitions[].url` JSON property
  casing is inferred by consistency with the independently confirmed `EntryPoints[].url`
  casing (confirmed from SAP's own open-source Piper library), not independently
  confirmed itself. Whether a fresh deployment's service endpoint appears immediately or
  after a propagation delay is also unconfirmed — see `docs/guides/service-endpoints.md`.
