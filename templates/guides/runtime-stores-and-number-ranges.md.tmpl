---
page_title: "Number Ranges, Variables, and Data Stores"
subcategory: "Cloud Integration"
description: |-
  What SAP's Message Stores API family (Number Ranges, Variables, Data Stores, Data Store
  Entries) supports, why this provider manages only Number Ranges' static configuration, and
  why the others are intentionally out of scope.
---

# Number Ranges, Variables, and Data Stores

SAP Cloud Integration's "Message Stores" OData V2 API groups four related object types that
integration flows use to persist data during message processing: **Number Ranges**, **Variables**,
**Data Stores**, and **Data Store Entries**. This provider manages only one of them —
[`sapintegrationsuite_number_range`](../resources/number_range.md), and only its static
configuration, never its runtime counter. The other three are deliberately not implemented as
Terraform resources or data sources at all. This guide explains why, in detail, because the
reasoning is not obvious from the API surface alone: every one of these four object types has a
real, documented, working API — the difference is what each API actually lets a practitioner
safely do with Terraform.

If you take away one thing from this guide: **a public API is not the same thing as a Terraform
resource.** See `docs/provider-scope.md` for that principle stated in general terms; this guide
is the concrete case study.

## Number Ranges

A Number Range generates unique interchange numbers for outbound EDI/EDIFACT documents — SAP's
own description: "While sending out a document when using EDI processing, a unique interchange
number must be added to each document." It has two very different parts:

- **Static configuration** a practitioner defines: `name`, `min_value`, `max_value`,
  `description`, `rotate`, `field_length`.
- **A live runtime counter** (SAP's `CurrentValue`, shown as **"Next Value"** in the Integration
  Suite Monitor UI — same value, different name) that advances every time deployed content
  consumes a number.

`sapintegrationsuite_number_range` manages the first part. It does not, and cannot safely, manage
the second.

### Why this resource has no Read

Every other resource in this provider implements a real `Read` that calls SAP's API and detects
drift. This one cannot: **SAP documents no `GET` operation for Number Ranges anywhere.** This was
checked exhaustively, not assumed from one missing example — see `docs/sap-api-references.md` for
the full trail (the curated example-requests index that lists a "Get..." page for every sibling
entity except this one; a full directory search; the overview resource table, which documents
unsupported query options for Data Stores and Variables but says nothing of the kind for Number
Ranges, implying there is no GET to apply query options to in the first place).

Because of that, `sapintegrationsuite_number_range`'s `Read` is a **documented no-op**: it copies
whatever Terraform already has in state back into state, and never contacts SAP. It cannot tell
you if someone changed `description` through the UI. It cannot tell you if the object was deleted
entirely. This is disclosed in the resource's own schema description, not just here.

### Why the runtime counter is a separate, write-only attribute

If `current_value` were an ordinary attribute, every `terraform plan` would either try to
"correct" SAP's live counter back to whatever Terraform last wrote (resetting real consumption
that happened through deployed EDI/EDIFACT content — a destructive, silent, and very unwelcome
surprise) or show permanent, unresolvable drift Terraform could never converge. Neither is
acceptable.

Instead, the counter is exposed as two attributes, mirroring the write-only credential-rotation
pattern this provider already uses for `sapintegrationsuite_user_credential`:

```hcl
resource "sapintegrationsuite_number_range" "invoice_numbers" {
  name         = "InvoiceNumbers"
  min_value    = "0"
  max_value    = "999999"
  description  = "Interchange numbers for outbound EDIFACT invoices"
  rotate       = true
  field_length = "6"

  current_value_wo         = "0"
  current_value_wo_version = "initial"
}
```

- `current_value_wo` is `WriteOnly` — Terraform never stores it in plan or state, matching every
  other `_wo` attribute in this provider.
- It is sent to SAP on every `Create` (SAP's documented example always includes `CurrentValue`
  when adding an object).
- On `Update`, it is sent **only when `current_value_wo_version` changes** from what is already
  in state. An apply that changes only `description` — or `min_value`, `max_value`, `rotate`,
  `field_length`, or all four at once — never sends `CurrentValue` at all. This is the mandatory
  guarantee behind this design, and it is covered by a dedicated regression test at both the
  client layer (`TestClient_UpdateNumberRange_OmitsCurrentValueWhenNil`) and the resource layer
  (`TestNumberRangeResource_Update_OmitsCurrentValueWhenVersionUnchanged`): given a Number Range
  whose counter has advanced to some arbitrary live value through real consumption, an apply that
  only touches static configuration produces a `PUT` request body containing no `CurrentValue`
  property whatsoever — not the value Terraform last knew, and not any other guessed value.

