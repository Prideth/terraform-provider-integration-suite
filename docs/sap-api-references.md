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

## `sapintegrationsuite_script_collection` / `..._deployment`

- **SAP product area**: Integration Suite / Cloud Integration
- **Official API**: Integration Content API
- **Entity sets / actions**: `ScriptCollectionDesigntimeArtifacts`,
  `IntegrationRuntimeArtifacts` (the same shared runtime-artifacts entity every other
  `*_deployment` resource in this provider uses), `DeployScriptCollectionDesigntimeArtifact`
  (action; singular form, matching every other design-time artifact type's deploy action)
- **Protocol**: OData V2
- **Operations**: GET, POST (create), PUT (update), DELETE, plus the `Deploy` action (POST)
- **Required roles**: `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`,
  `WorkspaceArtifactsDeploy`
- **Sources used**: SAP Help Portal content read via the `SAP-docs` GitHub organization's
  markdown mirror (the same primary-source path used for Message Mapping), plus the same
  independent third-party OData client (`github.com/lemaiwo/ci-mcp-server`) used as
  corroborating evidence for Message Mapping's Update model.
- **Documented constraints**: a script collection's technical ID must be unique across the
  entire tenant (not just the containing package), and its description is capped at 120
  characters — both confirmed via SAP's own documentation, neither enforced client-side by this
  provider.
- **Content format**: a ZIP archive of Groovy/JavaScript script files, transported opaquely as
  base64-encoded `ArtifactContent`, the same convention as every other file-based design-time
  resource in this provider. This provider does not parse, validate, or execute the scripts.
- **Update — resolved as `PUT`, same grounds as Message Mapping**: `ScriptCollectionDesigntimeArtifacts`
  shares `IntegrationDesigntimeArtifacts`' `(Id, Version)` key shape and confirmed `PUT`
  behavior; the same third-party OData client that disables generic update for
  `ValueMappingDesigntimeArtifacts` explicitly enables it here too, matching
  `IntegrationDesigntimeArtifacts` and `MessageMappingDesigntimeArtifacts`; no SAP KBA or other
  evidence of a documented `PUT` problem for this entity set was found. See
  `docs/resource-design.md` for the full reasoning.
- **Delete — unverified scope**: same open item as every other design-time artifact type in
  this provider — whether Delete removes only the active version or every version is
  unconfirmed against a primary source.
- **Deployment**: fire-and-poll through the shared `IntegrationRuntimeArtifacts` entity, the
  same model already confirmed for integration flows, value mappings, and message mappings —
  script collections are documented as existing as runtime artifacts inside the same deployed
  runtime packages.
- **No hidden coupling to referencing integration flows**: this provider does not scan or
  modify integration flow content to manage a reference to a script collection, the same
  ownership boundary already established for message mapping.

## `sapintegrationsuite_integration_adapter` / `..._deployment`

- **SAP product area**: Integration Suite / Cloud Integration — custom Integration Adapters
  built with the SAP Adapter SDK. **Cloud Foundry environment only**; every SAP source covering
  this feature repeats "This information is relevant only when you use SAP Cloud Integration in
  the Cloud Foundry environment."
- **Official API**: Integration Content API. The general "Integration Content" resource table
  (read via the `SAP-docs/btp-integration-suite` GitHub mirror) lists it as: "Integration
  Adapter: Represents an integration adapter (only available in the Cloud Foundry environment).
  You can use resource `IntegrationAdapterDesigntimeArtifacts` to import, deploy, or delete an
  integration adapter."
- **Entity set**: `IntegrationAdapterDesigntimeArtifacts`
- **Protocol**: OData V2
- **Evidence tier — weaker than the sibling design-time artifact types above**: SAP's own
  "Integration Adapter Example Requests, Cloud Foundry Environment" page — the direct
  counterpart to the complete example-request pages that back Integration Flow, Value Mapping,
  Message Mapping, and Script Collection above — shows only two operations. Everything else
  below is explicitly marked by its evidence tier.
- **Confirmed — Delete**: `DELETE /api/v1/IntegrationAdapterDesigntimeArtifacts(Id='SubsystemSymbolicName1')`.
  This is the only confirmed evidence for the entity's key shape: `Id` alone, not the composite
  `(Id, Version)` key every sibling design-time artifact entity set in this API uses.
- **Confirmed — Deploy**: `POST /api/v1/DeployIntegrationAdapterDesigntimeArtifact?Id='SubsystemSymbolicName1'`.
  Two details worth flagging explicitly since they're easy to get wrong by analogy:
  - Singular action name ("...Artifact"), matching every sibling deploy action in this API —
    confirmed directly, not assumed.
  - No `Version` query parameter, unlike every sibling deploy action (which all take both `Id`
    and `Version`). Consistent with the Id-only key finding above.
- **Not confirmed by an adapter-specific example — Create**: implemented as `POST
  IntegrationAdapterDesigntimeArtifacts` with a JSON body of `PackageId`, `Id`, `Name`, `Type`,
  `Application`, and base64 `ArtifactContent`, matching every sibling design-time artifact
  type's confirmed Create shape in this exact API and corroborated by a third-party technical
  walkthrough describing this exact request for this exact entity — but not by an SAP-published
  example request the way Create is confirmed for every sibling type.
- **Confirmed (UI documentation) — identity and duplicate handling**: "The integration adapter
  ID needs to be unique across the tenant" (not merely the package) and "If there's already an
  integration adapter with the same ID, the system throws an error." The latter is treated as
  positive evidence against a working reimport-to-update flow — see Update below.
- **Not confirmed — Read**: `GET IntegrationAdapterDesigntimeArtifacts(Id='...')`, the ordinary
  OData GET-by-key convention every entity set in this API follows, not confirmed by an
  adapter-specific example.
- **Not confirmed, and deliberately not implemented — Update**: no PUT/PATCH/reimport example
  was found anywhere for this entity. Combined with the confirmed duplicate-ID-is-an-error
  behavior, this provider implements no update path at all: every attribute on
  `sapintegrationsuite_integration_adapter` is `RequiresReplace`.
- **`Type`/`Application` — confirmed as UI concepts, not confirmed as an enum or free text**:
  SAP's UI documentation states `Type` "is used to categorize the adapters based on line of
  business" (documented examples: Analytics, CRM, ERP, Finance, HCM, Marketing) and
  `Application` "refers to the software/application for which the adapter provides ...
  connectivity" (documented example: Slack), but never states whether either is a closed API
  enum. No validator is applied to either attribute.
- **Not confirmed — deployment runtime status / undeploy**: `sapintegrationsuite_integration_adapter_deployment`
  reuses the same shared `IntegrationRuntimeArtifacts` polling/undeploy (`GetRuntimeArtifact`/
  `UndeployRuntimeArtifact`, already used by every other `*_deployment` resource in this
  provider) by analogy. SAP's documentation for the shared `IntegrationRuntimeArtifacts` deploy
  mechanism explicitly states "You can only deploy BUNDLE type integration artifacts
  (integration flows, value mappings, or OData services)" — confirming adapters are *excluded*
  from that generic deploy path (which is exactly why they have their own dedicated deploy
  action) but not confirming or ruling out whether a custom adapter, once deployed through its
  own action, becomes readable/undeployable through that same shared entity. A dedicated
  `BuildAndDeployStatus`-based mechanism remains a plausible alternative this project could not
  rule out. See `docs/guides/integration-adapters.md`.
- **Distinct lifecycles — confirmed as separate concepts, addressed explicitly to prevent
  conflation**: "Import Integration Adapters" (a different SAP Help Portal page) describes
  importing a *prebundled SAP Business Accelerator Hub* adapter from inside the integration flow
  editor — an entirely different, UI-triggered, auto-deploying flow with no separate OData
  identity, not the custom-`.esa`-upload feature this provider manages. `docs/guides/integration-adapters.md`
  documents this distinction prominently, per this project's standing policy of never silently
  conflating two different SAP lifecycles that happen to share a name.
- **File size limit**: no SAP-documented maximum `.esa` upload size was found. This provider
  enforces only its own generic protective bound (32 MiB, the same bound already used for script
  collections and value mappings), not an SAP-documented limit.
- **Required roles**: SAP's "Importing Custom Integration Adapter, Cloud Foundry Environment"
  page names `WorkspacePackagesEdit` and `WorkspaceArtifactsDeploy` as the role templates
  required for the various adapter tasks.
- **CSRF audit performed for this feature, no changes needed**: this feature's Create (`POST`),
  Delete (`DELETE`), and Deploy (`POST`) client methods go through the exact same shared
  transport chain (`internal/client/cloudintegration.Client` → `internal/client/odata/v2.Client`
  → `internal/client/http.Client.Do`) as every other resource in this provider — the CSRF layer
  (`internal/client/http/csrf.go`) dispatches purely on HTTP method
  (`POST`/`PUT`/`PATCH`/`DELETE`), so it applies automatically with zero adapter-specific code.
  No new or adapter-specific CSRF handling was written or was needed; the full test suite
  (`internal/client/auth`, `internal/client/cloudintegration`, `internal/client/http`,
  `internal/client/odata/v2`, `internal/client/partnerdirectory`,
  `internal/client/securitycontent`, `internal/features`, `internal/provider`) was run after
  adding this feature specifically to confirm no regression to Integration Packages, Integration
  Flows, Value Mappings, Message Mappings, Script Collections, Access Policies, Partner
  Directory, or Security Content.

## `data.sapintegrationsuite_service_endpoints`

- **SAP product area**: Integration Suite / Cloud Integration — runtime discovery of deployed
  content's exposed endpoints
- **Official API**: Integration Content API, documented as "Endpoints of Runtime Artifacts" —
  "You can use resource `ServiceEndpoints` to read all endpoints provided for integration flows
  and to get the number of endpoints" (confirming `$inlinecount`/count support). Base path
  `https://<host>/api/v1/ServiceEndpoints`, the same `/api/v1` OData V2 host as every other Cloud
  Integration resource this provider uses.
- **Entity set**: `ServiceEndpoints`, with two expandable navigation properties, `EntryPoints`
  and `ApiDefinitions`
- **Protocol**: OData V2. GET only — no create/update/delete operation is documented, since SAP
  generates these entities itself from deployed content.
- **Primary source**: SAP's own `ServiceEndpoints Example Requests` page (read via the
  `SAP-docs/btp-integration-suite` GitHub mirror of the official Help Portal content), which
  gives the complete confirmed contract:
  - `Name` and `Protocol` are the two documented, filterable top-level properties (`Name eq
    '...'`, `Protocol eq '...'`).
  - `EntryPoints` (expand via `$expand=EntryPoints`): an array of `EntryPoint`, with `Name`
    (required, String), `URL` (required — see casing note below), and `Type` (optional,
    enumerated String: `DEV`, `TEST`, `PROD`, `SANDBOX`).
  - `ApiDefinitions` (expand via `$expand=ApiDefinitions`): an array of `APIDefinition`, with
    `URL` (required) and `Type` (required, enumerated String: `oas-yaml`, `oas-json`, `raml`,
    `edmx`, `wsdl`).
  - Protocol values, per the same page's adapter table: SOAP adapter → `SOAP`, IDoc adapter →
    `SOAP`, OData V2 adapter → `ODATAV2`, AS2 → `AS2`, AS4 → `AS4`, HTTPS → `REST`. This provider
    exposes exactly this value; it does not reverse-map it back to an adapter type, since the
    mapping is not one-to-one (SOAP and IDoc both report `SOAP`).
