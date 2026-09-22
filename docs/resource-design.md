# Resource Design

For every resource, this document records the Terraform-suitability check from
`provider-scope.md`/rule set §15 (desired state, identity, read-back, create, update, delete,
drift, import) before committing to a schema.

## `sapintegrationsuite_integration_package`

- **Purpose**: manage a Cloud Integration content package (the container for flows, mappings,
  script collections).
- **SAP object**: `IntegrationPackages` (Integration Content API, OData V2).
- **Desired state**: yes — name, description, short text, version are stable, user-set fields.
- **Identity**: the SAP-assigned `Id` (a business key chosen at creation time) — stable, used
  as the Terraform ID directly.
- **Read-back**: `GET /api/v1/IntegrationPackages('{Id}')` returns the full object.
- **Create**: `POST /api/v1/IntegrationPackages`.
- **Update**: SAP does not support in-place metadata update for packages created via the API
  in every field; the provider only updates the fields SAP's API accepts and uses
  `RequiresReplace()` on `id` (immutable after creation).
- **Delete**: `DELETE /api/v1/IntegrationPackages('{Id}')`. A `404` on delete is treated as
  success.
- **Drift detection**: `Read` always re-fetches; renames/description edits made in the UI are
  visible on the next `terraform plan`.
- **Import**: `terraform import sapintegrationsuite_integration_package.example <Id>`.
- **Known limitation**: SAP-shipped/content-catalog packages are read-only; attempting to
  manage one as a Terraform resource will surface the SAP API's own error.

## `sapintegrationsuite_integration_flow`

- **Purpose**: manage the design-time content of a single integration flow.
- **SAP object**: `IntegrationDesigntimeArtifacts`, scoped under a package.
- **Desired state**: yes, but the "value" is binary content (a zip of the iFlow project), not
  a set of scalar fields. The provider models the artifact through a file plus a content
  hash, following the file-based pattern in §36 of the project brief:

  ```hcl
  resource "sapintegrationsuite_integration_flow" "metering" {
    package_id = sapintegrationsuite_integration_package.utilities.id
    flow_id    = "metering"
    name       = "Metering"

    content      = "${path.module}/iflows/metering.zip"
    content_hash = filesha256("${path.module}/iflows/metering.zip")
  }
  ```

- **Identity**: composite `<package_id>/<flow_id>`. Both are user-chosen business keys, stable
  across versions.
- **Read-back**: `GET IntegrationPackages('{package_id}')/IntegrationDesigntimeArtifacts` for
  metadata (name, version); the zip payload itself is only re-downloaded when
  `content_hash` indicates drift, to avoid unnecessary large downloads.
- **Create**: `POST IntegrationDesigntimeArtifacts` with the zip content base64-encoded.
- **Update**: SAP's design-time API is version-based, not in-place: an update creates a new
  design-time version of the same `flow_id`. The provider treats this as a normal Terraform
  update (no replace) because identity (`package_id`/`flow_id`) does not change.
- **Delete**: `DELETE IntegrationDesigntimeArtifacts(Id='{flow_id}',Version='{version}')`.
- **Drift detection**: content hash comparison against the last-read artifact; metadata
  (name) compared directly.
- **Import**: `terraform import sapintegrationsuite_integration_flow.metering UTILITIES/metering`.
- **Security**: content is handled via streaming/size-bounded reads to avoid zip-slip/oversized
  decompression issues on any future "read back and inspect" feature (see `SECURITY.md`).
- **Note**: deploying the flow is a **separate** resource (see below) — a design-time content
  change never implicitly deploys.

## `sapintegrationsuite_integration_flow_deployment`

- **Purpose**: express the desired *runtime* deployment state of an integration flow,
  independent of its design-time content lifecycle.
- **SAP object**: the `DeployIntegrationDesigntimeArtifact` action plus `IntegrationRuntimeArtifacts`
  for status.
- **Desired state**: yes — "this version of this flow should be deployed". `flow_version` is
  a Required input (typically wired to `sapintegrationsuite_integration_flow.<name>.version`)
  precisely so that a new design-time version produces a plannable diff on this resource; the
  design-time and deployment resources otherwise share no attribute that would change when
  content changes, so without `flow_version` a redeploy would never be triggered.
- **Identity**: the flow's `<package_id>/<flow_id>` (a tenant only ever has one active runtime
  deployment per flow id).
- **Create/Update**: `POST DeployIntegrationDesigntimeArtifact?Id='{flow_id}'&Version='{flow_version}'`,
  then poll `IntegrationRuntimeArtifacts(Id='{flow_id}')` until `Status` is `STARTED` or
  `ERROR`, using context-based exponential backoff with jitter and a configurable timeout —
  never a fixed sleep.
- **Delete**: `DELETE IntegrationRuntimeArtifacts(Id='{flow_id}')` (undeploy). `404` is success.
- **Drift detection**: `Read` writes the actually-deployed `Version` it observes back into
  `flow_version` (a Required, non-Computed attribute). Since Terraform diffs that against the
  desired value in configuration, a flow redeployed to a different version or left stale
  outside Terraform shows up as a plannable change on the next refresh.
- **Import**: `terraform import sapintegrationsuite_integration_flow_deployment.metering UTILITIES/metering`.
- **Async model**:

  ```
  Create/Update -> POST Deploy -> poll GET status -> STARTED | ERROR
  ```

## `sapintegrationsuite_access_policy`

- **Purpose**: manage an access policy that restricts which roles can operate on which
  Cloud Integration artifacts.
- **SAP object**: `AccessPolicies` (Security Content API, OData V2 — SAP documents this
  entity as supporting both read and write operations).