To deliberately push a new counter value — for example, correcting the counter after a manual
intervention outside Terraform — bump `current_value_wo_version` to any new string and set
`current_value_wo` to the value you want:

```hcl
resource "sapintegrationsuite_number_range" "invoice_numbers" {
  # ...unchanged static configuration...
  current_value_wo         = "1000"
  current_value_wo_version = "manual-correction-2026-01"
}
```

### An open, honestly-disclosed risk: what does an omitted field actually do?

The design above — omitting `CurrentValue` from the request body on an ordinary Update — is the
safest option this provider could construct given the API it has, but it rests on an assumption
this provider **could not verify**: that SAP's `PUT` treats a missing `CurrentValue` as "leave
unchanged" (a partial-merge interpretation), not "reset to a default" (a literal full-replace
interpretation). There is no `GET` to check which one actually happens. If your tenant's behavior
turns out to be the latter, please report it — this is exactly the kind of gap this provider
would rather document prominently than paper over with false confidence.

### Delete and Import are both refused, not guessed at

SAP documents no `DELETE` for this entity either. The Integration Suite Monitor UI shows an
**"Undeploy"** action for Number Ranges — explicitly distinct from "Delete" in SAP's own Actions
documentation — with no REST equivalent found anywhere. `terraform destroy` (or removing the
resource block and applying) therefore returns an explicit, actionable error rather than either
silently doing nothing (leaving Terraform's state out of sync with a belief that the object was
removed) or guessing at an unconfirmed operation against a live tenant object that deployed EDI
content may still depend on:

```shell
$ terraform destroy
╷
│ Error: Destroying sapintegrationsuite_number_range is not supported
│
│ SAP documents no delete operation for the NumberRanges API — only an
│ 'Undeploy' action in the Monitor UI with no confirmed REST equivalent...
│ To stop managing this Number Range with Terraform without changing
│ anything on the tenant, remove it from state with 'terraform state rm'.
```

Import is refused for the same underlying reason (no GET), via the same explicit-error pattern:

```shell
$ terraform import sapintegrationsuite_number_range.invoice_numbers InvoiceNumbers
╷
│ Error: Importing sapintegrationsuite_number_range is not supported
│
│ SAP documents no GET operation for the NumberRanges API, so this provider
│ has no way to read an existing Number Range's configuration back from the
│ tenant...
```

If a Number Range already exists on your tenant and you want Terraform to manage its static
configuration going forward, write a `sapintegrationsuite_number_range` resource block matching
its current settings and run `terraform apply` — but first verify, in the Monitor UI, what
happens when you `POST` a `Name` that already exists. SAP's documentation does not confirm
whether that errors, conflicts, or silently overwrites, and this provider does not guess.

### A UI capability this provider does not expose: multi-runtime deployment

The Monitor UI's Add/Edit dialog for a Number Range includes a **"Runtimes"** field: "One or more
runtime nodes to deploy the artifact to... including Cloud Integration and any active Edge
Integration Cell nodes." Neither of SAP's two documented API examples (Add, Update) shows any
runtime/location parameter. This provider's client always targets the implicit default runtime
and does not attempt to reconstruct or guess at Edge Integration Cell-specific deployment — if
your tenant has active Edge Integration Cell nodes and you need Number Ranges deployed there
specifically, use the Monitor UI for that until SAP documents the API parameter.

### Numeric fields are strings, on purpose

`min_value`, `max_value`, `field_length`, and `current_value_wo` are all Terraform **strings**,
not numbers, matching SAP's own wire format exactly (every one of these fields is a JSON string
in SAP's documented example payloads). SAP documents `max_value` as allowing up to 14 digits —
this provider validates the digit-string shape and range client-side, but deliberately never
converts these values to a Go/Terraform numeric type, to avoid any risk of precision loss on
values this large.

### Authorization

No Cloud Foundry role template specific to Number Ranges creation/update was found documented
anywhere in SAP's public documentation — unlike Data Stores and Variables, which have explicit
role templates (see below). If you find the specific role template SAP requires, please
contribute a documentation update; until then, this provider does not claim any specific role
collection is sufficient or necessary beyond what your tenant already requires for other Message
Stores API access.

