# SAP API References

Every implemented resource must trace back to an officially documented, SAP-supported public
API. This document is that trace.

> **Research note**: `help.sap.com`, `api.sap.com`, and `community.sap.com` were not directly
> reachable from the sandboxed environment this provider was initially developed in (blocked
> by network egress policy). Findings below were cross-checked through SAP's own
> `SAP-docs/btp-integration-suite` GitHub repository (the official Markdown source for the
> SAP Help Portal Integration Suite documentation) and SAP Business Accelerator Hub search
> results. Before relying on any exact field/entity name in production, verify it against the
> tenant's live `$metadata` document and the current SAP Business Accelerator Hub page —
> standard OData practice, and doubly important here given the access restriction noted
> above.

## `sapintegrationsuite_integration_package`

- **SAP product area**: Integration Suite / Cloud Integration
- **Official API**: Integration Content API (`CloudIntegrationAPI` package, SAP Business
  Accelerator Hub)
- **Entity set**: `IntegrationPackages`
- **Protocol**: OData V2
- **Operations**: GET, POST, PATCH, DELETE (PATCH limited to SAP-permitted metadata fields;
  used rather than PUT so fields outside the Terraform schema are not reset to their defaults)
- **Required roles**: `IntegrationOperationServer` / `IntegrationDeveloper` OAuth scopes (role
  collection names vary per tenant; assign the least-privileged Integration Suite role
  collection covering "Integration Content" design-time operations)

## `sapintegrationsuite_integration_flow` / `..._deployment`

- **SAP product area**: Integration Suite / Cloud Integration
- **Official API**: Integration Content API
- **Entity sets / actions**: `IntegrationDesigntimeArtifacts`, `IntegrationRuntimeArtifacts`,
  `DeployIntegrationDesigntimeArtifact` (action)
- **Protocol**: OData V2
- **Operations**: GET, POST (create), PUT (create a new design-time version of an existing
  flow), DELETE, plus the `Deploy` action (POST)
- **Required roles**: as above, plus deploy-specific scopes for the runtime artifact
  operations

## `sapintegrationsuite_value_mapping` / `..._deployment`

- **SAP product area**: Integration Suite / Cloud Integration
- **Official API**: Integration Content API
- **Entity sets / actions**: `ValueMappingDesigntimeArtifacts`,
  `ValueMappingDesigntimeArtifactSaveAsVersion` (action, not currently used — see Update below),
  `IntegrationRuntimeArtifacts` (the same shared runtime-artifacts entity
  `sapintegrationsuite_integration_flow_deployment` uses — confirmed via SAP's own "Runtime
  Status API" description as covering "currently deployed integration artifacts" generally,
  not one entity per design-time artifact type), `DeployValueMappingDesigntimeArtifact` (action;
  singular form re-checked this phase — see `docs/resource-design.md`)
- **Protocol**: OData V2
- **Operations**: GET, POST (create), DELETE, plus the `Deploy` action (POST). No PUT/Update —
  see below.
- **Required roles**: `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`,
  `WorkspaceArtifactsDeploy`
- **Confirmed constraint**: a value mapping cannot be created with zero entries — at least one
  mapping entry must be part of the artifact's content at creation time.
- **Update — resolved conservatively, no in-place update implemented**: SAP documents a distinct
  `ValueMappingDesigntimeArtifactSaveAsVersion` action (POST, taking the artifact's technical ID
  and a caller-supplied new version identifier). An earlier version of this provider called
  `PUT` against the keyed entity instead, by analogy with `IntegrationDesigntimeArtifacts`.
  Re-investigating this contract, `help.sap.com`, `api.sap.com`, `community.sap.com`,
  `blogs.sap.com`, and every reachable mirror/proxy for them were blocked by this environment's
  network egress policy, so the exact `PUT` vs. `SaveAsVersion` semantics could not be confirmed
  against primary documentation or an actual request/response trace. Reachable secondary
  evidence — a dedicated SAP Knowledge Base Article (3502529) treating "changing the version of
  a ValueMapping" as its own distinct, separately gated operation, and an independent
  third-party OData client that explicitly disables generic update for
  `ValueMappingDesigntimeArtifacts` while leaving it enabled for the sibling
  `IntegrationDesigntimeArtifacts`/`MessageMappingDesigntimeArtifacts`/`ScriptCollectionDesigntimeArtifacts`
  entity sets — points away from a working generic `PUT` for this entity set specifically.
  Given that, `sapintegrationsuite_value_mapping` does not retain the `PUT` call: `name`,
  `content`, and `content_hash` are all `RequiresReplace`, so any change to them replaces the
  resource (`Create` a new artifact, `Delete` the old one) instead of relying on an unverified
  update path. `UpdateValueMapping` no longer exists in
  `internal/client/cloudintegration/value_mapping.go`. Implementing true in-place update via
  `ValueMappingDesigntimeArtifactSaveAsVersion` is deferred to v0.2.x, once its request/response
  contract can be confirmed against a live tenant or a reachable primary source; see
  `docs/resource-design.md` for the full reasoning.
