---
page_title: "Data Space Integration"
subcategory: "Additional Capabilities"
description: |-
  What Data Space Integration offers through its API, which of its objects would suit Terraform,
  and why this provider does not manage any of them yet.
---

# Data Space Integration

Data Space Integration is the capability of SAP Integration Suite for taking part in data spaces
such as Catena-X. A data space lets companies exchange data peer to peer under agreed rules,
based on the Dataspace Protocol. SAP's concepts page names the Eclipse Dataspace Connector as the
connector framework involved.

**This provider does not manage anything in Data Space Integration yet.** This guide explains
which objects would suit Terraform, what SAP documents about their API, and what is missing.

## The objects, and which of them are configuration

A participant can be a **provider** (offering data), a **consumer** (using data offered by
others), or both. The objects fall into two groups.

Configuration that a provider or administrator defines and maintains, and that Terraform could
manage:

| Object | What it is | Lifecycle notes from SAP's UI documentation |
|---|---|---|
| Asset | A set of data offered to the data space, with a data address (HTTP endpoint or S3 bucket) | The ID (`A-Z`, `a-z`, `0-9`, `.`, `_`, `-`) is generated if not given and cannot be changed. SAP warns not to edit the properties of an asset used in an active agreement. The data address contains credentials: a Basic password for HTTP, an access key and secret for S3. |
| Policy | Rules on who may access an asset (access policy) or how it may be used (usage policy) | Rules combine an attribute (Business Partner Number, Business Partner Group, Membership), an operator (`is`, `is any of`, …) and a value. Consumers can see every detail of a usage policy. |
| Contract definition | An offer: assets bundled with one access policy and one usage policy | The two policies must differ. Editable only while not in use. |
| Company policy | Rules restricting which offers a consumer may accept | SAP predelivers some; those cannot be edited or deleted, only copied. Requires `DataspaceBusinessAdmin`. |
| Contract reference | A manually kept record of a contract with a business partner, used in policies | Requires `DataspaceBusinessAdmin`. |

Runtime processes, which are not desired state and would stay out of scope in any case: catalog
requests, contract negotiations, contract agreements, and transfer processes. They are what a
consumer's applications do at run time, like message processing in Cloud Integration.

## What SAP documents about the API

SAP's page *Using APIs to Work With Data Space Integration* says the capability "provides APIs for
accessing and managing resources" and points to the package `dataspaceintegration` on the SAP
Business Accelerator Hub. The Hub's public catalog lists one API there, **Data Space
Integration** (`DSIAPI`), version 2.0.0, of type REST, described as "Manage Data Space
Integration resources via API". The API reference itself is only visible after an SAP login.

The Help pages show concrete requests only for the consumer side, all under `/api/dsi/v1`:

- `POST /api/dsi/v1/catalog` with `counterPartyAddress` (the partner's connector URL ending in
  `/api/v1/dsp`), to discover offers
- `POST /api/dsi/v1/contract-negotiation`, then `GET …/contract-negotiation/{id}/state` and
  `GET …/contract-negotiation/{id}/agreement`
- `POST /api/dsi/v1/transfer-process` and `GET …/transfer-process/{id}/edr`
- the *Convenient Data Request* API, which runs the whole consumer flow in one call

For assets, policies, contract definitions, company policies and contract references, SAP
documents only the UI procedure. The role description for `DataspaceProvider` ("create assets,
policies, and contract definitions") does not say whether that includes the API. Whether the
API manages these objects, and with which fields, is therefore not publicly confirmed.

The Eclipse Dataspace Connector has a public open-source management API. The provider does not
borrow field names from it: SAP's API under `/api/dsi/v1` is its own layer, and assuming the two
match would be a guess.

## Credentials: one set per connector

API access uses its own service instance, *Data Space Integration API Access* with plan `api`,
created in a Cloud Foundry space. When creating it you choose the role (`AuthGroup_DataspaceConsumer`,
`AuthGroup_DataspaceProvider`, or both), the grant type `client_credentials`, and the
**connector name**. The service key is either a client ID and secret or a certificate.

SAP requires **one service instance per connector** and warns that sharing the wrong instance
"can lead to data leaks in which a party can gain access to another party's data". An
implementation would therefore need a separate credential block, used with one provider alias
per connector, and could not reuse the Cloud Integration credentials.

## What would unblock an implementation

The missing piece is the specification of `DSIAPI`. It is a REST API, so unlike Integration
Assessment there is no OData `$metadata` to fetch from a tenant. Two ways to get the contract:

- Download the API specification from the Business Accelerator Hub while logged in with an SAP
  account, and check the provider-side endpoints against it.
- SAP publishes the provider-side endpoints in its Help documentation.

With a confirmed contract, the natural order would be policies, then assets (with the data
address credentials as write-only attributes), then contract definitions, which reference both.
Company policies and contract references would follow. Onboarding a connector to a data space
is documented as UI-only and would stay outside Terraform.
