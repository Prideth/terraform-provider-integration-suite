# SAP Integration Suite public API re-audit, September 2026

This document records the September 2026 re-audit of every SAP Integration Suite capability
against SAP's public API surface: what was checked, which earlier conclusions turned out to be
wrong, what changed in the provider, and what is still open. It is written for maintainers and
contributors. Per-object evidence lives in `docs/sap-api-references.md` and in the feature
catalog (`internal/features/catalog.go`); user-facing explanations live in the guides.

All changes are on the branch `feature/sap-2026-api-gap-closure` and listed in the Unreleased
section of `CHANGELOG.md`.

## Why a re-audit was needed

Earlier research phases ran without access to `help.sap.com`, `api.sap.com` or a tenant. Several
wire contracts were therefore reconstructed from UI labels and prose. The re-audit found that
some of them were wrong in ways that would fail on a real tenant: the access policy reference
sent property names SAP's entity does not have, the integration adapter sent `Type` and
`Application`, which do not exist, and the service endpoint data source read an API definition
`Type` that is always empty. The ROADMAP's statement that "current public API coverage is
complete" rested on those contracts and has been revised.

## Method and evidence

Evidence was ranked as follows, and a lower source never overrode a higher one:

1. **The tenant's own `$metadata`.** A Cloud Integration `/api/v1/$metadata` document from a
   development tenant settles property names, keys, types, length facets and function imports.
   It is kept out of the repository (`.specs/`, excluded via `.git/info/exclude`).
2. **SAP's API specifications** on the Business Accelerator Hub, where reachable.
3. **SAP Help**, read from the SAP-docs GitHub mirror (`SAP-docs/btp-integration-suite`), because
   `help.sap.com` and `api.sap.com` are single-page applications that return no content to a
   plain HTTP client.
4. **SAP's own tooling and samples**: the API Management Client SDK 3.0.6 (disassembled),
   `SAP/cicd-actions-for-sap-integration-suite`, `SAP/apibusinesshub-api-recipes`.

Two techniques are reusable and documented in `CONTRIBUTING.md`:

- **Contract tests against `$metadata`.** `internal/testutil/edmx` parses a `$metadata` document,
  and each OData client package has a `metadata_contract_test.go` that checks every wire
  struct's JSON fields (including inherited ones), every key and every function import the
  client uses, plus selected `MaxLength` facets. The tests skip without a document, so CI stays
  green without tenant access; a skipped run is not evidence.
- **The Hub's catalog service.** `https://api.sap.com/odata/1.0/catalog.svc` answers anonymously
  at package level: the complete package list (1,971 packages, paged via `__next`) and each
  package's APIs with type and version. It ignores `$filter`, so filtered queries prove nothing,
  and specification files stay behind the Hub login.

A read-only probe script was run against the development tenant. Apart from access policies it
returned `403 Forbidden`, because the OAuth client lacked role templates. That is why several
items below still say "waiting for a tenant check".

## Results per area

| Area | Verdict after the re-audit | Main changes |
|---|---|---|
| Access policies | Supported, contract corrected | Property names from SAP's own access policy tooling and the `$metadata`; `Edm.Int64` keys; `PUT` for updates; runtime assignments data source |
| Current API Management, Integration Cell | No public API, unchanged | MCP servers added to the catalog; evidence from the Integration Content API, Client SDK 3.0.6 and What's New |
| Security Content | Supported, fields completed | OAuth2 client credential fields `ClientAuthentication`, `ScopeContentType`, `Resource`, `Audience`; keystore entry metadata; catalog entries for OAuth2 password and SAML bearer credentials, where-used and PGP keyrings |
| Cloud Integration | Supported, several fixes | Integration flow configuration resource; `redeploy_triggers`; `save_as_version` via the documented SaveAsVersion function imports; adapter and service endpoint fields aligned with `$metadata`; reclassifications of value mapping entries, data types, message types, service interfaces |
| Archiving (Cloud Integration, B2B) | New catalog entry, no resource | Activation is a one-way function import without deactivation; per-flow settings are UI-only |
| Edge Integration Cell | Experimental | `runtime_location_id` on deployments, credentials, certificates, key pairs and Partner Directory resources, using the documented `/location/<id>/api/v1` service root |
| Partner Directory | Supported, fixes | User credential parameters updated in place (documented POST upsert) and never overwritten on create; authorized users must be lowercase; binary parameter limit follows the `$metadata` (1.5 MiB); `runtime_location_id` kept in state |
| Classic API Management | Unchanged, one open check | Client SDK 3.0.6 and virtual host findings; the Hub lists *API Portal - Transport (CF)* as the official ZIP import and export API |
| API Composition | Experimental resource | Business data graph resource and data source; schema follows SAP's configuration file format; asynchronous processing with timeouts; the Hub lists the API as OData V4 |
| Integration Assessment | Public API, contract unknown | Both APIs are OData per the Hub; the `$metadata` fetched with a service key would unblock Landscape Configuration |
| Data Space Integration | Public API, contract unknown | One REST API (`DSIAPI` 2.0.0); SAP documents only consumer runtime calls; new guide |
| OData Provisioning | No public management API | Reclassified from "research required"; the `ODPAPIAccess` role is runtime access |
| Trading Partner Management, Integration Advisor, Migration Assessment | No public API, unchanged | Re-checked against current docs and the full Hub package list; B2B Scenarios API identified as B2B monitoring |
| Roles | Documented | New "Authorization and Roles" guide |

