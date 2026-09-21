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
- **Desired state**: yes — "this version of this flow should be deployed".
- **Identity**: the flow's `<package_id>/<flow_id>` (a tenant only ever has one active runtime
  deployment per flow id).
- **Create/Update**: `POST DeployIntegrationDesigntimeArtifact?Id='{flow_id}'&Version='{version}'`,
  then poll `IntegrationRuntimeArtifacts(Id='{flow_id}')` until `Status` is `STARTED` or
  `ERROR`, using context-based exponential backoff with jitter and a configurable timeout —
  never a fixed sleep.
- **Delete**: `DELETE IntegrationRuntimeArtifacts(Id='{flow_id}')` (undeploy). `404` is success.
- **Drift detection**: `Read` compares the currently deployed `Version` against the version
  Terraform expects; a flow undeployed or redeployed out-of-band is detected on refresh.
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
- **Update**: `PUT`/`MERGE AccessPolicies('{id}')` for `description`; `role_name` uses
  `RequiresReplace()` if SAP does not support renaming (to be confirmed against the tenant's
  live `$metadata` — see the caveat in `sap-api-references.md`).
- **Delete**: `DELETE AccessPolicies('{id}')`.
- **Drift detection**: `Read` re-fetches the policy and its reconciliation/runtime status.
- **Import**: `terraform import sapintegrationsuite_access_policy.utilities <id>`.
- **Runtime awareness**: where the API reports per-runtime (Integration Cell / Edge
  Integration Cell) replication/reconciliation status values (for example `PENDING`,
  `SUCCESS`, `FAILED`), the provider surfaces them as computed attributes and, on
  create/update, polls until a terminal state is reached rather than returning immediately.

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

## Deferred to v0.2.x and later

`sapintegrationsuite_capability`, `sapintegrationsuite_api_artifact`,
`sapintegrationsuite_api_artifact_deployment`, value mapping / script collection / message
mapping resources, security material resources (user credentials, OAuth credentials,
keystore), and Partner Directory resources are designed at the API level in
`api-capability-matrix.md` but intentionally not implemented in v0.1.0, to keep the initial
release small and high quality (see `ROADMAP.md`).