- **Delete semantics — unverified scope**: `DeleteValueMapping` deletes via the same
  `(Id, Version='active')` key used for reads. Whether this removes only the active version or
  every version of the artifact was not confirmed against a primary source (same network
  restrictions as above). This provider never creates more than one version of a value mapping
  concurrently, so it does not change this resource's observable behavior, but it means
  `terraform destroy` is not guaranteed to remove every version SAP stored — flagged here rather
  than asserted as "deletes all versions".
- **Deferred — entry-level operations**: `UpsertValMaps` (POST, insert/update individual
  mapping rows — confirmed to 404 if the target source/target agency-identifier scheme does not
  already exist), `UpdateDefaultValMap` (POST, sets a scheme's default value via a `ValMapId`
  GUID obtained from a separate lookup), and `DeleteValMaps` (exact deletion granularity
  unconfirmed) are real, existing APIs that this phase does not implement — see
  `docs/resource-design.md` for why guessing their wire format was rejected in favor of
  documenting the gap.

## `sapintegrationsuite_message_mapping` / `..._deployment`

- **SAP product area**: Integration Suite / Cloud Integration
- **Official API**: Integration Content API
- **Scope note**: this is the reusable, package-level Message Mapping *artifact*
  (`MessageMappingDesigntimeArtifacts`), not the inline/local message mapping step configurable
  directly inside an integration flow. See `docs/resource-design.md` for the distinction.
- **Entity sets / actions**: `MessageMappingDesigntimeArtifacts`,
  `MessageMappingDesigntimeArtifactSaveAsVersion` (action, confirmed to exist, not used by this
  provider — see Update below), `IntegrationRuntimeArtifacts` (the same shared runtime-artifacts
  entity `sapintegrationsuite_integration_flow_deployment` and
  `sapintegrationsuite_value_mapping_deployment` use), `DeployMessageMappingDesigntimeArtifact`
  (action; singular form, matching the sibling actions for the other design-time artifact types)
- **Protocol**: OData V2
- **Operations**: GET, POST (create), PUT (update — see below), DELETE, plus the `Deploy` action
  (POST)
- **Required roles**: `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`,
  `WorkspaceArtifactsDeploy`
- **Sources used**: SAP Help Portal content read via the `SAP-docs` GitHub organization's
  markdown mirror of the official Cloud Integration documentation (a legitimate primary source —
  SAP's own published documentation, mirrored verbatim for community feedback — reached this
  phase despite `help.sap.com` itself being blocked by this environment's network egress
  policy), plus an independent third-party OData client
  (`github.com/lemaiwo/ci-mcp-server`) built directly against this same API, used as
  corroborating (not primary) evidence.
- **Content format — confirmed, not inferred from Integration Flow**: a message mapping
  artifact's content is a mapping definition (`*.mmap`) file; SAP's own artifact-upload UI
  accepts it as (or bundled inside) a ZIP archive. This provider transports that ZIP opaquely as
  base64-encoded `ArtifactContent`, the same as `IntegrationDesigntimeArtifacts` and
  `ValueMappingDesigntimeArtifacts` — it does not parse the `.mmap` file or any XSD/WSDL/EDMX/
  Swagger-OpenAPI schema files the mapping may reference for source/target message structures.
- **Update — resolved as `PUT`, on entity-specific grounds**: unlike
  `sapintegrationsuite_value_mapping` (which has no in-place update — see that resource's entry
  below), this phase found positive evidence supporting `PUT` specifically for
  `MessageMappingDesigntimeArtifacts`: it shares `IntegrationDesigntimeArtifacts`' exact
  `(Id, Version)` key shape and confirmed version-creating `PUT` behavior; the same third-party
  OData client that explicitly disables generic update for `ValueMappingDesigntimeArtifacts`
  explicitly *enables* it for `MessageMappingDesigntimeArtifacts` (matching
  `IntegrationDesigntimeArtifacts` and `ScriptCollectionDesigntimeArtifacts`); and no SAP KBA or
  other evidence of a documented `PUT` problem for this entity set was found (unlike Value
  Mapping's KBA 3502529). `MessageMappingDesigntimeArtifactSaveAsVersion` exists here too, but
  this phase's research clarified that `SaveAsVersion` is a universal action across this whole
  API family (confirmed to exist for `IntegrationDesigntimeArtifacts` as well, coexisting with
  its confirmed `PUT`), not evidence against `PUT` by itself — see `docs/resource-design.md` for
  the full reasoning. This provider does not use `SaveAsVersion` since it does not ask users to
  manage an explicit version string.
- **Delete — unverified scope, same open item as Value Mapping**: `DeleteMessageMapping` deletes
  via `(Id, Version='active')`, the same key used for reads. Whether this removes only the
  active version or every version of the artifact was not confirmed against a primary source.
- **Deployment — `IntegrationRuntimeArtifacts`, not `BuildAndDeployStatus`**: investigated both
  per this phase's instructions. `BuildAndDeployStatus(TaskId='…')` is documented in the context
  of a different artifact family's build-then-deploy pipeline (OData API artifacts, keyed by a
  `TaskId` a build operation returns), not `MessageMappingDesigntimeArtifacts`. SAP's Runtime
  Status API documentation and secondary sources both describe message mappings as deployed
  runtime artifacts monitored through the same shared `IntegrationRuntimeArtifacts` entity
  already used by the other `*_deployment` resources; `runtime_artifact.go` and
  `runtime_deployment.go` are reused unmodified. See `docs/resource-design.md` for the full
  reasoning.