- **Desired state**: yes — `role_name`, `description`.
- **Identity**: SAP-assigned technical ID returned on create; `role_name` is unique per tenant
  but the provider uses the technical ID as the stable Terraform ID once known, consistent
  with the rest of the Security Content API.
- **Create**: `POST AccessPolicies` with `RoleName`/`Description`.
- **Update**: `PATCH AccessPolicies('{id}')` for `description` (PATCH, not PUT, so that fields
  outside the Terraform schema are left untouched); `role_name` uses `RequiresReplace()` since
  SAP does not document renaming a policy's role.
- **Delete**: `DELETE AccessPolicies('{id}')`.
- **Drift detection**: `Read` re-fetches the policy and its reconciliation/runtime status.
- **Import**: `terraform import sapintegrationsuite_access_policy.utilities <id>`.
- **Runtime awareness (partially implemented)**: `reconciliation_status` is surfaced as a
  computed attribute whenever the API reports one, but Create/Update currently return as soon
  as the policy itself is created/updated — they do **not** poll until reconciliation reaches a
  terminal state. This is deliberate for now: the exact status enum (values, terminal states,
  per-runtime shape for Integration Cell vs. Edge Integration Cell) has not been confirmed
  against a live tenant, and polling on an unverified enum risks either hanging on a status
  value we don't recognize as terminal or returning "success" prematurely. Implementing the
  poll is tracked as follow-up work once the status model is confirmed; see
  `docs/sap-api-references.md`.

## `sapintegrationsuite_access_policy_reference`

- **Purpose**: manage a single artifact reference (artifact type + match condition) attached
  to an access policy.
- **Suitability evaluation (§17)**: artifact references have their own server-assigned ID once
  created, their own create/delete lifecycle independent of other references on the same
  policy, and ordering is not documented as significant. This favors a **separate resource**
  over a nested block, because:
  - independent API identity: yes (server-assigned reference ID)
  - independent CRUD: yes
  - import of a single reference: yes, without needing the whole policy's reference set
  - drift detection per reference: possible independently
  - a nested-block model would force replacing the entire reference set on any single change
- **SAP object**: artifact references nested under `AccessPolicies('{id}')`.
- **Identity**: composite `<access_policy_id>/<reference_id>`.
- **Schema**:

  ```hcl
  resource "sapintegrationsuite_access_policy_reference" "utilities_flows" {
    access_policy_id = sapintegrationsuite_access_policy.utilities.id

    artifact_type = "IntegrationFlow"
    attribute     = "Name"
    operator      = "EQUALS"
    value         = "metering"
  }
  ```

  `artifact_type` is validated against the artifact types SAP documents as supported today:
  `IntegrationFlow`, `ODataAPI`, `RestAPI`, `SoapAPI`, `ScriptCollection`, `ValueMapping`,
  `MessageMapping`, `MessageQueue`, `GlobalDataStore`, `GlobalVariable`. No values beyond
  what SAP documents are accepted by the validator.
- **Create/Delete**: `POST`/`DELETE` against the nested collection.
- **Import**: `terraform import sapintegrationsuite_access_policy_reference.utilities_flows <access_policy_id>/<reference_id>`.

## Value Mapping API model

Before designing the Terraform resources, this section separates two distinct SAP concerns
that are easy to conflate: the **design-time artifact** (a versioned container, structurally a
sibling of `IntegrationDesigntimeArtifacts` in the same Integration Content API), and
**individual mapping entries** inside it (managed through a different, entry-oriented set of
OData actions with its own preconditions).

**Confirmed operations** (Integration Content API, OData V2, `CloudIntegrationAPI` package):

| Operation | Method | Purpose |
|---|---|---|
| `ValueMappingDesigntimeArtifacts` | GET | Read artifact metadata (`Id`, `Version`, `Name`, `PackageId`), navigable from `IntegrationPackages` |
| `ValueMappingDesigntimeArtifacts` | POST | Create a new value mapping artifact from uploaded content |
| `ValueMappingDesigntimeArtifactSaveAsVersion` | POST | Save the artifact's content under an explicit, caller-supplied new version identifier |
| `ValueMappingDesigntimeArtifacts(Id=…,Version=…)` | DELETE | Delete the artifact |
| `DeployValueMappingDesigntimeArtifact?Id='…'&Version='…'` | POST | Deploy a specific version to the runtime |
| `IntegrationRuntimeArtifacts(Id=…)` | GET / DELETE | Read deployment status / undeploy — the **same shared runtime-artifacts entity** already used by `sapintegrationsuite_integration_flow_deployment`, not a distinct "ValueMappingRuntimeArtifacts" entity |
| `UpsertValMaps` | POST | Insert/update individual mapping-entry rows inside an *existing* scheme. Reachable secondary sources describe it as query-parameter-only (no request body — sending one reportedly causes a 400), returning 202 with no response body on success. |
| `UpdateDefaultValMap` | POST | Set the default value for a scheme, identified by a `ValMapId` GUID. That GUID is not returned by `UpsertValMaps`'s response (which per the above has no body); the same secondary sources describe retrieving it via a separate `ValueMappingDesigntimeArtifacts` read after the upsert, not from the upsert call itself. |
| `DeleteValMaps` | — | Delete mapping-entry rows (exact granularity unconfirmed) |
| Required roles | — | `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`, `WorkspaceArtifactsDeploy` |

