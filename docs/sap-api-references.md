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

## `sapintegrationsuite_custom_tag_configuration`

- **SAP product area**: Integration Suite / Cloud Integration — tenant-wide governance
  configuration, not a per-package or per-artifact object
- **Official API**: Integration Content API. Confirmed directly from two SAP Help Portal pages
  (read via the `SAP-docs/btp-integration-suite` GitHub mirror): "Create New Custom Tags
  Configuration" and "Get Custom Tags Defined on the Tenant".
- **Entity set**: `CustomTagConfigurations`, with a single, fixed, confirmed key: `CustomTags`.
  There is exactly one configuration per tenant.
- **Protocol**: OData V2
- **Create/Update — confirmed verbatim, including the exact example payload**: `POST
  /api/v1/CustomTagConfigurations`, body `{"CustomTagsConfigurationContent":
  "<base64-encoded-content>"}`, where the decoded content for a mandatory "Owner" tag is
  `{"customTagsConfiguration":[{"tagName":"Owner","isMandatory":true}]}` — this exact
  base64-round-trip is asserted byte-for-byte in
  `internal/client/cloudintegration/custom_tag_configuration_test.go`. SAP's documentation
  states: "If a custom tags configuration is already available on the tenant, add the following
  query parameter to the request: `Overwrite=true`". This client always sends `Overwrite=true`
  on every write (see the doc comment on `SetCustomTagConfiguration` for why: SAP never
  documents what a plain `POST` does once a configuration exists, so always using the documented
  "already exists" path avoids depending on an unconfirmed distinction).
- **Read — confirmed verbatim, including the "$value" envelope difference**: `GET
  /api/v1/CustomTagConfigurations('CustomTags')/$value`. Unlike every other entity in this
  client, the response is the raw decoded JSON directly — OData's `$value` raw-media-stream
  convention — not wrapped in the standard `{"d": {...}}` envelope `v2.DecodeEntity` expects
  elsewhere, so this client decodes the response body directly instead.
- **Delete — reverified, confirmed absent**: no delete or clear operation is documented anywhere
  for this entity across either page or the general Integration Content resource table (which
  describes the entity only as maintainable "using the Cloud Integration Settings section" or
  "the OData API (Custom Tags interface)", with no mention of removal). This is a materially
  different situation than a design-time artifact type with an unconfirmed delete *scope* (for
  example whether delete removes one version or all) — here there is no delete operation
  documented at all. `sapintegrationsuite_custom_tag_configuration`'s Delete therefore returns
  an explicit error rather than a guessed "clear via empty overwrite" implementation or a no-op
  — see `docs/resource-design.md` and `docs/guides/custom-tag-configurations.md`.
- **`Overwrite=true` semantics — strongly implied, not stated verbatim**: the documented request
  body is always the complete configuration, never a delta, and the query parameter is literally
  named `Overwrite`, which this provider reads as "full replace" (tags absent from a new write
  are removed). SAP's documentation never uses the word "replace" itself, so this is flagged as
  this provider's own reasonable interpretation, not an independently confirmed fact.
- **Singleton semantics investigated (per the research checklist for this feature)**:
  - Is `CustomTags` always the fixed key? **Confirmed** — it appears literally in SAP's own GET
    example.
  - Can there be multiple configurations? No evidence of any mechanism for more than one; the
    entire API is structured around the one fixed key.
  - Does a plain `GET` (without `/$value`) return useful metadata? Not shown in either
    documented example; not implemented, since guessing at undocumented properties was rejected.
  - Is posting an empty configuration valid, and does it clear everything? **Not confirmed
    either way** — deliberately not relied upon by this provider's Delete (see above).
  - Tag name / permitted-value uniqueness, case sensitivity, and whether SAP preserves submitted
    order: **none of these are documented**. This provider enforces tag-name uniqueness itself,
    models `permitted_values` as a set (making exact duplicates unrepresentable), and treats
    both `tags` and `permitted_values` ordering as not semantically meaningful (see
    `docs/guides/custom-tag-configurations.md`).
