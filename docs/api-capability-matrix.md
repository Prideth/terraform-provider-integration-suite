# API Capability Matrix

Scope: SAP Cloud Integration's public "Integration Content", "Security Content", and
"Partner Directory" OData V2 APIs (published under the `CloudIntegrationAPI` package on
SAP Business Accelerator Hub), plus the Access Policy API. A Terraform column of "Resource"
is only used where Create, Read, and a stable identity are all realistic; "Unsupported" is
used whenever a Terraform resource would not be able to fulfil its contract (see
`resource-design.md` for the full suitability check per object).

| Domain | Object | Runtime | Public API | Protocol | GET | CREATE | UPDATE | DELETE | DEPLOY | Terraform |
|---|---|---|---|---|---|---|---|---|---|---|
| Integration Content | Integration Package | Design-time | Yes | OData V2 | Yes | Yes | Partial (metadata only; SAP-provided packages are read-only) | Yes | N/A | Resource + Data Source |
| Integration Content | Integration Flow (Designtime Artifact) | Design-time | Yes | OData V2 | Yes | Yes (new version) | Yes (new version, versioned not in-place) | Yes | via `DeployIntegrationDesigntimeArtifact` action | Resource + Data Source |
| Integration Content | Integration Runtime Artifact | Runtime | Yes | OData V2 | Yes | N/A (created by Deploy) | N/A | Yes (undeploy) | N/A | Declarative Deployment Resource |
| Integration Content | Value Mapping (Designtime Artifact) | Design-time | Yes | OData V2 | Yes | Yes | Yes, via `PUT` — open question, see `sap-api-references.md` (SAP also documents a distinct `ValueMappingDesigntimeArtifactSaveAsVersion` action not yet used) | Yes | via `DeployValueMappingDesigntimeArtifact` action | Resource + Data Source |
| Integration Content | Value Mapping Entries (`UpsertValMaps`/`UpdateDefaultValMap`/`DeleteValMaps`) | Design-time | Yes | OData V2 | Partial (`ValMapId` lookup only) | Yes (into an existing scheme only — 404 otherwise) | Yes (`UpdateDefaultValMap`) | Yes (granularity unconfirmed) | N/A | Unsupported — exact payload/path shapes not confirmed against a reachable primary source, see resource-design.md |
| Integration Content | Script Collection (Designtime Artifact) | Design-time | Yes | OData V2 | Yes | Yes | Yes (new version) | Yes | via `DeployScriptCollectionDesigntimeArtifact` action | Resource (planned v0.1.x stretch) |
| Integration Content | Message Mapping (Designtime Artifact) | Design-time | Yes | OData V2 | Yes | Yes | Yes (new version) | Yes | via `DeployMessageMappingDesigntimeArtifact` action | Resource (planned) |
| Integration Content | Service Endpoints | Runtime | Yes | OData V2 | Yes (read-only) | No | No | No | N/A | Data Source (planned) |
| Integration Content | Message Processing Logs | Runtime | Yes | OData V2 | Yes (read-only) | No | No | No | N/A | Unsupported — monitoring data, not infrastructure state (see provider-scope.md §66) |
| Integration Content | Message Stores / Data Stores | Runtime | Yes | OData V2 | Yes | Partial | Partial | Yes | N/A | Unsupported (v0.1.x) — payload/queue content is operational data, not desired state |
| Security Content | User Credentials | Design-time | Yes | OData V2 | Yes (metadata only; secret values are never returned) | Yes | Yes | Yes | N/A | Resource (planned, `write-only` semantics) |
| Security Content | OAuth2 Client Credentials | Design-time | Yes | OData V2 | Yes (metadata only) | Yes | Yes | Yes | N/A | Resource (planned) |
| Security Content | Keystore Entries (certificates) | Design-time | Yes | OData V2 | Yes | Yes | Yes | Yes | N/A | Resource (planned) |
| Security Content | Certificate-User Mapping | Design-time | Yes | OData V2 | Yes | Yes | Yes | Yes | N/A | Resource (planned) |
| Security Content | Access Policies | Design-time | Yes (confirmed: "retrieved [read and write] by an OData V2 API", SAP official documentation) | OData V2 | Yes | Yes | Yes | Yes | N/A | Resource + Data Source (v0.1.0) |
| Security Content | Access Policy Artifact References | Design-time | Yes (nested under Access Policies) | OData V2 | Yes | Yes | Yes | Yes | N/A | Resource (v0.1.0) — has its own composite identity, see resource-design.md |
| Partner Directory | Partner Directory Entries | Design-time | Yes | OData V2 | Yes | Yes | Yes | Yes | N/A | Unsupported (v0.1.x) — deferred, not yet schema-designed |
| API Management (classic) | API Providers / Proxies / Products | Design-time & Runtime | Yes (separate REST API, "Accessing API Management APIs Programmatically") | REST/OData mixed | Yes | Yes | Yes | Yes | Yes | Unsupported (v0.1.x) — deferred to a later minor version |
| API Gateway (new model) | API Artifact | Design-time | Not yet confirmed in enough detail for a stable schema | Unknown | Unknown | Unknown | Unknown | Unknown | Unknown | Unsupported (v0.1.x) — planned v0.2.x once confirmed |

Rows marked "Unsupported (v0.1.x)" are real API surfaces that a future minor version can
reasonably add; rows in `provisioning-capability-matrix.md` marked "Unsupported – manual
bootstrap required" are gaps with no known public API at all.
