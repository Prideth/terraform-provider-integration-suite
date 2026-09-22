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
  `ValueMappingDesigntimeArtifactSaveAsVersion` (action, not currently used — see the open
  question below), `IntegrationRuntimeArtifacts` (the same shared runtime-artifacts entity
  `sapintegrationsuite_integration_flow_deployment` uses — confirmed via SAP's own "Runtime
  Status API" description as covering "currently deployed integration artifacts" generally,
  not one entity per design-time artifact type), `DeployValueMappingDesigntimeArtifact` (action)
- **Protocol**: OData V2
- **Operations**: GET, POST (create), PUT (this provider's Update implementation — see the open
  question below), DELETE, plus the `Deploy` action (POST)
- **Required roles**: `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`,
  `WorkspaceArtifactsDeploy`
- **Confirmed constraint**: a value mapping cannot be created with zero entries — at least one
  mapping entry must be part of the artifact's content at creation time.
- **Open question — Update semantics**: SAP documents a distinct
  `ValueMappingDesigntimeArtifactSaveAsVersion` action (POST, taking the artifact's technical
  ID and a caller-supplied new version identifier) alongside plain `POST`/`PUT`. Whether normal
  content updates should go through `PUT` (as implemented, by analogy with
  `IntegrationDesigntimeArtifacts`) or through this action instead has not been confirmed:
  `help.sap.com`, `api.sap.com`, `community.sap.com`, and `blogs.sap.com` were all unreachable
  from this development environment, and no other reachable source gave the exact
  request/response shape needed to decide with confidence. Verify against a live tenant before
  relying on `Update` in production; see `docs/resource-design.md` for the full reasoning.
- **Deferred — entry-level operations**: `UpsertValMaps` (POST, insert/update individual
  mapping rows — confirmed to 404 if the target source/target agency-identifier scheme does not
  already exist), `UpdateDefaultValMap` (POST, sets a scheme's default value via a `ValMapId`
  GUID obtained from a separate lookup), and `DeleteValMaps` (exact deletion granularity
  unconfirmed) are real, existing APIs that this phase does not implement — see
  `docs/resource-design.md` for why guessing their wire format was rejected in favor of
  documenting the gap.

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