- **A documentation artifact worth flagging**: SAP's "Get Custom Tags Defined on the Tenant"
  page's second example response renders two permitted values ("Mr. Bean" and "Ms. Bean") as a
  *single* array element containing a comma-separated string —
  `"permittedValues":["Mr. Bean, Ms. Bean"]` — rather than two separate array elements. This
  contradicts the one-element-per-value convention every other array-typed field in this API
  uses (including `EntryPoints`/`ApiDefinitions` on `ServiceEndpoints`), so it is treated as a
  documentation authoring artifact rather than a confirmed wire format; this client always
  serializes one array element per permitted value.
- **Required role**: SAP's general Integration Content API documentation states that maintaining
  custom tags requires the role template `WebToolingSettingsProductProfiles.savetenantconfiguration`
  ("part of role collection `PI_Administrator`" in the Cloud Foundry environment, "part of
  authorization group `AuthGroup.Administrator`" in Neo). The role **template**, not the entire
  `PI_Administrator` role collection, is what this provider's documentation states is actually
  required — a narrower custom role collection granting just that template should work equally
  well, and this provider does not claim the full `PI_Administrator` collection is mandatory.

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

## `sapintegrationsuite_number_range` and the Message Stores API family (Variables, Data Stores, Data Store Entries)

- **SAP product area**: Integration Suite / Cloud Integration — the "Message Stores" OData V2 API
  (`https://api.sap.com/api/MessageStore`), covering `NumberRanges`, `Variables`, `DataStores`,
  and `DataStoreEntries` alongside `Entries`/`EntryAttachments`/`EntryProperties` (Message Store,
  already covered elsewhere) and JMS Resources.
- **Research method**: every page below was fetched from the `SAP-docs/btp-integration-suite`
  GitHub mirror (`docs/ci/Development/` and `docs/ci/Operations/`), the same official-mirror
  technique used throughout this project. `api.sap.com` itself redirects unauthenticated
  requests to a login page and could not be used directly to inspect `$metadata`.

### Number Ranges — confirmed operations and the missing GET

The curated **"Message Stores Example Requests"** index page
(`message-stores-example-requests-02c57df.md`) is the authoritative list of every documented
example for this API family. For every sibling entity it links a "Get ..." example page; for
Number Ranges it links exactly two:

- [Add a Number Ranges Object](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/add-a-number-ranges-object-b1bd945.md):
  `POST /api/v1/NumberRanges`, body
  `{"CurrentValue":"0","Name":"My NRO Object","MinValue":"0","MaxValue":"9999","Description":" Number Range Object ","Rotate":"true","FieldLength":"4"}`
  — asserted byte-for-byte in `internal/client/cloudintegration/number_range_test.go`.
- [Update a Number Ranges Object](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/update-a-number-ranges-object-139a6b2.md):
  `PUT /api/v1/NumberRanges('{objectName}')`, same body shape. This page also states explicitly:
  "The Current Value returned by the API corresponds to the Next Value shown in the Monitoring
  tab of the UI. The difference is only in terminology" — confirmed verbatim, cross-referenced
  by `managing-number-ranges-b6e17fa.md` (the UI documentation) independently.

**No GET operation — for either the collection or a single object — is documented anywhere.**
This was checked exhaustively, not assumed from absence on one page:

- The example-requests index above lists nothing beyond Add/Update.
- No `get-a-number-range*`/`get-number-range*` page exists under `docs/ci/Development/` (checked
  via a directory listing of the whole folder).
- `message-stores-1aab5e9.md` (the Message Stores overview/resource table) documents the
  `NumberRanges` resource in prose only, with no GET example, unlike its entries for
  `DataStores`, `DataStoreEntries`, and `Variables`, which each state their unsupported query
  options explicitly (implying a GET exists to apply those options to) — Number Ranges has no
  such statement.
- The overview page's resource table has a genuinely truncated caveat directly against
  "Number Ranges": `> ### Note: \n> Not supported in.` with no recoverable clause — re-fetched
  twice to rule out a fetch artifact; this is a truncation in SAP's own published source. Its
  meaning (an environment? a runtime type?) is **not confirmed**, and this provider does not
  guess at it.

**No DELETE operation is documented for Number Ranges either.** The overview page's general
CSRF-token paragraph mentions "POST, PUT, and DELETE" as the three modifying-action types across
the whole Message Stores API family, but this is generic boilerplate applying to the family as a
whole (DELETE is separately, specifically confirmed for Data Store Entries/Variables via the
`DataStoresAndQueuesDelete` role template — see below), not evidence of a Number-Ranges-specific
DELETE. `managing-number-ranges-b6e17fa.md` (the UI/Operations page) instead documents an
**"Undeploy"** action, explicitly distinct from "Delete" in its own Actions list, with no visible
REST equivalent anywhere in the API documentation.

**A UI-only multi-runtime deployment dimension was also found, with no API-level counterpart.**
The same Operations page documents a "Runtimes" field on the Add/Edit dialog: "One or more
runtime nodes to deploy the artifact to... including Cloud Integration and any active Edge
Integration Cell nodes," and describes name-uniqueness as evaluated against "any of the selected
runtimes." Neither of the two documented API examples (Add, Update) shows any runtime/location
parameter. This provider's client always targets the implicit default runtime and does not
attempt to reconstruct or guess at this parameter.

**Field semantics confirmed from `managing-number-ranges-b6e17fa.md`** (UI documentation,
consistent with the API examples):

- `MinValue`: "should be greater than or equal to 0."
- `MaxValue`: "should be less than 15 digit[s]" / "Must be fewer than 15 digits."
- `FieldLength`: zero-pads the displayed value; "maximum value allowed for this attribute is 14";
  a value of 0 applies no padding.
- `Rotate`: "If this attribute is set and the number range reaches specified maximum value, then
  the current value resets to specified minimum value" — confirmed verbatim, matching this
  provider's `rotate` semantics exactly.

**Role template**: no Cloud Foundry role template specific to Number Ranges creation/update was
found documented anywhere (unlike Data Stores/Variables, which have explicit
`DataStoresAndQueuesRead`/`DataStorePayloadsRead`/`DataStoresAndQueuesDelete` templates — see
below). The Neo-environment `tasks-and-permissions-556d557.md` page lists Monitor-app tasks
"View number ranges" and "Add, edit, or undeploy number ranges" gated behind the Integration
Developer / Tenant Administrator roles, but this documents UI authorization in the older Neo
environment, not a confirmed Cloud Foundry API role/scope for the public REST endpoints.

**Consequence for this resource's design** (see `docs/resource-design.md` for the full
suitability walkthrough): without a GET, `sapintegrationsuite_number_range` cannot implement a
Read that verifies anything against the tenant, cannot detect drift, and cannot support
`terraform import`. It is implemented as a **write-only-lifecycle resource**: Create and Update
call the two confirmed operations for real, Read is a documented no-op that trusts local state,
Delete and Import both return explicit errors rather than guessing at unconfirmed operations. The
runtime counter (`CurrentValue`) is handled as a version-gated write-only attribute
(`current_value_wo`/`current_value_wo_version`) that is sent to SAP only when the practitioner
deliberately bumps the version — every other Update omits `CurrentValue` from the request body
entirely, since this provider has no way to confirm any remembered value is still correct. Two
unconfirmed risks are inherent to this design and documented rather than hidden: (1) whether an
omitted `CurrentValue` on `PUT` is preserved unchanged or reset to a default by SAP is not
confirmed either way, since there is no GET to check the result; (2) whether `POST`ing a `Name`
that already exists on the tenant errors, conflicts, or silently overwrites is likewise
unconfirmed.