## Variables — no resource, no data source

A Variable is a value an integration flow's **Write Variables** step writes during message
processing, to be read by a later step in the same flow (a local variable) or by any other flow
deployed on the tenant (a global variable). SAP documents exactly **one** public operation for
this entity:

```
GET /api/v1/Variables(VariableName='{VariableName}',IntegrationFlow='{IntegrationFlowName}')/$value
```

This downloads the raw value — nothing else. There is no collection GET (you must already know
the exact variable name and integration flow), no `POST`, no `PUT`, and no confirmed `DELETE`.
Variables are created and updated **exclusively** by deployed integration flow content, through
an entirely different ownership model than Terraform-managed infrastructure — SAP's own
documentation is explicit that this is a mechanism "to share data across different integration
flows," not a configuration object. Variables also expire automatically after 400 days of
inactivity, extended by every successful processing run.

A read-only `data.sapintegrationsuite_variable` was considered and rejected. The single confirmed
read endpoint returns nothing but the raw runtime value, with no metadata fields (no
`Visibility`/`UpdatedAt`/`RetainUntil`) to fall back to as a safer, non-payload alternative. Every
`terraform plan` against such a data source would either show spurious churn as the underlying
business value changes for reasons entirely outside Terraform's control, or silently capture a
snapshot of runtime business data into `.tfstate`. This provider's principle — a Terraform state
file is not an Integration Suite runtime database — applies here without exception.

## Data Stores and Data Store Entries — no resources, no data sources

A **Data Store** is a runtime container that comes into existence implicitly, the first time an
integration flow's Data Store Write step (or an XI adapter configured with
`Temporary Storage = Data Store`) writes an entry to it. A **Data Store Entry** is one persisted
message — payload, headers, and processing metadata (`Status`, `MessageId`, `DueAt`, `CreatedAt`,
`RetainUntil`) — inside a Data Store.

SAP documents `GET`-only access to both:

- `GET /api/v1/DataStores?overdueonly=true` — an aggregate **monitoring** endpoint, returning
  message counts per store, not configuration. This is the same class of data as this provider's
  already-excluded Message Processing Logs: useful for observability, not something Terraform
  reconciles.
- `GET .../DataStoreEntries(...)` (a single entry, or all entries for a store) — runtime business
  message records.

There is no independent `POST`/`PUT` for either entity anywhere — a Data Store and its entries
come into existence only as side effects of deployed integration flow content executing.
**Delete does exist, but not as a REST operation**: SAP documents a design-time "Data Store
Delete" integration-flow step, explicitly scoped to "single entries only" (it "can't be used to
delete whole data stores"), operating on messages as an integration flow processes them — not
something this provider could reasonably wrap in a `terraform destroy` without misrepresenting
what that verb means for every other resource in this provider.

Neither entity gets a resource or a data source. This is intentional and, for Data Store Entries
specifically, treated as **`out_of_scope`** rather than merely `not_implemented`: exposing
message payloads and processing status in Terraform state has no legitimate infrastructure-as-code
purpose, and "an entry can be deleted somehow" is not a reason to model that deletion as
`terraform destroy` any more than it would be for a Message Processing Log record.

Both `DataStores` and `Variables` confirm, verbatim, that they do not support `$filter`,
`$inlinecount`, `$orderby`, `$skip`, `$top`, `$expand`, or `$select` — two separate statements in
SAP's documentation, one per entity, not a single rule this provider generalized across the whole
API family. Number Ranges makes no such statement (and has no GET to apply query options to
regardless).

### Authorization

For Data Stores/Data Store Entries/Variables, SAP documents these Cloud Foundry role templates:
`DataStoresAndQueuesRead` (view data store entries), `DataStorePayloadsRead` (download payloads,
view variables), `DataStoresAndQueuesDelete` (delete data store entries/variables). The existence
of a delete-scoped role template confirms some delete capability is authorized for this data,
consistent with the design-time Delete step described above, even though this provider found no
REST `DELETE` example to call directly.

## Edge Integration Cell

The only hint of an Edge Integration Cell-specific API path in this entire research pass was the
Number Ranges "Runtimes" UI field mentioned above. No `/location/<runtime-location-id>/...`-style
path, or any other Edge-specific endpoint, was found documented for any of the four object types
covered in this guide. This remains genuinely unconfirmed, and nothing Edge-specific is
implemented here as a result — consistent with this provider's rule against guessing at
undocumented endpoints.
