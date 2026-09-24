---
page_title: "Integration Assessment"
subcategory: "Additional Capabilities"
description: |-
  Integration Assessment's confirmed API surface, its full entity inventory classified by
  Terraform suitability, and exactly why this provider implements none of it yet.
---

# Integration Assessment

Integration Assessment implements SAP's Integration Solution Advisory Methodology (ISA-M): a
guided approach to assessing an organization's integration landscape and strategy, and to
recording decisions about which integration technology should handle which kind of integration.
It is a separate SAP BTP service subscription (entitlement `integration-assessment`), not a
sub-feature of Cloud Integration or either API Management model.

If you came here looking for `sapintegrationsuite_integration_assessment_application` or
similar: **nothing in this capability is implemented yet, and this guide explains exactly why**,
together with the confirmed entity inventory this provider's decision rests on.

## Authentication is confirmed, and genuinely separate again

Integration Assessment authenticates through its own BTP service instance — **Integration
Assessment APIs**, plan `default` — provisioned independently of the Integration Suite
subscription this provider otherwise assumes. Its service key is confirmed to contain:

- `entities` — base URL for the Entities API
- `management` — base URL for the Management API
- `clientid` / `clientsecret` / `url` (token server) — the same OAuth 2.0 client-credentials
  shape this provider already uses for Cloud Integration, current API Management, and Classic
  API Management

Two distinct base URLs, not one, is itself a confirmed and notable finding: whatever the
`entities` vs. `management` split actually means at the wire level was not investigated further,
since (see below) no implementation was reached that would need to distinguish between them.

This provider does not add a `provider.integration_assessment` configuration block in this
phase. Adding provider schema with nothing behind it to configure would be dead surface area —
the same discipline this provider already applies elsewhere (see `docs/provider-scope.md`). If
and when a resource or data source is implemented here, the block will be added alongside it,
following the same pattern as `provider.api_management`.

## The confirmed entity inventory

SAP's own "Integration Assessment APIs" documentation page lists every entity in this capability,
each with a one-paragraph description — genuinely the most complete *inventory* this provider has
found for any capability audited without a lucky primary-source PDF (compare
`docs/guides/classic-api-management.md`, where a complete official user guide with worked
examples was found). What is missing, despite substantial effort (the SAP-docs mirror, the
official "SAP Integration Solution Advisory Methodology" PDF user guide, and SAP's own TechEd
IN262 hands-on sample repository, all checked and all UI-procedure-only), is any worked
request/response example for any entity — so this inventory is classified by *shape and
described limits*, not by a confirmed wire contract.

### Master data — SAP-maintained ISA-M taxonomy

Domain, Style, Use Case Pattern, Integration Pattern, Key Characteristic (with its own Group,
Value, and Recommendation sub-entities), Deployment Model, Domain Determination. This is the
reference taxonomy the whole methodology is built on — SAP ships it, and a tenant can review and
adjust it (SAP's documentation includes a dedicated "Update Content Maintained by SAP" procedure).
If a public API is ever confirmed for it, this looks like a genuine data-source candidate at
minimum; whether the "adjust" capability rises to full resource-level suitability would need its
own suitability check once the API itself is confirmed.

### Landscape configuration — the strongest resource candidate

Application, Application Instance, Technology, Technology Instance, Vendor, and their
association entities (Technology Domain, Technology Style, Technology Key Characteristic). This
is practitioner-authored configuration, not reference data or workflow state, and SAP documents
concrete per-tenant limits that confirm real, bounded, persisted storage: a maximum of 20,000
Applications, 20,000 Application Instances, 50 Technologies, 150 Technology Instances, and
10,000 Vendors. If SAP's field-level Create/Read/Update/Delete contract for this group is ever
confirmed, it is the first place this provider would look to implement Terraform resources for
this capability.

### Requests and assessment workflow — out of scope regardless

Request and Request Line Item, the entry points for a business solution/interface request
workflow, plus the Integration Flow/Message Flow content a request references and Request Line
Item Technology Instance Decision. SAP documents an explicit status machine for Request: `draft`
(automatic on Create) → `new` (Submit) → `in progress` (automatic, once interface requests are
assessed, or via Reopen) → `completed` (Complete). This is workflow/project state, the same
category this provider already excludes for Cloud Integration's Message Processing Logs and
Developer Hub's Subscription object — out of scope by its nature, independent of whether a field
contract is ever confirmed for it.

## What this provider manages today

Nothing. See `internal/features/catalog.go`'s `integration_assessment.*` entries for the
per-group classification recorded above, and `docs/sap-api-references.md` for the full research
trail, including the specific sources checked and found to contain UI procedures only.

## Revisiting this decision

The gap here is narrower than it looks: the entity inventory, the authentication mechanism, and
the general API shape (two base URLs) are all genuinely confirmed. What is missing is one
specific kind of evidence — a worked request or response body for any entity — the same kind of
gap this provider already treats as `public_api_incomplete` elsewhere (see Classic API
Management's API Proxy). If SAP's Business Accelerator Hub page for the
`SAPIntegrationAssessment` package ever becomes reachable without an SAP support login, or a
primary source with worked examples surfaces, start with Landscape Configuration — the
strongest-evidenced candidate — before Master Data, and treat Requests/assessment workflow as
settled out of scope rather than reopening it.
