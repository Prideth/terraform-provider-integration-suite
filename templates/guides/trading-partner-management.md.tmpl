---
page_title: "Trading Partner Management"
subcategory: "Additional Capabilities"
description: |-
  Why Trading Partner Management has no public design-time API today, how it relates to the
  Partner Directory this provider already manages, and exactly what evidence this conclusion
  rests on.
---

# Trading Partner Management

Trading Partner Management (TPM) is SAP Integration Suite's design-time capability for B2B/EDI
partner onboarding: a company profile (with subsidiaries), trading partner profiles,
communication partner profiles, agreement templates, and trading partner agreements themselves —
managed under *Design* > *B2B Scenarios* in the SAP Integration Suite UI.

If you came here looking for `sapintegrationsuite_trading_partner_agreement` or similar: **it
does not exist, and this guide explains exactly why**.

## The short version

This provider found **no public API** for any Trading Partner Management design-time object,
after a thorough documentation research pass covering roughly ninety pages in SAP's own
documentation. Every one of these objects is real, working SAP functionality, fully accessible
through the SAP Integration Suite web UI — this provider does not invent an API SAP has not
documented, so none of it is implemented as a Terraform resource or data source.

## How this conclusion was reached

The research method that worked for Classic API Management and Integration Assessment — look for
a dedicated "Accessing ... APIs Programmatically" or "Creating Service Instance and Service Key"
page — was applied here too, and came back empty. Every capability in this provider that does
have a confirmed public API has at least one such page in SAP's documentation; Trading Partner
Management has none, across its entire documentation tree. A targeted search within that tree for
"API", "REST", and "OData" across its overview, task/permission, export, and configuration pages
returned zero matches. This is treated as reasonably strong negative evidence, not proof by
absence alone — SAP's documentation practice for every *other* capability this provider has
researched has been consistent enough (a dedicated access page exists whenever a public API
does) that its consistent absence here is itself meaningful.

What TPM does document, all UI-only:

- **Export/Import**: a *Download* button on Company Profile, Trading Partner, Agreement
  Template, and Agreement produces a browser-downloaded JSON file (`company.json`,
  `TradingPartner_<name>.json`, `Template_<name>.json`, `Agreement_<name>.json`). This confirms
  these objects are internally JSON-shaped, but a UI file download is not a REST endpoint this
  provider could build a resource against.
- **Activation**: activating a trading partner agreement is a UI action with a confirmed,
  significant side effect — see below.

## Trading Partner Management generates Partner Directory content — it does not replace it

This provider already manages the **Partner Directory** (`sapintegrationsuite_partner_string_parameter`,
`sapintegrationsuite_partner_binary_parameter`, `sapintegrationsuite_alternative_partner`,
`sapintegrationsuite_partner_authorized_user`, `sapintegrationsuite_partner_user_credential_parameter`
— see `docs/guides/partner-directory.md`), a confirmed, generally-applicable runtime
configuration store any integration flow can parameterize against. Trading Partner Management is
a *specific, guided way to populate that same store for B2B/EDI scenarios*, not a separate
storage layer:

> "When a trading partner agreement gets activated, the complete agreement information gets
> pushed into the partner directory. An entry is created in the partner directory for each
> business transaction activity in the agreement and for each Interchange Envelope extraction."

Generated entries are visible (read-only) under *Design* > *B2B Scenarios* > *Partner Directory
Data*, prefixed `SAP_TPM`. This provider deliberately keeps the two concepts distinct rather than
conflating them:

- **Partner Directory** (implemented): a generic runtime key-value store this provider manages
  directly, entry by entry, through its own confirmed OData V2 API.
  - **Trading Partner Management** (not implemented): a design-time workflow that, as a *side
  effect* of an imperative *Activate* action, bulk-generates a specific shape of Partner
  Directory entries. Even if TPM's own API were confirmed tomorrow, this generation step would
  stay out of scope for the same reason this provider does not model other imperative
  generate/replicate/publish actions as Terraform resources: `terraform apply` represents desired
  state, not a one-shot action to trigger. A practitioner who wants a specific Partner Directory
  entry to exist already has the tool for that — the resources listed above — regardless of
  whether TPM's UI happens to be the mechanism that produced an equivalent entry for someone else.

## What this provider manages today

Nothing new. Existing Partner Directory support (see `docs/guides/partner-directory.md`) already
covers the runtime store TPM happens to write into; nothing about researching TPM changed that
guide's conclusions or scope.

## Revisiting this decision

Re-check for a dedicated API-access documentation page (the same signal that worked for Classic
API Management and Integration Assessment) before assuming this conclusion has changed. If SAP
ever publishes one, evaluate Company Profile and Trading Partner/Communication Partner Profile
first — they are the more clearly practitioner-authored, desired-state-shaped objects — and treat
Partner Directory generation as out of scope regardless, per the reasoning above.