## Earlier conclusions that were corrected

- Access policy references: invented property names (`ArtifactType`, `Attribute`, `Operator`,
  `Value`) replaced by SAP's (`Type`, `ConditionAttribute`, `ConditionType`, `ConditionValue`,
  plus `Name` and `Description`).
- Integration adapter: `Type` and `Application` do not exist on the entity.
- Service endpoints: the API definition has `Name`, not `Type`.
- Partner Directory: authorized users are stored lowercase (previously "unconfirmed");
  user credential parameters can be updated (previously "no update API"); binary parameters
  allow 1.5 MiB (previously 260 KB).
- API Composition: the Configuration API uses an API Composition instance with plan
  `configuration`, not Process Integration Runtime with plan `integration-flow`.
- Data Space Integration: a REST API, not OData.
- OData Provisioning: no management API; `ODPAPIAccess` had been misread as a sign of one.
- Hub searches: a filtered catalog query had been taken as evidence; the full package list is
  used instead.

## Breaking changes

Listed with migration notes in `CHANGELOG.md`: the access policy reference schema and removed
`reconciliation_status`; integration adapter without `type` and `application`; service endpoint
`api_definitions[].name`; lowercase `user` for authorized users. Import IDs for access policies
must be numeric.

## Open items and what would close them

| Item | Blocked on | How to close it |
|---|---|---|
| Classic API Management wire structs, virtual host read schema, further CF APIs (key store, trust store, applications, developers, rules, access control) | `Management.svc/$metadata` | Download `/apiportal/api/1.0/Management.svc/$metadata` with an API portal service key (plan `apiportal-apiaccess`); add a contract test |
| API Proxy content upload | Transport API specification | Download *API Portal - Transport (CF)* from the Hub with an SAP login |
| API Composition PATCH body, delete, service key fields | Configuration API `$metadata`, a live test | Service key of the `configuration` plan |
| Integration Assessment Landscape Configuration | `EntitiesAPI` `$metadata` | Service key of *Integration Assessment APIs* |
| Data Space Integration assets, policies, contract definitions | `DSIAPI` specification | Hub download with an SAP login |
| Secure parameters, number range GET/DELETE, OAuth2 custom parameters, access policy constants | Tenant check | Re-run the tenant probe after assigning the missing role templates |
| Edge Integration Cell targeting | A tenant with an Edge Integration Cell | Run the probe with `-RuntimeLocationId` |
| Value mapping entries | Payload and delete granularity | SAP documentation or a tenant test |

## Reproducing the checks

1. Place a `$metadata` document at `.specs/cloudintegration-metadata.xml` (or set
   `SAP_INTEGRATION_SUITE_METADATA_FILE`) and run `go test ./internal/client/...`.
2. For the Hub, list packages with
   `curl "https://api.sap.com/odata/1.0/catalog.svc/ContentEntities.ContentPackages?\$select=TechnicalName,DisplayName&\$format=json"`
   and follow `__next`; list a package's APIs with
   `.../ContentPackages('<name>')/Artifacts?$format=json`.
3. The SAP-docs mirror is at `https://github.com/SAP-docs/btp-integration-suite`; the per-area
   folders are `docs/ISuite_*`.