- **`Url` JSON casing — confirmed from SAP's own Piper (open-source CI/CD) library, not from
  prose alone**: SAP's documentation prose describes the property's *type* as "URL", which does
  not by itself establish the wire-format JSON key casing. `github.com/SAP/jenkins-library`'s
  `cmd/integrationArtifactGetServiceEndpoint.go` parses a real `ServiceEndpoints` response with
  `entryPoints.Path("results.0.Url")` — confirming the actual JSON property is `Url`, not
  `URL`. This provider's `EntryPoint.URL` Go field is tagged `json:"Url"` accordingly. The same
  Piper source also confirms the response envelope shape this provider already assumes
  everywhere (`{"d": {"results": [...]}}`) and that `EntryPoints`, once expanded, nests its own
  `{"results": [...]}` array (`internal/client/odata/v2.ExpandedCollection[T]`, a new small
  shared type distinct from the top-level paged-collection envelope, since SAP does not document
  paging for an expanded nested navigation property the way it does for a top-level collection
  request).
- **`ApiDefinitions[].Url` casing — inferred, not independently confirmed**: no example source
  touching the `ApiDefinitions` expansion specifically (as opposed to `EntryPoints`) was found.
  This provider applies `Url` by consistency with the confirmed `EntryPoints` casing and SAP's
  Pascal-case OData convention elsewhere on this same entity, and records this as an open
  verification item in `internal/features/catalog.go`'s Limitations for
  `cloud_integration.service_endpoints`.