### Variables — read-only download, no creation API

[Download a Variable](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/download-a-variable-94e6799.md)
documents the **only** public operation for this entity: `GET
/api/v1/Variables(VariableName='{VariableName}',IntegrationFlow='{IntegrationFlowName}')/$value`,
which downloads the raw value via OData's `$value` convention — no structured metadata
(`Visibility`/`UpdatedAt`/`RetainUntil`) is returned by this endpoint. No collection GET exists
(the composite key must already be known), and no POST/PUT/DELETE is documented anywhere.
[Define Write Variables](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/define-write-variables-de04b75.md)
confirms Variables are created and updated exclusively by an integration flow's "Write
Variables" step (Constant/Header/XPath/Expression/Property source, a "Global Scope" checkbox
controlling Global vs. local `IntegrationFlow` visibility), and separately confirms: "A variable
gets expired after the retention period, which is 400 days," extended by every successful
processing run, and "Variables can't be downloaded using the data store viewer."

**No resource, no data source** — see `docs/resource-design.md` and `docs/provider-scope.md`.
There is no Create/Update API to back a resource at all, and the one confirmed read endpoint
returns nothing but the arbitrary runtime value itself, with no safer metadata-only projection
available.

### Data Stores and Data Store Entries — GET-only, runtime-created

[Get All Data Stores with Overdue Messages](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/get-all-data-stores-with-overdue-messages-5173f5c.md):
`GET /api/v1/DataStores?overdueonly=true`, an aggregate monitoring endpoint returning
`NumberOfMessages`/`NumberOfOverdueMessages` per store — not configuration.
[Get Single Data Store Entry](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/get-single-data-store-entry-8b86912.md)
and
[Get All Data Store Entries for a Data Store](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/get-all-data-store-entries-for-a-data-store-acbef52.md)
confirm `DataStoreEntries` fields verbatim: `Id`, `DataStoreName`, `IntegrationFlow`, `Type`,
`Status`, `MessageId`, `DueAt`, `CreatedAt`, `RetainUntil` — all runtime message state.

