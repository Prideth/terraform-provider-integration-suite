---
page_title: "Migration Assessment"
subcategory: "Additional Capabilities"
description: |-
  Why Migration Assessment has no public API today, and why its own object model would stay out
  of scope even if one were confirmed.
---

# Migration Assessment

Migration Assessment helps a tenant evaluate migrating integration scenarios from an on-premises
SAP Process Orchestration system (7.31 SP28+, 7.40 SP23+, or 7.50 SP06+) to SAP Integration
Suite: register a source system, extract its integration scenario data, evaluate that data
against migration readiness rules, and review the resulting assessment-category classifications,
readiness scores, and effort estimates. The actual migration tooling that moves content into
integration flows is a separate, already-covered Cloud Integration feature — Migration Assessment
only assesses.

If you came here looking for `sapintegrationsuite_migration_assessment_source_system` or a
resource to trigger an evaluation: **neither exists, and this guide explains exactly why**.

## No public API, and a smaller doc footprint than most capabilities audited

Migration Assessment's own documentation tree is small — around fifteen pages — and entirely
checked. No dedicated API-access or service-key page exists anywhere in it, the same
negative-evidence signal that has reliably indicated "no public API" for every other capability
audited this way in this project (Current API Management, Trading Partner Management, Integration
Advisor).

One detail is worth being precise about: Migration Assessment's documentation *does* mention
APIs — but as something Migration Assessment itself calls, not something a Terraform provider
could call into it. To extract data, Migration Assessment reaches into the registered source
system's own SAP Process Orchestration APIs (through the SAP Destination service and, typically,
Cloud Connector) — the opposite direction from a public API this provider could manage Migration
Assessment's own objects (source systems, requests, results) through. SAP's own known-limitations
page even notes a gap in the *other* direction: "Migration Assessment cannot collect performance
metrics ... for Integration Engine (ABAP) components through the available APIs" — again, APIs
Migration Assessment consumes from the source system, not ones it exposes.

## Even with a confirmed API, this capability's own objects would stay out of scope

This is worth stating plainly, because it means revisiting this conclusion later is unlikely to
change the outcome the way it might for, say, Classic API Management's API Proxy. Every object in
Migration Assessment is either:

- **Action-triggered workflow**: SAP's own procedure for creating a Data Extraction Request
  describes choosing *Create*, which immediately starts the extraction and produces a resulting
  `Completed` / `Completed with warnings` / `Completed with errors` status — an imperative action
  with a result, not a declarative object a `terraform apply` would converge toward.
- **Reporting output**: Scenario Evaluation results are assessment-category classifications
  (*Ready to Migrate* / *Adjustment Required* / *Evaluation Required*), migration-readiness
  scores, and effort estimates — analysis output, the same category this provider already
  excludes for Cloud Integration's own Message Processing Logs.

This provider does not create action-style resources for "run assessment," "retry extraction," or
similar — the same standing policy already applied to Trading Partner Management's agreement
activation and Developer Hub's Subscription approval workflow.

## What this provider manages today

Nothing, and nothing here changes what this provider already manages elsewhere: Migration
Assessment's own source-system registration and requests are entirely separate from the Cloud
Integration content (`sapintegrationsuite_integration_flow` and siblings) this provider already
manages, which is what the *actual* migration tooling — a distinct feature — eventually produces.

## Revisiting this decision

Re-check for a dedicated API-access documentation page before assuming this conclusion has
changed. Even if one appears, evaluate the source-system registration object on its own merits
first (the closest thing here to practitioner-authored configuration) — Data Extraction and
Scenario Evaluation should stay classified as workflow/reporting regardless, per the reasoning
above.
