---
page_title: "Integration Advisor"
subcategory: "Additional Capabilities"
description: |-
  Why Integration Advisor has no public design-time API today, how it relates to the Cloud
  Integration flows it injects generated artifacts into, and exactly what evidence this
  conclusion rests on.
---

# Integration Advisor

Integration Advisor is SAP Integration Suite's collaborative B2B interface-content-design
capability: Message Implementation Guidelines (MIGs), Mapping Guidelines (MAGs — standard,
overlay, and XSLT), custom Type Systems, Codelists, Shared Code, and Global Parameters, managed
under its own dedicated workspace.

If you came here looking for `sapintegrationsuite_message_implementation_guideline` or similar:
**it does not exist, and this guide explains exactly why**.

## The short version

This provider found **no public API** for any Integration Advisor design-time object, after
checking roughly eighty-five pages across its documentation tree. Every one of these objects is
real, working SAP functionality, fully accessible through Integration Advisor's own UI — this
provider does not invent an API SAP has not documented, so none of it is implemented as a
Terraform resource or data source.

## How this conclusion was reached, and one finding worth explaining carefully

The research method that worked for Classic API Management and Integration Assessment — look for
a dedicated "Accessing ... APIs Programmatically" or "Creating Service Instance and Service Key"
page — was applied here too, and, as with Trading Partner Management, came back empty.

One page initially looked promising and deserved a full read rather than a quick dismissal:
"Creating OAuth Client Credentials for Cloud Foundry Environment." On inspection, it describes
creating a *Process Integration Runtime* service instance (plan `api`, role
`WorkspacePackagesEdit`) — **the same Cloud Integration OAuth credential this provider already
uses for its own `oauth` block**, not a separate credential for Integration Advisor's own
content. It exists in Integration Advisor's documentation because those credentials are needed
for the *injection* step described below, not for managing MIGs, MAGs, Type Systems, or
Codelists themselves. This is worth recording precisely because it is exactly the kind of
false-positive lead that a less careful research pass could misread as evidence of a public
Integration Advisor API.

## Runtime artifact injection targets Cloud Integration — it is not an Integration Advisor API

Integration Advisor generates runtime artifacts (XSLT transformations, validation schemas, and
similar) from a Mapping Guideline, and can push them directly into an integration flow's
resources on a target Cloud Integration tenant — confirmed as a UI wizard: *Mapping Guideline* >
*Inject* > *SAP Cloud Integration Flow Resources* > choose a tenant/package/integration flow >
*Inject*. The target tenant can be Integration Advisor's own built-in Cloud Integration tenant, or
an external one configured as a BTP Destination (using the same Process Integration Runtime OAuth
credentials mentioned above).

This is a UI wizard, not a documented REST call, and — independent of that — it is a one-shot
imperative action (push these artifacts into that integration flow, right now) rather than
desired-state configuration Terraform's plan/apply model could represent, even if a REST
equivalent were confirmed tomorrow. This provider already manages Cloud Integration integration
flow content directly (`sapintegrationsuite_integration_flow`); Integration Advisor's injection
mechanism is a design-time authoring convenience that writes into that same surface, not a
separate desired-state object this provider would model on its own.

## What this provider manages today

Nothing new. `sapintegrationsuite_integration_flow` already manages the Cloud Integration content
an Integration Advisor injection would modify; nothing about researching Integration Advisor
changes that resource's existing scope.

## Revisiting this decision

Re-check for a dedicated API-access documentation page (the signal that has worked for Classic
API Management and Integration Assessment) before assuming this conclusion has changed, and be
careful not to mistake a page describing OAuth credentials for a *different*, already-integrated
capability (as the Cloud Foundry OAuth page here could be misread) for evidence of Integration
Advisor's own API.