[Define Data Store Write Operations](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/define-data-store-write-operations-46260ee.md)
(referenced by the overview page as how both `DataStores` and `DataStoreEntries` come into
existence) and
[Define Data Store Delete Operations](https://raw.githubusercontent.com/SAP-docs/btp-integration-suite/main/docs/ci/Development/define-data-store-delete-operations-5efa3ac.md)
confirm both write and delete are **design-time integration-flow steps**, not REST operations on
these entities: "This step deletes an entry from a transient data store... the delete operation
can't be used to delete whole data stores, but single entries only." No POST/PUT/DELETE against
`DataStores` or `DataStoreEntries` is documented as a REST call anywhere.

Confirmed verbatim from `message-stores-1aab5e9.md`: both the `DataStore` and `Variables` APIs
"do not support the following query options: `$filter`, `$inlinecount`, `$orderby`, `$skip`,
`$top`, `$expand`, `$select`" — two separate, explicit statements, one per entity, not a single
blanket statement this provider generalized across the whole API family. `NumberRanges` carries
no such statement (and, per above, has no GET to apply query options to regardless).

**Required roles** (from `message-stores-1aab5e9.md`'s Permissions section, Cloud Foundry): to
view data store entries, `DataStoresAndQueuesRead`; to download payloads and view variables,
`DataStorePayloadsRead`; to delete data store entries/variables, `DataStoresAndQueuesDelete`. The
existence of `DataStoresAndQueuesDelete` confirms a delete capability is authorized for Data
Store Entries/Variables at the permission-template level, consistent with the confirmed
design-time Delete step above, even though no REST DELETE example was found for either entity.

**No resource, no data source for either entity** — see `docs/resource-design.md` and
`docs/provider-scope.md`. `DataStores` has no independent creation API (it is an implicit runtime
container); its one GET is monitoring data, the same class already excluded via
`cloud_integration.message_processing_logs`. `DataStoreEntries` is unambiguously runtime business
message data with no creation/deletion REST API at all.

### Edge Integration Cell

The only hint of an Edge Integration Cell-specific runtime-location API path came from the
Number Ranges "Runtimes" UI field mentioning "any active Edge Integration Cell nodes" (see
above). No `/location/<runtime-location-id>/api/v1/...`-style path variant, or any other
Edge-specific endpoint, was found documented for `NumberRanges`, `Variables`, `DataStores`, or
`DataStoreEntries` in any page fetched this session. This remains genuinely unconfirmed and is
not implemented — consistent with this provider's rule against guessing at undocumented
endpoints.

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

## `sapintegrationsuite_certificate`, `sapintegrationsuite_key_pair`, and the rest of Security Content

- **SAP product area**: Integration Suite / Cloud Integration — the "Security Content" OData V2
  API (`https://api.sap.com/api/SecurityContent`), covering `KeystoreEntries`, `Keystores`,
  `KeystoreResources`, `HistoryKeystoreEntries`, `CertificateResources`,
  `KeyPairGenerationRequests`, `UserCredentials`, `OAuth2ClientCredentials`, and (conceptually,
  per SAP's own overview) `SecureParameters` and `CertificateUserMapping`.
- **Research method**: the same official-mirror technique used throughout this project —
  `docs/ci/Development/security-content-e01d3f0.md` (the overview/resource table) and
  `docs/ci/Development/security-content-example-requests-acb89ef.md` (the curated,
  authoritative index of every worked example SAP documents for this API family — the same
  pattern that resolved the Number Ranges research: an entity absent from this index has no
  documented example anywhere, even if it is mentioned conceptually elsewhere).

### Hex alias encoding — confirmed explicitly, not just by example

SAP's Security Content overview page states the rule directly, not merely by example: **"Hex
representation is used for the `Alias` field... Hex representation of the alias does mean that a
UTF-8-encoded byte array is built from the alias string. A hex string is then calculated from
this byte array."** It also explains why: **"the server doesn't allow slashes or backslashes in a
URI, even if they're percent encoded."** This is the same rule this project had already inferred
for Partner Directory's `AlternativePartners` (`Hexagency`/`Hexscheme`/`Hexid`) from one example
request URL alone; Security Content states it as an explicit rule. The shared implementation
(`internal/client/odata/v2/hexkey.go`, `EncodeUTF8Hex`/`DecodeUTF8Hex`) is asserted against every
alias SAP's own documentation uses as a worked example across both API families (`smtp.mail.
yahoo.com`, `mycertificate`, `mykeypair`, `baltimore cybertrust root`, `agency1`), plus Unicode,
punctuation, semicolon, slash, and backslash cases — see
`internal/client/odata/v2/hexkey_test.go`.

### Keystore Entries — confirmed read operations, confirmed field gap

- `GET /api/v1/KeystoreEntries` (all entries) and `GET /api/v1/KeystoreEntries('{Hexalias}')`
  (single entry by alias) are both confirmed with an identical documented example response:
  `Hexalias`, `Alias`, `KeyType`, `KeySize`, `ValidNotBefore`, `ValidNotAfter`, then a literal
  `....` truncation in SAP's own source.
- The overview page's prose separately states a keystore entry also carries "Subject DN and
  Issuer DN, and administrative information such as the time when the entry was last modified" —
  confirming those properties *exist*, but never giving their exact JSON property name or
  casing. This project does not guess at them: `sapintegrationsuite_certificate` derives subject
  DN, issuer DN, serial number, and a SHA-256 fingerprint locally instead, by parsing the
  certificate bytes it can already confirmedly retrieve with Go's own `crypto/x509` — see
  `internal/client/securitycontent/x509meta.go`.
- `GET /api/v1/KeystoreEntries('{Hexalias}')/Certificate/$value` (Export Certificate): confirmed,
  response content type `application/pkix-cert`, PEM body.
- `GET /api/v1/KeystoreEntries('{Hexalias}')/Sshkey/$value` (Export Public Key in OpenSSH
  Format): confirmed, response is the raw `ssh-rsa AAAA... <comment>` line. SAP documents this as
  supported only for RSA- and DSA-keyed entries ("Other algorithms, for example elliptic curve
  (EC), aren't supported").
- `PUT /api/v1/KeystoreEntries('{Hexalias}')?renameAlias=<new>` (Rename the Alias of a Keystore
  Entry): confirmed to exist, but deliberately not used by this provider — `alias` is
  `RequiresReplace` on both `sapintegrationsuite_certificate` and `sapintegrationsuite_key_pair`
  instead, keeping lifecycle behavior predictable, matching this provider's established
  precedent (`sapintegrationsuite_number_range`'s `name`). SAP's documentation for this operation
  states "The request body must contain a property of the `KeystoreEntry` entity type with a
  null value. Otherwise, you get an exception" — an unusual enough contract on its own to avoid
  relying on until there is a concrete reason to.
- **No field distinguishes SAP-owned from tenant-administrator-owned entries.** The overview page
  states in prose that "a keystore typically contains entries that belong to the tenant
  administrator and entries that are owned by SAP," but no property name for this distinction
  appears anywhere. This provider does not invent a heuristic (for example matching alias name
  patterns) to detect ownership in advance; instead, `sapintegrationsuite_certificate` and
  `sapintegrationsuite_key_pair` rely on and surface whatever error SAP's own server-side
  protection returns when an Update or Delete is actually attempted against a protected alias —
  the only mechanism this provider can trust for a fact SAP does not expose as data.

### Certificate — confirmed Create/Update, confirmed Delete via the shared keystore mass-delete

- `PUT /api/v1/CertificateResources('{Hexalias}')/$value` (Import and Update Certificate):
  confirmed for both create and update — SAP's own documentation flags the quirk explicitly:
  **"Although an HTTP PUT method is used, a new OData entity is created with the sample
  request."** SAP's documented example request body is enclosed in literal square brackets,
  `[-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----]`; no other example anywhere in
  either API family this project reviewed uses that convention for an inline value example, so
  this is treated as documentation formatting, not literal bytes to send — this client's
  `PutCertificate` sends the caller's PEM content exactly as given, with no bracket wrapping.
  This project could not independently confirm the request `Content-Type`; the documented GET
  response's `application/pkix-cert` is reused for the PUT as well, on the assumption that a
  `$value` endpoint's read and write representations are ordinarily the same media type — not an
  independently confirmed fact, flagged in code (`certificate.go`).
- No per-entity `DELETE` is documented for `CertificateResources` or `KeystoreEntries`. Delete
  uses the same confirmed `KeystoreResources('system')?deleteEntries=true` mass-deletion
  operation described below, with exactly the one alias the resource owns.

### Key Pair — confirmed generation, confirmed field table, confirmed absence of Update/GET-by-request

- `POST /api/v1/KeyPairGenerationRequests` (Generate a Key Pair): every field's
  mandatory/optional status, type, and default is confirmed verbatim from SAP's own "Input
  Properties" table — `Hexalias` (mandatory but "any dummy hex value," recalculated by SAP from
  `Alias`), `Alias`, `KeyType` (enum `RSA`/`DSA`/`EC`, default `RSA`), `SignatureAlgorithm`
  (enum depends on `KeyType`, default `SHA-512/RSA`), `KeySize` (mandatory for `RSA`/`DSA`,
  default 4096; for `EC`, 112-571 or use `KeyAlgorithmParameter` instead),
  `KeyAlgorithmParameter` (named EC curve, `EC` only), `CommonName` (mandatory), `Country`
  (mandatory, "two characters required"), `OrganizationUnit`/`Organization`/`Locality`/`State`/
  `Email` (all optional), `ValidNotBefore`/`ValidNotAfter` (optional, OData V2 JSON date literal
  wire format `/Date(<millis>)/`, confirmed directly from SAP's own worked example
  `"ValidNotBefore":"/Date(1469685029780)/"` — see `internal/client/odata/v2/date.go`).
  This provider's client always sends a fixed dummy `Hexalias` ("00"), matching SAP's own
  documented statement that the value is ignored and recalculated.
- **A documentation inconsistency worth flagging**: SAP's own worked request example writes
  `"SignatureAlgorithm":"SHA256/RSA"` (no hyphen), while the same page's field-table enum for the
  identical value is `"SHA-256/RSA"` (with a hyphen, matching every other enum value on the same
  table, including the default `"SHA-512/RSA"`). This client always sends the hyphenated form
  from the enum table and validates against it, treating the un-hyphenated inline example as a
  documentation artifact — the same category of quirk as Custom Tag Configuration's
  "Mr. Bean, Ms. Bean" comma-separated array example.
- **No GET is documented for `KeyPairGenerationRequests` or a distinct `KeyPairResources`
  entity** — the generation request is a one-shot command, not a persistent object with its own
  read-back. The persistent representation is the resulting `KeystoreEntries` entry, read via the
  same confirmed `GET KeystoreEntries('{Hexalias}')` used for Certificate — but that GET only
  confirms `KeyType`/`KeySize`/`ValidNotBefore`/`ValidNotAfter` back; `SignatureAlgorithm`,
  `KeyAlgorithmParameter`, and every subject DN field (`CommonName`, `Country`, etc.) are **not**
  confirmed returned by any documented GET. `sapintegrationsuite_key_pair` reads back and
  refreshes only the confirmed subset on every plan; the rest is trusted from the last successful
  write, never re-verified, which is why this feature is cataloged `partial` rather than
  `supported`.
- **No update operation is documented** for a generated key pair's material — every attribute
  that defines it is `RequiresReplace`.
- Delete uses the same confirmed `KeystoreResources('system')?deleteEntries=true` mass-deletion
  operation, exactly one alias.

### SSH Key — reverified and corrected: no separate resource exists

SAP's Security Content API overview's own Resources table lists **no independent "SSH Key"
entry** — only Certificate, Key Pair, Keystore Entry, Keystore, Keystore History, User
Credentials, Secure Parameter, OAuth2 Client Credentials, Certificate-to-User-Mapping (Neo), and
Access Policies/Artifact References. The Operations-guide page
`docs/ci/Operations/creating-a-key-pair-ssh-key-pair-b8a8601.md` ("Creating a Key Pair/SSH Key
Pair") confirms why: its UI dialog is literally the same field set for both — "In the *Current*
tab, choose *Create* > *Key Pair* or *Create* > *SSH Key* depending upon your requirement," with
one shared attribute table (Alias, Key Type, Key Size, Signature Algorithm, the subject DN
fields, validity dates). No separate `SSHKeyGenerationRequests` field contract (mandatory/
optional status, types, example body) was found documented anywhere in either the Development or
Operations documentation tree. Given this, `sapintegrationsuite_key_pair` covers the SSH use case
directly: an RSA or DSA key pair's public key is exposed as `public_key_openssh` via the
confirmed `GetSSHPublicKey` export, with no separate resource. One inconsistency worth noting:
the "Creating a Key Pair/SSH Key Pair" UI dialog lists Key Type options as RSA/EC ("DSA (only for
key pair)"), while the SSH-export documentation says RSA and DSA are supported and EC is not —
these two SAP-authored pages do not fully agree with each other; this provider follows the
export endpoint's own documented restriction (RSA/DSA) for `public_key_openssh` population.

### Certificate Chain — reverified: folded into Key Pair, no independent contract found

The overview page's Key Pair resource description states the API can be used to **"create a
certificate signing request, or import and export the related certificate chain"** — Certificate
Chain is documented as a *capability of* Key Pair, not an independently exampled resource. No
`CertificateChainResources` entry appears in the curated example-requests index, and no CSR/
signing-response field contract was found documented in either the Development or Operations
tree. Not implemented this phase; see `docs/resource-design.md` for the suitability conclusion
and the likely eventual naming (`sapintegrationsuite_key_pair_certificate_chain`) if a concrete
contract is later confirmed.

### Secure Parameter and Known Hosts — reverified, both remain without a confirmed contract

- **Secure Parameter**: the overview page's Resources table does list "Secure Parameter"
  conceptually, alongside User Credentials/OAuth2 Client Credentials (which do have confirmed
  full contracts). But the curated example-requests index — the same authoritative page that
  correctly enumerated every other confirmed Security Content operation — lists **zero** example
  requests for it, and its own "Deploying a Secure Parameter Artifact" documentation page
  describes only the Eclipse/Node-Explorer design-time deployment wizard ("In the Node Explorer
  you have selected a tenant, in the context menu you have selected *Deploy Artifacts*..."), not
  a REST contract. This, combined with prior third-party evidence of an OData error resolving a
  `SecureParameters` entity set, keeps this feature unconfirmed rather than implemented.
- **Known Hosts**: does not appear in the overview page's Resources table at all — a stronger
  negative signal than Secure Parameter's conceptual-but-unexampled listing. Its own "Deploying
  an SSH Known Hosts Artifact" documentation describes only the Manage Security Material UI
  (*Create* > *Known Hosts (SSH)*, *Browse*/*Add*/*Deploy*), with no REST endpoint mentioned
  anywhere. Corrected from `research_required` to `no_public_api`.

### Certificate-User Mapping — unchanged from prior research: Neo-only, no Cloud Foundry equivalent

Reconfirmed during this pass: SAP's certificate-to-user mapping documentation ("Managing
Certificate-to-User Mappings", "Client Certificate Authentication and Certificate-to-User
Mapping (Inbound)", "Setting Up Inbound HTTP Connections with Certificate-to-User Mapping")
exists only under the **Neo** environment, with no Cloud Foundry equivalent found anywhere in
SAP's published documentation. Since this provider targets Cloud Foundry, this stays
`PublicAPI: false` / `no_public_api`.

### Required role

SAP's Security Content overview page documents the Cloud Foundry role template
`MonitoringDataRead` as required to access this OData API — the same role already documented for
the Message Stores API family. This provider does not manage that role or role collection.

### Whole-keystore operations — reverified, confirmed high-risk, deliberately not implemented

- `POST /api/v1/KeystoreResources` (Import/Back up a Keystore): confirmed contract — `Name`
  (fixed value `"system"`, SAP's documentation states "Currently, only `system` is allowed (SAP
  supports only one keystore)"), `Resource` (base64-encoded JKS/JCEKS keystore file),
  `password` (the keystore/private-key password — SAP documents "All private keys must have the
  same password"). Confirms the exact blast-radius risk this provider's `provider-scope.md` and
  `ROADMAP.md` already flag: a single call can create, update, leave unchanged, or remove many
  keystore entries at once, based on the imported file's contents — entries this provider has no
  way to know are or are not owned by a different Terraform module or administrator. Deliberately
  not implemented, independent of how well the contract is now understood.
- `PUT /api/v1/KeystoreResources('system')?deleteEntries=true` (Trigger Mass Deletion of
  Keystore Entries): confirmed contract, and the operation `sapintegrationsuite_certificate` and
  `sapintegrationsuite_key_pair` both use for Delete — see
  `internal/client/securitycontent/keystore.go`. Body: `{"Aliases":"alias1;alias2;alias3"}`,
  **plain alias text, not hex-encoded** (unlike `KeystoreEntries`' own key predicate). SAP
  documents exact escaping rules with two worked examples: a literal `;` in an alias becomes
  `\;`, and a literal `\` becomes `\\`. This project's own inference (confirmed by reproducing
  both worked examples byte-for-byte in
  `internal/client/securitycontent/keystore_test.go`) is that backslashes must be escaped
  *before* semicolons — escaping in the other order would re-escape the backslash this operation
  itself inserts for a semicolon, corrupting the result; SAP's documentation states both
  individual rules but does not spell out their combined order. Both
  `sapintegrationsuite_certificate` and `sapintegrationsuite_key_pair` call this with exactly the
  one alias they own — never a caller-supplied list — so a Terraform destroy of one resource can
  never be constructed in a way that also submits an unrelated alias.
- `PUT /api/v1/HistoryKeystoreEntries('{Hexalias}')?copy=true&destinationAlias=<...>` (Restore an
  Entry from the SAP Key History Keystore): confirmed contract, a restore/copy operation on
  SAP-managed historical key material, not a generic CRUD entity — no GET was found documented
  for `HistoryKeystoreEntries` either, so even a read-only data source is not implemented.
  Consistent with this provider's stance that history/audit data should never become a mutable
  Terraform resource.
- `GET /api/v1/Keystores('system')?$expand=Entries&$select=LastModifiedTime,Entries` (Get
  Keystore Properties and Entries): confirmed, but not separately implemented as a data source —
  it returns the same entry list `data.sapintegrationsuite_keystore_entries` already exposes,
  plus a keystore-level `LastModifiedTime` this project judged not worth a second, overlapping
  data source for.

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