The deploy action's exact name (singular `...Artifact`, not plural `...Artifacts`) was re-checked
this phase after a report that current SAP documentation might use the plural form. Every
reachable secondary source (independent search results summarizing SAP community/blog content)
consistently gives the singular form used above, matching the sibling actions
`DeployIntegrationDesigntimeArtifact` and `DeployMessageMappingDesigntimeArtifact`; no source
found gave the plural form. `help.sap.com`/`api.sap.com`, which would settle this conclusively,
were not reachable from this environment to fetch and read directly (see the Update entry below
for the full list of what was tried and blocked).

**A real, confirmed constraint shapes the whole design**: a value mapping artifact cannot be
saved with zero entries — SAP requires at least one mapping entry to exist before the object
can be created at all. This, together with the entry-level API's own precondition (attempting
to upsert into a source/target agency-identifier combination that doesn't already exist as a
defined scheme returns 404 — the scheme must already exist before values can be added to it),
indicates the entry-level API manages *rows within* a scheme that is itself part of the
artifact's design-time content, not a fully independent object graph creatable purely through
`UpsertValMaps` calls.

## `sapintegrationsuite_value_mapping`

- **Purpose**: manage a Cloud Integration value mapping design-time artifact as a file-based
  content package, the same way `sapintegrationsuite_integration_flow` manages an integration
  flow's ZIP content — the content itself (the scheme definitions and their entries) is
  authored as a unit and uploaded, rather than field-by-field through Terraform.
- **SAP object**: `ValueMappingDesigntimeArtifacts`.
- **Desired state**: yes — `name`, and content identified by a hash, following exactly the
  `content`/`content_hash` pattern already proven for `sapintegrationsuite_integration_flow`.
- **Identity**: composite `<package_id>/<mapping_id>`, both user-chosen business keys.
- **Create**: `POST ValueMappingDesigntimeArtifacts` with `Id`, `Name`, `PackageId`, and
  base64-encoded `ArtifactContent`.