- **Combined `$expand=EntryPoints,ApiDefinitions`**: SAP's example requests demonstrate each
  expansion separately, never combined in one request. This provider combines them in a single
  request using OData V2's standard comma-separated `$expand` syntax (the same `Query.Expand
  []string` mechanism already used for `$select` elsewhere in this codebase) to avoid an N+1
  request pattern; no documented reason was found that `ServiceEndpoints` would reject a
  combined expand, and this is standard, unremarkable OData V2 behavior.
- **Pagination**: `$top`/`$skip`/`$inlinecount` support was added in the February 2020 release
  (v3.21.x) per SAP's own release notes (community source). This provider always fetches every
  page via the shared `GetAllPages` helper (the same server-driven `__next`-link paging every
  other collection-returning client method in this provider uses), rather than assuming a single
  page.
- **No confirmed technical ID**: unlike `AccessPolicies` (numeric `Id`) or `IntegrationPackages`
  (user-assigned `Id`), no SAP source found documents a stable identity field for a
  `ServiceEndpoints` entry beyond `Name`/`Protocol` together, and even that combination's
  uniqueness was not confirmed for every possible deployment configuration (see
  `docs/resource-design.md` for why this rules out a singular `data.sapintegrationsuite_service_endpoint`
  lookup).
- **Deterministic ordering**: SAP does not document a guaranteed response order for the
  collection, or for either expanded nested collection. This provider sorts all three
  deterministically before writing Terraform state — see `internal/client/cloudintegration/service_endpoint.go`.
- **Required roles**: not separately confirmed for this research pass; assumed to fall under the
  same Integration Content read scopes already required for `IntegrationDesigntimeArtifacts`/
  `IntegrationRuntimeArtifacts`, pending confirmation.

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
  "to give dedicated users access to the artifacts protected by the access policy, you define
  a role and associate it with the access policy using SAP Business Technology Platform
  cockpit, and only users that are assigned to that role can access the artifacts". Read
  literally, this describes a BTP **role** (for example a custom role built from a role
  template in the subaccount's Security > Roles area), not the BTP **role collection** itself
  — a role collection bundling that role is what actually gets assigned to users. This
  provider does not manage the BTP-side role or role collection (see
  `docs/provider-scope.md`); it treats `role_name` as an opaque string the practitioner
  supplies and never interprets, resolves, or cross-checks against BTP. It is still not
  confirmed against OData `$metadata` whether `AccessPolicies.RoleName` is simply a label
  matched against that BTP-side role's name, or carries additional Integration
  Suite-specific semantics. Until confirmed, `role_name` stays `RequiresReplace()` (immutable)
  and this provider avoids relying on renaming it having any particular effect.
- **`operator` semantics — confirmed**: SAP's documentation states plainly that choosing the
  "equals" operator requires the exact artifact name/ID as the value, while "matches" requires
  "a valid Java Regular Expression" that must be "supported by the Java Pattern class" — i.e.
  `java.util.regex.Pattern`, not a wildcard or glob syntax. This resolves a prior open
  question in this document; see `docs/resource-design.md` for the resulting example-value
  guidance. The exact wire-format casing of the `Operator` property's values
  (`EQUALS`/`MATCHES`, as currently coded, versus `equals`/`matches` or another casing) remains
  unconfirmed — SAP's prose and UI labels do not by themselves establish the OData enum's wire
  casing, and this project could not reach `$metadata` or a live tenant to confirm it in this
  research pass.
- **`attribute` values — confirmed set, unconfirmed casing**: SAP's documentation confirms
  exactly two attribute choices, "Name" and "ID" (referencing an artifact by name or by its
  technical ID). The exact wire-format casing (`Id` vs. `ID`, as currently coded) is likewise
  unconfirmed against `$metadata` or a live tenant.
- **Runtime replication and reconciliation — confirmed as a real Integration Suite feature,
  unconfirmed as part of the public API surface**: SAP's "Manage Access Policies" application
  documentation (including the Edge Integration Cell-specific variant of that page) describes
  replicating an access policy to one or more runtimes — the Cloud Integration runtime,
  Integration Cell, and Edge Integration Cell are all named as replication targets — and
  checking each target's reconciliation status, one of `Fail`, `Success`, or `Pending`, via an
  icon in the UI's "Runtimes" column. This confirms the underlying concept is real and not
  this provider's invention. It does **not** confirm that this behavior (viewing or triggering
  replication, reading per-runtime reconciliation status) is exposed through the public
  `AccessPolicies` OData API this provider uses, as distinct from being an application-UI-only
  capability layered on top of it; every attempt to reach a primary source describing the
  `AccessPolicies` entity's actual OData properties for this research pass was blocked
  (`help.sap.com`, `api.sap.com`, and `community.sap.com` are all unreachable from this
  project's environment) or returned no result naming a `ReconciliationStatus`-shaped
  property. Given that, this provider keeps `reconciliation_status` as an existing,
  best-effort `Computed` field (harmless if SAP's API never populates it) but does not add any
  Integration Cell- or Edge Integration Cell-specific resource, and does not implement polling
  to a terminal reconciliation state — see `docs/resource-design.md`. This remains the single
  largest confirmed gap between what the SAP application can do and what this provider's
  public API access can verify.
- **Required roles**: Integration Suite "Manage Security" / access-policy administration
  scopes; SAP's documentation additionally names the `PI_Administrator` role collection as
  required to create and edit access policies through the application UI.

## `sapintegrationsuite_user_credential` / `sapintegrationsuite_oauth2_client_credential`

- **SAP product area**: Integration Suite / Security — Security Content ("Security Material" in
  the tenant UI, under *Monitor* > *Manage Security* > *Security Material*)
- **Official API**: Security Content API, published on SAP Business Accelerator Hub, same
  `/api/v1` OData V2 host as the Cloud Integration content APIs (confirmed via
  `internal/client/cloudintegration/client.go`'s existing doc comment, which already noted this
  client "covers the 'Integration Content' and 'Security Content' OData V2 services under
  `/api/v1`" before this feature family was implemented). A dedicated
  `internal/client/securitycontent` client package is used instead, kept separate from
  `cloudintegration` because these entities have fundamentally different secret semantics.
- **Entity sets**: `UserCredentials`, `OAuth2ClientCredentials`
- **Protocol**: OData V2
- **Operations**: GET, POST, PUT, DELETE. `PUT` is used for Update (full redeploy), matching
  SAP's Manage Security Material UI, which documents an explicit *Edit* action for Credentials
  artifacts ("You can also edit and redeploy an existing artifact") and states that editing an
  OAuth2 Client Credentials artifact specifically requires re-entering the client secret every
  time. This project could not confirm whether a successful `PUT` returns the updated entity
  body or `204 No Content`, so both resources re-read the entity with `GET` after every `PUT`
  rather than trusting the `PUT` response shape — the same defensive pattern already used by
  `UpdateAccessPolicy`.
- **`UserCredentials` fields — confirmed**: `Name` (the artifact's alias, used as both display
  name and OData key — SAP's own UI documentation states "the artifact name is used as an alias
  for the confidential data"), `User`, `Password` (write-only, never read back), `Description`.
  Corroborated by a documented third-party example payload (`POST .../UserCredentials` with a
  JSON body of `Name`, `Kind`, `Description`, `User`, `Password`, `CompanyId`) rather than this
  project's own inspection of a live tenant's `$metadata`.
- **`UserCredentials` fields — confirmed existence, unconfirmed casing**: `Kind` (SAP's UI calls
  this "Type": empty/unset for a generic Basic/username-token credential, `SuccessFactors`, or
  `OpenConnectors`) and `CompanyId` (only meaningful when `Kind` is `SuccessFactors`; SAP's UI
  hides this field for every other kind). Both are corroborated by the same third-party example
  payload as above, not `$metadata`.
- **`UserCredentials` fields — not exposed**: a deployment status (SAP's UI shows
  Stored/Deployed/Error for security material generally). This project could not confirm the
  OData property name for it and would rather omit a `Computed` attribute than expose one that
  is silently always empty.
- **`OAuth2ClientCredentials` fields — confirmed**: `Name`, `Description`, `TokenServiceUrl`,
  `ClientId`, `ClientSecret` (write-only, never read back), `Scope`. Confirmed directly from
  SAP's Help Portal documentation for "Deploying an OAuth2 Client Credentials Artifact", which
  describes each field in prose with an unambiguous meaning — a stronger source tier than the
  `UserCredentials` third-party example payload.
- **`OAuth2ClientCredentials` fields — documented in the UI, not implemented**: Grant Type
  (whether the grant type is sent as part of the URL or the request body), Client Authentication
  (whether the client ID/secret are sent as a body parameter or an `Authorization` header),
  Resource, Audience, and up to 20 custom Key/Value/"Send as Part of" parameters. SAP's Help
  Portal documents all of these in prose, but this project could not confirm their OData
  property names or JSON shapes against `$metadata` or a documented example payload, so they are
  deliberately left unimplemented rather than guessed at.
- **Write-only secret design**: see `docs/guides/security-content.md` for the full rationale.
  In short: `password_wo`/`client_secret_wo` are Terraform Plugin Framework `WriteOnly`
  attributes (requires Terraform CLI 1.11+), paired with a plain `password_wo_version`/
  `client_secret_wo_version` string that is the sole signal Terraform uses to decide whether to
  redeploy the credential. Both Go client types (`UserCredential`, `OAuth2ClientCredential`)
  structurally have no field for the secret, so it cannot end up in state, a diagnostic, or a
  log line regardless of what SAP's response body contains.
- **Required roles**: not separately confirmed for this research pass; assumed to fall under the
  same Integration Suite "Manage Security" scopes as Access Policies and Keystore
  administration, pending confirmation.

## `security.*` catalog entries not yet implemented

See `docs/guides/security-content.md` for the full list and reasoning (Keystore Entries,
Certificate, Key Pair, SSH Key, Certificate Chain, Certificate-User Mapping, Secure Parameter,
Known Hosts, OAuth2 Authorization Code, OAuth2 SAML Bearer Assertion, PGP keyrings). The one
correction worth calling out here specifically: **Certificate-User Mapping** was previously
cataloged as `PublicAPI: true` / `not_implemented`. Reverifying it for this feature family found
that SAP's certificate-to-user mapping documentation ("Managing Certificate-to-User Mappings",
"Client Certificate Authentication and Certificate-to-User Mapping (Inbound)", "Setting Up
Inbound HTTP Connections with Certificate-to-User Mapping") exists only under the **Neo**
environment, with no Cloud Foundry equivalent found anywhere in SAP's published documentation.
Since this provider targets Cloud Foundry, the catalog entry is corrected to `PublicAPI: false`
/ `no_public_api`.

## Partner Directory API

- **SAP product area**: Integration Suite / Cloud Integration — Partner Directory
- **Official API**: Partner Directory OData V2 API, under the same `/api/v1` service root as
  Integration Content and Security Content — confirmed via SAP's own "Partner Directory",
  "Partner Directory Concepts", "Partner Directory Entity Types", and "Requests for String
  Parameter, Binary Parameter, and Authorized User" documentation pages.
- **Entity sets and confirmed operations**:
  - `Partners` — GET (read) is documented; no confirmed create operation. SAP documents Pid
    uniqueness as "ensured by the tenant owner application" (the caller picks the value, it is
    not server-allocated), and describes deleting a Pid as capable of removing every entry
    belonging to it in the Partner Directory "via one call". No `sapintegrationsuite_partner`
    resource exists because of this — see `docs/resource-design.md`.
  - `StringParameters` — GET/POST/PUT/DELETE, key `(Pid, Id)`. A documented example request
    body: `{"Pid":"partner1","Id":"sp1","Value":"sp1v"}`.
  - `BinaryParameters` — GET/POST/PUT/DELETE, key `(Pid, Id)`, `Value` base64-encoded.
    Documented `ContentType` values: `xml`, `xsl`, `xsd`, `json`, `text`, `zip`, `gz`, `zlib`,
    `crt` (with encoding-suffixed variants such as `xml;encoding=UTF-8` also valid — this
    provider does not restrict `content_type` to the short documented list for that reason).
    SAP documents a 260 KB maximum `Value` size and recommends storing larger uncompressed
    XML/XSL/XSD content as a `zip` instead (auto-unzipped by the XML Validator and XSLT Mapping
    steps).
  - `AlternativePartners` — GET/POST/PUT/DELETE. The documented entity carries both plain
    (`Agency`, `Scheme`, `Id`, `Pid`) and hex-encoded (`Hexagency`, `Hexscheme`, `Hexid`) fields;
    a documented example request URL, `AlternativePartners(Hexagency='6167656e637931',...)`,
    confirms the hex form is the actual OData key and that it is the lowercase hex encoding of
    the plain value's UTF-8 bytes (`"agency1"` → `"6167656e637931"`). Creation uses the plain
    fields; SAP computes the hex key itself.
  - `AuthorizedUsers` — GET/POST/PUT/DELETE, key `User`. SAP documents this as many-to-one
    (one communication user per Pid, a Pid can have several). Documented example body:
    `{"User": "...", "Pid": "PartnerZ"}`. Whether `User` is case-normalized internally was not
    confirmed against a primary source; this provider does not normalize it.
  - `UserCredentialParameters` — POST (create) and DELETE are documented; no confirmed
    PUT/PATCH for updating an existing credential's password. Documented example body:
    `{"Pid":"Receiver_1","Id":"USER","User":"...", "Password":"..."}`. SAP documents the
    generated security-artifact alias format as `pd:<Pid>:<Id>:UserCredential`, matching the
    property set via the exchange property `RECEIVER_CREDENTIAL` in scripts. SAP additionally
    documents that `UserCredentialParameter` and `CertificateUserMapping` cannot be included
    together with other entity types in a single OData ChangeSet (batch) request.
- **Password read-back — could not confirm either way**: no primary source was found
  definitively stating whether a GET (or the POST response) on `UserCredentialParameters`
  returns the password. Given that genuine uncertainty for a security-sensitive credential,
  this provider's `UserCredentialParameter` Go type (used for every response this client
  decodes) simply has no `Password` field at all, so the answer to "does this provider ever
  expose it" is "no" regardless of what SAP's API actually does.
- **CSRF protection**: SAP's OData V2 services on Cloud Foundry/BTP require a valid
  `X-CSRF-Token` for POST/PUT/PATCH/DELETE, obtained via a GET with `X-CSRF-Token: fetch`
  against the same resource, independently of OAuth authentication (OAuth proves identity;
  CSRF proves the write was not forged). This is documented generically for SAP's Cloud
  Integration `api/v1` OData services, not specifically restated on the Partner Directory
  pages, but there is no indication Partner Directory's shared service root is exempt from it.
  Implemented once in the shared HTTP client (`internal/client/http/csrf.go`) rather than
  per-resource, so it applies to every write this provider makes, not just Partner Directory's.
- **Pagination**: SAP's OData V2 services return a `__next` link (a complete absolute URL) in
  `d.__next` for server-driven paging; `GetAllPages` in `internal/client/odata/v2` follows it
  until exhausted. String Parameters in particular is documented as capable of holding large
  numbers of entries per tenant.
- **Required roles**: the Cloud Foundry role template `AuthGroup_TenantPartnerDirectoryConfigurator`
  is documented as required to work with a tenant's Partner Directory entities via the OData
  API (`AuthGroup_Administrator` also works). This provider does not manage that role or role
  collection — see `docs/provider-scope.md`.
- **Security note**: SAP explicitly documents that ordinary Partner Directory data (string and
  binary parameters) is stored unencrypted. This provider's documentation and resource
  descriptions explicitly warn against storing secrets there; only
  `UserCredentialParameters` is treated as a credential store, and even that is modeled
  conservatively — see the write-only `password_wo` design in `docs/resource-design.md`.

## Deferred APIs (tracked, not yet implemented)

| Area | API | Status |
|---|---|---|
| Classic API Management | "Accessing API Management APIs Programmatically" REST/OData APIs | Confirmed public, deferred to a later minor version |
| New API Gateway / API Artifacts | Design-time API for API-centric artifacts with Runtime Profile (Integration Cell / Edge Integration Cell) | Existence confirmed via UI/feature docs; exact public API surface not yet confirmed in enough detail for a stable Terraform schema — deferred to v0.2.x |
| Security material — keystore entries, certificates, key pairs, SSH keys, certificate chains | Security Content API | Existence confirmed, exact `$metadata` field casing not confirmed — deferred, see `docs/guides/security-content.md` (user credentials and OAuth2 client credentials are now implemented) |
| Value mapping entry-level management | `UpsertValMaps`, `UpdateDefaultValMap`, `DeleteValMaps` | Confirmed public, deferred — exact payload/path shapes and delete granularity not confirmed against a reachable primary source; see `docs/resource-design.md` |

## Explicitly ruled out

- **Integration Suite capability activation** (Cloud Integration, API Management, Integration
  Cell, Edge Integration Cell): no public activation API found. SAP's own documentation
  describes activation as a UI action ("Settings" → "Runtime" → *Activate*). Not implemented;
  see `docs/provisioning-capability-matrix.md`.
- Any endpoint only reachable by reverse-engineering the Integration Suite UI's network
  traffic. None were used, and none will be, regardless of how convenient they would be.