- **No hidden coupling to referencing integration flows**: SAP's own documentation confirms an
  integration flow's deployment does not automatically deploy a message mapping it references;
  this provider does not add automatic-deployment behavior SAP itself does not provide, and does
  not scan or modify integration flow content to manage that reference.

## `sapintegrationsuite_access_policy` / `..._reference`

- **SAP product area**: Integration Suite / Security
- **Official API**: Security Content API (documented by SAP as covering "keystore entries,
  user credentials, certificate-to-user mappings", and — per SAP's own access-policy
  documentation — access policies themselves, described as retrievable "by an OData V2 API"
  with both read and write operations)
- **Entity set**: `AccessPolicies`, with nested artifact references
- **Protocol**: OData V2
- **Operations**: GET, POST, PATCH, DELETE
- **Supported artifact reference types**: confirmed against SAP KBA 3447540 ("Integration
  Package" artifact type is explicitly *not* available when maintaining an access policy) and
  a community post enumerating the supported set, which matches this provider's
  `SupportedArtifactTypes` list exactly: `IntegrationFlow`, `ODataAPI`, `RestAPI`, `SoapAPI`,
  `ScriptCollection`, `ValueMapping`, `MessageMapping`, `MessageQueue`, `GlobalDataStore`,
  `GlobalVariable`. The human-readable UI labels are confirmed (e.g. "REST API"); the exact
  casing/spelling of the wire-format enum values (`RestAPI` vs. `REST_API` vs. something else)
  is this provider's best inference from OData naming conventions elsewhere in the same API
  and still needs verification against a live tenant's `$metadata` or an actual create
  response.
- **Role association caveat**: SAP's own documentation on managing access policies states that
  the role granting access to the artifacts an access policy protects is associated "using SAP
  Business Technology Platform cockpit" — i.e. through a BTP role collection, which is
  out of this provider's scope (see `docs/provider-scope.md`). It is not yet confirmed whether
  the `AccessPolicies` entity's `RoleName` field is simply a label referenced by that
  BTP-side role collection, or carries additional semantics on the Integration Suite side.
  Until confirmed, treat `role_name` as write-once-at-creation and avoid relying on renaming
  it having any particular effect.
- **Required roles**: Integration Suite "Manage Security" / access-policy administration
  scopes

## Deferred APIs (tracked, not yet implemented)

| Area | API | Status |
|---|---|---|
| Classic API Management | "Accessing API Management APIs Programmatically" REST/OData APIs | Confirmed public, deferred to a later minor version |
| New API Gateway / API Artifacts | Design-time API for API-centric artifacts with Runtime Profile (Integration Cell / Edge Integration Cell) | Existence confirmed via UI/feature docs; exact public API surface not yet confirmed in enough detail for a stable Terraform schema — deferred to v0.2.x |
| Partner Directory | Partner Directory API (part of the same `CloudIntegrationAPI` package) | Confirmed public, deferred — not yet schema-designed |
| Security material (user credentials, OAuth2 client credentials, keystore entries) | Security Content API | Confirmed public, deferred — needs write-only/sensitive-value design pass first |
| Value mapping entry-level management | `UpsertValMaps`, `UpdateDefaultValMap`, `DeleteValMaps` | Confirmed public, deferred — exact payload/path shapes and delete granularity not confirmed against a reachable primary source; see `docs/resource-design.md` |

## Explicitly ruled out

- **Integration Suite capability activation** (Cloud Integration, API Management, Integration
  Cell, Edge Integration Cell): no public activation API found. SAP's own documentation
  describes activation as a UI action ("Settings" → "Runtime" → *Activate*). Not implemented;
  see `docs/provisioning-capability-matrix.md`.
- Any endpoint only reachable by reverse-engineering the Integration Suite UI's network
  traffic. None were used, and none will be, regardless of how convenient they would be.