- **Read**: `GET ValueMappingDesigntimeArtifacts(Id='{mapping_id}',Version='active')`.
- **Update — resolved conservatively, not implemented via PUT**: an earlier version of this
  resource called `PUT` against the keyed `(Id, Version)` entity, by analogy with
  `sapintegrationsuite_integration_flow`. That assumption was re-investigated this phase and
  deliberately not kept:
  - `help.sap.com`, `api.sap.com`, `community.sap.com`, `blogs.sap.com`, and every reachable
    mirror/proxy for them (web.archive.org, a Google Translate proxy, a text-extraction proxy)
    were all blocked by this environment's network egress policy, so the exact
    `ValueMappingDesigntimeArtifactSaveAsVersion` request/response shape and any authoritative
    statement about what plain `PUT` does for this entity set could not be fetched and read
    directly.
  - What could be gathered from reachable secondary sources (search-engine summaries of SAP
    community/blog posts, plus a public SAP Knowledge Base Article, 3502529, titled "HTTP/403
    Forbidden response while trying to change Version of the ValueMapping") consistently treats
    changing a value mapping's version as a distinct, separately named, separately gated
    operation (`ValueMappingDesigntimeArtifactSaveAsVersion`, taking a caller-supplied new
    version identifier) — not something that happens implicitly as a side effect of a generic
    `PUT`, the way it does for `IntegrationDesigntimeArtifacts`.
  - An independent third-party OData client built directly against this same API
    (`github.com/lemaiwo/ci-mcp-server`) explicitly disables its generic "update" operation for
    `ValueMappingDesigntimeArtifacts` specifically, while leaving the identical operation enabled
    for `IntegrationDesigntimeArtifacts`, `MessageMappingDesigntimeArtifacts`, and
    `ScriptCollectionDesigntimeArtifacts` — the same family of design-time artifact entity sets.
    That is a deliberate, specific difference, not a gap in that project's coverage.

  None of this is primary-source confirmation, but taken together it points the same direction:
  retaining the `PUT` call would have been shipping a guess this project's own standing rule
  says not to ship, especially given the explicit instruction not to keep an unverified `PUT`
  just because it resembles the Integration Flow API. **Resolution (Option A from the phase's
  conservative-fallback list): `sapintegrationsuite_value_mapping` no longer has an in-place
  Update.** `name`, `content`, and `content_hash` are now all `RequiresReplace`; changing any of
  them makes Terraform create a new value mapping artifact and delete the old one, which only
  relies on `Create` and `Delete` — both independently confirmed, unlike the update path.
  `internal/client/cloudintegration/value_mapping.go` no longer has an `UpdateValueMapping`
  function at all; see the comment there for the full reasoning. Implementing true in-place
  update via `ValueMappingDesigntimeArtifactSaveAsVersion` is deferred to v0.2.x, once its exact
  contract can be confirmed against a live tenant or a reachable primary source — see
  `docs/sap-api-references.md`.
- **Delete**: `DELETE ValueMappingDesigntimeArtifacts(Id='{mapping_id}',Version='active')`.
- **Version**: Computed only, exactly like `sapintegrationsuite_integration_flow` — SAP
  assigns it, Terraform never asks the user to manage a version string, which is also what
  keeps a `terraform apply` with unchanged content from producing a new version.
- **Import**: `terraform import sapintegrationsuite_value_mapping.example UTILITIES/company-codes`.
- **`content_hash` stays explicit, not provider-computed**: re-evaluated this phase, since
  requiring the user to write both `content` and `content_hash` (typically
  `content_hash = filesha256(content)`) could look like redundant manual bookkeeping the
  provider could do itself. Keeping it explicit was kept as the better design:
  - `filesha256()` is a Terraform built-in function evaluated as part of the configuration
    graph, not something the user maintains by hand — in practice this is the same
    `source_hash`/`etag`-style pattern used by file-backed resources across the Terraform
    ecosystem (for example `aws_s3_object`'s `source_hash`), not something unusual to this
    provider.
  - Making `content_hash` fully provider-computed would mean reading the local file from
    inside a custom plan modifier during `terraform plan`, purely to detect drift before any
    API call happens. That is technically possible with the Terraform Plugin Framework, but it
    reads a local file as a side effect of planning outside the declarative config-graph value
    the plan is otherwise built from — a bigger change to this provider's planning model than
    this phase's scope (resolving the Value Mapping API contract) justifies.
  - `sapintegrationsuite_integration_flow` already ships the identical explicit
    `content`/`content_hash` pattern; changing it only for `sapintegrationsuite_value_mapping`
    would make the two file-based resources behave inconsistently for no API-contract reason.
    Any change here belongs in its own cross-cutting pass covering both resources, not folded
    into a Value Mapping-focused phase.
  - Conclusion: `content_hash` remains an explicit, user-supplied attribute for both resources.
- **Drift detection**: `Read` re-fetches metadata on every refresh, exactly like
  `sapintegrationsuite_integration_flow`; the same import limitation applies to `content` (see
  that resource's entry above) since SAP does not return a local file path for existing
  content. Right after import, with `content`/`content_hash` left unset in configuration,
  `terraform plan` shows no changes — `UseStateForUnknown` keeps the planned value equal to the
  (null) imported state, so `RequiresReplace` sees no diff and does not fire. The first apply
  that *does* supply `content`/`content_hash` to bring an imported artifact's content under
  management is expected to replace the resource (see the Update entry above), not update it in
  place — this is called out explicitly in `examples/brownfield/main.tf`.

## `sapintegrationsuite_value_mapping_deployment`

- **Purpose**: express the desired runtime deployment state of a value mapping, independent of
  its design-time content, mirroring `sapintegrationsuite_integration_flow_deployment` exactly.
- **SAP object**: the `DeployValueMappingDesigntimeArtifact` action plus the shared
  `IntegrationRuntimeArtifacts` entity for status/undeploy (see the table above — this is not
  a separate runtime entity per artifact type).
- **Schema and lifecycle**: identical shape to `sapintegrationsuite_integration_flow_deployment`
  — a required `mapping_version` input (normally wired to
  `sapintegrationsuite_value_mapping.<name>.version`) that both triggers redeployment when the
  design-time version changes and, once `Read` writes the actually-deployed version back into
  it, surfaces drift from an out-of-band redeploy or undeploy. Uses the same context-aware
  polling with exponential backoff and jitter, never a fixed sleep.
- **Shared runtime abstraction, re-reviewed this phase**: `internal/client/cloudintegration/runtime_artifact.go`
  and `internal/provider/runtime_deployment.go` are shared, unmodified, between
  `sapintegrationsuite_integration_flow_deployment` and this resource. Re-checked specifically
  for whether that sharing forces genuinely different artifact types through one abstraction —
  it does not: SAP's own documentation of the "Runtime Status API" describes
  `IntegrationRuntimeArtifacts` and the deploy/undeploy/status model as covering "currently
  deployed integration artifacts" generically, not per design-time artifact type, and the same
  third-party OData client referenced under Update above (`ci-mcp-server`) independently models
  `IntegrationRuntimeArtifacts` as a single generic entity set, not one per artifact type.
  Deliberately, the sharing stops at runtime status/undeploy: `DeployValueMappingDesigntimeArtifact`
  is its own action (distinct from `DeployIntegrationDesigntimeArtifact`, just following the
  same naming and query-parameter convention), and nothing about design-time Update/versioning
  is shared — value mapping and integration flow diverge there (see Update above), and the code
  does not pretend otherwise. "Shared polling is good only where semantics are genuinely
  shared" continues to hold for this abstraction.
- **Import**: `terraform import sapintegrationsuite_value_mapping_deployment.example UTILITIES/company-codes`.

## Deliberately not implemented this phase: entry-level management

`sapintegrationsuite_value_mapping_entry` (or any resource wrapping `UpsertValMaps`,
`UpdateDefaultValMap`, or `DeleteValMaps` directly) is **not implemented**. This was a genuine
evaluation, not a default:

- **Independent identity**: plausible — `UpdateDefaultValMap` addresses a scheme through a
  `ValMapId` GUID, suggesting schemes (and possibly entries) do have server-assigned
  identities of their own.
- **What blocks implementation**: the exact request/response payload shapes for all three
  operations, the precise granularity of `DeleteValMaps` (a single entry, a whole scheme, or a
  bulk selection), and how `ValMapId` is obtained from a `ValueMappingDesigntimeArtifacts` read
  could not be confirmed against any reachable primary source. Guessing these would violate
  this project's standing rule against inventing wire-format details, and would risk exactly
  the destructive-ambiguity problem the ownership-semantics question warns about: an
  under-specified "Terraform owns all entries" model could delete rows Terraform never created
  if the delete operation's actual scope turns out to be broader than assumed.
  - a confirmed precondition (entries can only be upserted into an *already-existing* scheme,
    confirmed by the documented 404-on-nonexistent-combination behavior) means entry management
    cannot stand alone as a resource anyway — it would depend on scheme structure that is
    itself part of the artifact's uploaded content, not a separate creatable object.
- **Path forward**: once the exact API contract is confirmed (ideally against a live tenant's
  `$metadata` and an actual request/response trace), the ownership model most consistent with
  the evidence gathered so far is likely closer to Option B (separate
  `sapintegrationsuite_value_mapping_entry` resources with a composite ID derived from the
  artifact plus the source/target agency-identifier tuple) than Option A (nesting, which would
  force replacing the entire entry set on any single change) — but this is not yet a design
  decision, only a direction for the next investigation.

## Message Mapping API model

A reusable Message Mapping is a **package-level design-time artifact**, structurally the same
kind of object as an integration flow or a value mapping — not the inline/local message mapping
step that can be configured directly inside an integration flow without ever becoming a
standalone artifact. The two are easy to conflate because SAP's own UI uses the same term
"Message Mapping" for both:

```
Integration Package
│
├── Message Mapping Artifact (MessageMappingDesigntimeArtifacts)
│      │
│      └── reusable by Integration Flows via a message mapping flow step
│
└── Integration Flow
       │
       └── may reference a Message Mapping Artifact, or define an inline
           mapping that never becomes a separate artifact
```

This feature is only about the first kind — the reusable, package-level artifact reachable
through `MessageMappingDesigntimeArtifacts`. It does not model, own, or manage inline mapping
configuration embedded directly in an integration flow's own content, which stays entirely
inside that integration flow's `content`/`content_hash` as far as this provider is concerned.

**Confirmed via SAP's own documentation** (SAP Help Portal, read through the `SAP-docs`
GitHub organization's markdown mirror of the Cloud Integration documentation — the primary,
official source, not a blog or forum): creating a message mapping artifact requires a package
context, a technical ID, a display name, and optional description; the artifact's content is a
mapping definition (`*.mmap`) file, uploaded as (or bundled inside) a ZIP archive; and the
artifact must be deployed before any integration flow that references it can use it — there is
no automatic deployment of a referenced message mapping when the referencing integration flow
is deployed. That last point matters for this provider's ownership model (see the deployment
resource below): a message mapping's deployment is independent of, and not implicitly triggered
by, anything the flows that reference it do.

**Confirmed operations** (Integration Content API, OData V2, `CloudIntegrationAPI` package):

| Operation | Method | Purpose |
|---|---|---|
| `MessageMappingDesigntimeArtifacts` | GET | Read artifact metadata (`Id`, `Version`, `Name`, `PackageId`), navigable from `IntegrationPackages` |
| `MessageMappingDesigntimeArtifacts` | POST | Create a new message mapping artifact from uploaded content |
| `MessageMappingDesigntimeArtifacts(Id=…,Version=…)` | PUT | Update an existing artifact's content, creating a new design-time version — see Update below |
| `MessageMappingDesigntimeArtifacts(Id=…,Version=…)` | DELETE | Delete the artifact — see Delete below for what "delete" actually removes |
| `DeployMessageMappingDesigntimeArtifact?Id='…'&Version='…'` | POST | Deploy a specific version to the runtime (singular action name, matching the sibling actions for integration flows and value mappings) |
| `IntegrationRuntimeArtifacts(Id=…)` | GET / DELETE | Read deployment status / undeploy — the same shared runtime-artifacts entity already used by `sapintegrationsuite_integration_flow_deployment` and `sapintegrationsuite_value_mapping_deployment`, confirmed applicable here too (see the deployment resource entry below) |
| `MessageMappingDesigntimeArtifactSaveAsVersion?Id='…'&SaveAsVersion='…'` | POST | Save the current content under an explicit, caller-supplied version string, as a named milestone alongside the version SAP assigns automatically on every `PUT` — not used by this provider (see Update below) |
| Required roles | — | `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`, `WorkspaceArtifactsDeploy` (same Integration Content API roles already required for integration flows and value mappings) |

**`SaveAsVersion` is a universal action across this API family, not a Value-Mapping-specific
concept.** While re-investigating the Value Mapping API contract in an earlier phase, this
project treated `ValueMappingDesigntimeArtifactSaveAsVersion` as if it might be Value Mapping's
*only* real update path, precisely because a plain `PUT` for that entity set could not be
confirmed to work and a third-party OData client explicitly disabled generic update for it. This
phase's research shows the actual shape of `SaveAsVersion` more clearly: it exists as
`IntegrationDesigntimeArtifactSaveAsVersion?Id='…'&SaveAsVersion='…'` for integration flows too
— an entity set where `PUT`-based update is independently confirmed and already implemented — so
`PUT` and `SaveAsVersion` are evidently not mutually exclusive alternatives in general; they
coexist, with `SaveAsVersion` letting a caller pin an explicit version string as a named
milestone on top of whatever version `PUT` assigns automatically. This does not change any
conclusion this project reached about Value Mapping specifically (that remains governed by its
own entity-specific evidence, documented in `docs/sap-api-references.md`), but it does mean the
mere existence of a `SaveAsVersion` action for Message Mapping is not, by itself, a reason to
avoid `PUT` here. What actually decides Message Mapping's Update model is entity-specific
evidence for *this* entity set — see Update below.

## `sapintegrationsuite_message_mapping`

- **Purpose**: manage the design-time content of a reusable Cloud Integration message mapping
  artifact, uploaded from a local content file, the same way
  `sapintegrationsuite_integration_flow` manages an integration flow's ZIP content.
- **SAP object**: `MessageMappingDesigntimeArtifacts`.
- **Identity**: composite `<package_id>/<mapping_id>`, both user-chosen business keys, exactly
  mirroring `sapintegrationsuite_integration_flow` and `sapintegrationsuite_value_mapping`.
- **Create**: `POST MessageMappingDesigntimeArtifacts` with `Id`, `Name`, `PackageId`, and
  base64-encoded `ArtifactContent`. No SAP-documented minimum-content precondition was found for
  message mapping (unlike value mapping's confirmed "at least one entry" requirement) — none is
  enforced client-side either, consistent with this provider's rule of not inventing
  constraints SAP does not document.
- **Read**: `GET MessageMappingDesigntimeArtifacts(Id='{mapping_id}',Version='active')`. The
  `Version='active'` alias is the same one already confirmed and tested for
  `sapintegrationsuite_integration_flow` and `sapintegrationsuite_value_mapping`, on the same
  entity family; not re-derived from scratch for this entity, but not assumed to be
  automatically valid for every future entity in this family either.
- **Update — implemented via `PUT`, on different grounds than Value Mapping**: unlike Value
  Mapping, this phase found positive, entity-specific evidence that `PUT` is the right mechanism
  here, not just an analogy with integration flows:
  - `MessageMappingDesigntimeArtifacts` shares the exact same `(Id, Version)` composite-key
    shape as `IntegrationDesigntimeArtifacts`, whose `PUT`-based, version-creating update is
    independently confirmed and already implemented.
  - An independent third-party OData client built directly against this API
    (`github.com/lemaiwo/ci-mcp-server`, the same one whose configuration was used as
    corroborating evidence against Value Mapping's `PUT`) explicitly enables its generic
    "update" operation for `MessageMappingDesigntimeArtifacts`, the same as it does for
    `IntegrationDesigntimeArtifacts` and `ScriptCollectionDesigntimeArtifacts` — and unlike
    `ValueMappingDesigntimeArtifacts`, where that same tool explicitly disables it. That is a
    deliberate difference the tool's author drew between these entity sets, not a gap in
    coverage.
  - No SAP Knowledge Base Article or other evidence of a documented problem with changing a
    message mapping's version via `PUT` was found (unlike Value Mapping, where KBA 3502529
    documents exactly that problem for that entity set specifically).
  - Following SAP's own `IntegrationDesigntimeArtifacts` behavior, `PUT` here creates a new
    design-time version of the same artifact ID; this provider treats that as a normal Terraform
    Update (no replace), because identity (`package_id`/`mapping_id`) does not change — the
    same design already proven for `sapintegrationsuite_integration_flow`.
  - `MessageMappingDesigntimeArtifactSaveAsVersion` exists (see the API model above) but is not
    used: this provider does not ask the user to manage an explicit version string, so there is
    nothing for it to do here that `PUT`'s automatic versioning does not already cover.
- **Delete**: `DELETE MessageMappingDesigntimeArtifacts(Id='{mapping_id}',Version='active')`.
  Whether this removes only the active version or every version of the artifact was not
  confirmed against a primary source (the same open question already flagged for
  `sapintegrationsuite_value_mapping`'s Delete) — documented as an open item rather than
  asserted as "removes all versions".
- **Version**: Computed only. SAP assigns it on every `PUT`; Terraform never asks the user to
  manage a version string, which is also what keeps a `terraform apply` with unchanged content
  from producing a new version (idempotent apply — see Terraform version semantics below).
- **Import**: `terraform import sapintegrationsuite_message_mapping.example UTILITIES/customer-mapping`.
- **Drift detection**: `Read` re-fetches metadata on every refresh. `content`/`content_hash`
  follow the exact same explicit, user-supplied, `RequiresReplace`-free pattern already
  documented and re-evaluated for `sapintegrationsuite_value_mapping` (see that resource's
  entry above for the content-hash design reasoning, which applies here unchanged); the same
  import limitation applies since SAP does not return a local file path for existing content.
- **Content format**: transported opaquely as a ZIP archive (base64-encoded `ArtifactContent`),
  exactly like `sapintegrationsuite_integration_flow` and `sapintegrationsuite_value_mapping`.
  This provider does not parse, validate, or interpret what is inside the archive — not the
  `.mmap` mapping definition, and not any XSD/WSDL/EDMX/Swagger-OpenAPI schema files SAP's
  mapping editor lets a message mapping reference for its source/target message structures. It
  only transports and manages the artifact as a unit, the same boundary already established for
  the other file-based design-time resources.

## Terraform version semantics for Message Mapping

Selected **Model A** from this phase's three candidate designs (content changes update the
design-time artifact in place; the provider does not expose or require a caller-supplied version
string; a separate deployment resource references whatever version the artifact resource last
produced) over:

- Model B (an explicit `desired_version` Terraform input) — rejected because nothing in the
  confirmed API contract requires or even exposes a caller-chosen version number for a normal
  content update; `SaveAsVersion`'s caller-supplied version string is an optional, separate
  action this provider does not use (see Update above), not a required part of the update path.
- Model C (draft vs. published version as separate resource concerns) — rejected because no
  SAP documentation surfaced a draft/published distinction for message mapping artifacts
  independent of the same `Version='active'` alias already used uniformly across this API
  family; inventing that distinction without evidence would violate this project's standing
  rule against guessing wire/state semantics.

Model A is also exactly what `sapintegrationsuite_integration_flow` already implements, so this
is a consistency choice as well as an evidence-based one: an unchanged `terraform apply` does not
create a new SAP version, because content-hash comparison against the last-read artifact is what
decides whether `Update` is even called, and the `version` attribute is Computed-only so
Terraform itself never treats a version-string mismatch as configuration drift.

## `sapintegrationsuite_message_mapping_deployment`

- **Purpose**: express the desired runtime deployment state of a message mapping, independent
  of its design-time content, mirroring `sapintegrationsuite_integration_flow_deployment` and
  `sapintegrationsuite_value_mapping_deployment` exactly.
- **SAP object**: the `DeployMessageMappingDesigntimeArtifact` action plus the shared
  `IntegrationRuntimeArtifacts` entity for status/undeploy.
- **Why `IntegrationRuntimeArtifacts` and not `BuildAndDeployStatus`**: SAP's Integration
  Content API also exposes a `BuildAndDeployStatus(TaskId='…')` entity, which this phase
  investigated specifically rather than assuming it applies here. The evidence found ties
  `BuildAndDeployStatus` to a different artifact family's build-then-deploy pipeline (OData API
  artifacts, which SAP documents as needing to be built before they can be deployed, keyed by a
  `TaskId` a build operation returns — not by the artifact's own `Id`), not to
  `MessageMappingDesigntimeArtifacts`. By contrast, SAP's own Runtime Status API documentation
  describes `IntegrationRuntimeArtifacts` as covering "currently deployed integration
  artifacts" generally, and secondary sources specifically describe message mappings as
  existing as runtime artifacts inside deployed runtime packages alongside integration flows,
  script collections, and adapters — monitored through that same shared entity. `Deploy` for a
  message mapping does not return a `TaskId` the way a `BuildAndDeployStatus`-fronted deploy
  would; it follows the same fire-and-poll-`IntegrationRuntimeArtifacts` shape already
  implemented for integration flows and value mappings. `runtime_artifact.go` and
  `runtime_deployment.go` are reused unmodified, since the semantics genuinely match — the same
  standard already applied when this abstraction was reviewed for value mapping.
- **Schema and lifecycle**: identical shape to `sapintegrationsuite_value_mapping_deployment` —
  a required `mapping_version` input (normally wired to
  `sapintegrationsuite_message_mapping.<name>.version`) that both triggers redeployment when the
  design-time version changes and, once `Read` writes the actually-deployed version back into
  it, surfaces drift from an out-of-band redeploy or undeploy. Uses the same context-aware
  polling with exponential backoff and jitter, never a fixed sleep.
- **No hidden coupling to referencing integration flows**: confirmed by SAP's own documentation
  that deploying an integration flow does not automatically deploy a message mapping it
  references — this provider does the same: `sapintegrationsuite_message_mapping_deployment` is
  a resource in its own right that a configuration must explicitly create, exactly mirroring
  SAP's own behavior rather than adding automatic-deployment behavior SAP itself does not
  provide. The message mapping resource does not scan, own, or modify integration flow content
  that references it; ownership of that reference stays entirely with whichever integration
  flow's content contains it.
- **Import**: `terraform import sapintegrationsuite_message_mapping_deployment.example UTILITIES/customer-mapping`.

## `data.sapintegrationsuite_message_mapping`

- **Purpose**: read-only lookup of an existing message mapping's metadata, mirroring
  `data.sapintegrationsuite_value_mapping` exactly — useful for brownfield adoption or for a
  configuration that wants to reference metadata/version of a mapping without owning it.
- **SAP object**: `GET MessageMappingDesigntimeArtifacts(Id='{mapping_id}',Version='active')`,
  the same read path the resource uses.
- Implemented because the read semantics are exactly as stable as
  `sapintegrationsuite_value_mapping`'s — not added merely for symmetry with every resource
  having a matching data source.

## Script Collection API model

A Script Collection is a bundle of reusable Groovy/JavaScript scripts, created within an
integration package so the same scripts can be shared across any number of integration flows,
mirroring the same reusable-artifact pattern already implemented for Message Mapping. Confirmed
via SAP's own documentation (again read through the `SAP-docs` GitHub organization's markdown
mirror, a legitimate primary source despite `help.sap.com` itself being blocked): a script
collection's technical ID must be unique for the entire tenant (not just the package), and its
description is capped at 120 characters — both are documented constraints, not this provider's
own invention, and neither is enforced client-side, consistent with this project's rule against
inventing constraints SAP does not document.

**Confirmed operations** (Integration Content API, OData V2, `CloudIntegrationAPI` package):

| Operation | Method | Purpose |
|---|---|---|
| `ScriptCollectionDesigntimeArtifacts` | GET | Read artifact metadata (`Id`, `Version`, `Name`, `PackageId`) |
| `ScriptCollectionDesigntimeArtifacts` | POST | Create a new script collection artifact from uploaded content |
| `ScriptCollectionDesigntimeArtifacts(Id=…,Version=…)` | PUT | Update an existing artifact's content, creating a new design-time version |
| `ScriptCollectionDesigntimeArtifacts(Id=…,Version=…)` | DELETE | Delete the artifact — scope unconfirmed, same open item as every other design-time artifact in this family |
| `DeployScriptCollectionDesigntimeArtifact?Id='…'&Version='…'` | POST | Deploy a specific version to the runtime (singular action name, confirmed via SAP's own documentation, matching the sibling actions for every other design-time artifact type) |
| `IntegrationRuntimeArtifacts(Id=…)` | GET / DELETE | Read deployment status / undeploy — the same shared runtime-artifacts entity already used by every other `*_deployment` resource in this provider; script collections are documented as existing as runtime artifacts alongside integration flows, value mappings, and message mappings inside the same deployed runtime packages |
| Required roles | — | `WorkspacePackagesConfigure`, `WorkspacePackagesEdit`, `WorkspaceArtifactsDeploy` (same Integration Content API roles already required for every other design-time artifact type) |

**Update — implemented via `PUT`, same grounds as Message Mapping.** `ScriptCollectionDesigntimeArtifacts`
shares the exact `(Id, Version)` composite-key shape as `IntegrationDesigntimeArtifacts` and
`MessageMappingDesigntimeArtifacts`, both of which have confirmed, version-creating `PUT`
behavior. The same independent third-party OData client (`github.com/lemaiwo/ci-mcp-server`)
referenced when Message Mapping's Update model was decided explicitly enables generic update for
`ScriptCollectionDesigntimeArtifacts` too — the same as `IntegrationDesigntimeArtifacts` and
`MessageMappingDesigntimeArtifacts`, and unlike `ValueMappingDesigntimeArtifacts`, where it is
disabled. No SAP Knowledge Base Article or other evidence of a documented `PUT` problem for this
entity set was found. `PUT` here is therefore treated as a normal Terraform Update (no replace),
identical to `sapintegrationsuite_message_mapping` and `sapintegrationsuite_integration_flow`.

**Content format**: transported opaquely as a ZIP archive (base64-encoded `ArtifactContent`),
the same convention already confirmed for every other file-based design-time resource in this
provider (Integration Flow, Value Mapping, Message Mapping) — SAP's own Web IDE export/import
mechanism for design-time artifacts uses this ZIP+`ArtifactContent` shape uniformly across the
whole `CloudIntegrationAPI` package, not something inferred solely from Integration Flow. This
provider does not parse, validate, or execute the Groovy/JavaScript scripts inside the archive —
only transports and manages the artifact as a unit, the same boundary already established for
the other file-based resources.

## `sapintegrationsuite_script_collection`

- **Purpose**: manage the design-time content of a reusable Cloud Integration script collection,
  uploaded from a local content file, the same way `sapintegrationsuite_message_mapping` manages
  a message mapping artifact's ZIP content.
- **SAP object**: `ScriptCollectionDesigntimeArtifacts`.
- **Identity**: composite `<package_id>/<script_collection_id>`, mirroring every other
  file-based design-time resource in this provider. SAP additionally documents the technical ID
  as unique across the whole tenant, not just the package — this provider does not need to
  enforce that itself, since a duplicate ID is something SAP's own Create call would reject.
- **Create**: `POST ScriptCollectionDesigntimeArtifacts` with `Id`, `Name`, `PackageId`, and
  base64-encoded `ArtifactContent`. No SAP-documented minimum-content precondition was found
  (unlike value mapping's confirmed "at least one entry" requirement).
- **Read**: `GET ScriptCollectionDesigntimeArtifacts(Id='{script_collection_id}',Version='active')`,
  the same `Version='active'` alias already confirmed for every other entity in this family.
- **Update**: `PUT` against the keyed `(Id, Version)` entity — see the Update entry in the API
  model above for the full reasoning. Treated as a normal Terraform update (no replace), since
  identity does not change.
- **Delete**: `DELETE ScriptCollectionDesigntimeArtifacts(Id='{script_collection_id}',Version='active')`.
  Whether this removes only the active version or every version of the artifact is unconfirmed
  against a primary source, the same open item already flagged for every sibling resource.
- **Version**: Computed only. SAP assigns it on every `PUT`; an unchanged `terraform apply` does
  not produce a new version, for the same reason already established for the other file-based
  resources (content-hash comparison decides whether Update is even called).
- **Import**: `terraform import sapintegrationsuite_script_collection.example UTILITIES/shared-scripts`.
- **Drift detection**: `Read` re-fetches metadata on every refresh; `content`/`content_hash`
  follow the same explicit, user-supplied pattern already documented and re-evaluated for
  `sapintegrationsuite_value_mapping` and `sapintegrationsuite_message_mapping` — see those
  resources' entries for the content-hash design reasoning, which applies here unchanged.

## `sapintegrationsuite_script_collection_deployment`

- **Purpose**: express the desired runtime deployment state of a script collection, independent
  of its design-time content, mirroring `sapintegrationsuite_message_mapping_deployment` exactly.
- **SAP object**: the `DeployScriptCollectionDesigntimeArtifact` action plus the shared
  `IntegrationRuntimeArtifacts` entity for status/undeploy.
- **Schema and lifecycle**: identical shape to `sapintegrationsuite_message_mapping_deployment`
  — a required `script_collection_version` input that both triggers redeployment when the
  design-time version changes and, once `Read` writes the actually-deployed version back into
  it, surfaces drift from an out-of-band redeploy or undeploy. Uses the same context-aware
  polling with exponential backoff and jitter, never a fixed sleep. `runtime_artifact.go` and
  `runtime_deployment.go` are reused unmodified, since script collections are documented as
  sharing the same runtime-artifact model as every other design-time artifact type here.
- **No hidden coupling to referencing integration flows**: a script collection's deployment is
  a resource in its own right that a configuration must explicitly create; this provider does
  not scan or modify integration flow content to manage that reference, the same ownership
  boundary already established for message mapping.
- **Import**: `terraform import sapintegrationsuite_script_collection_deployment.example UTILITIES/shared-scripts`.

## `data.sapintegrationsuite_script_collection`

- **Purpose**: read-only lookup of an existing script collection's metadata, mirroring
  `data.sapintegrationsuite_message_mapping` exactly.
- **SAP object**: `GET ScriptCollectionDesigntimeArtifacts(Id='{script_collection_id}',Version='active')`.

## Deferred to v0.2.x and later

`sapintegrationsuite_capability`, `sapintegrationsuite_api_artifact`,
`sapintegrationsuite_api_artifact_deployment`, message mapping entry-level or dependent-resource
management (schema files referenced by a mapping are managed as part of the opaque content
archive, not as separate Terraform resources), value mapping entry-level management (see above),
security material resources (user credentials, OAuth credentials, keystore), and Partner
Directory resources are designed at the API level in `api-capability-matrix.md` but
intentionally not implemented yet, to keep each release small and high quality (see
`ROADMAP.md`).
