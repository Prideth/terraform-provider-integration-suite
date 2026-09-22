---
page_title: "Partner Directory"
subcategory: "Partner Directory"
description: |-
  What SAP Integration Suite's Partner Directory is, its entity model, and how this provider
  manages it.
---

# Partner Directory

SAP Integration Suite's **Partner Directory** stores B2B/trading-partner configuration —
addresses, endpoints, schemas, certificates, alternate identifiers, and authorized users —
scoped to a **Partner ID (Pid)**, so integration flows can look values up per trading partner at
runtime instead of hardcoding them. This guide covers the entity model, what this provider
manages, and — just as importantly — what it deliberately does not.

## Entity model and Terraform ownership

```
Partner Directory
│
├── Partner (Pid)                       — read-only (data sources only)
│
├── String Parameter                    — sapintegrationsuite_partner_string_parameter
├── Binary Parameter                    — sapintegrationsuite_partner_binary_parameter
├── Alternative Partner                 — sapintegrationsuite_alternative_partner
├── Authorized User                     — sapintegrationsuite_partner_authorized_user
└── User Credential Parameter           — sapintegrationsuite_partner_user_credential_parameter
```

Every entity below a Pid is an **independent Terraform resource**. There is deliberately no
`sapintegrationsuite_partner` resource that owns a partner's string parameters, binary
parameters, alternative partners, and authorized users as a single block — a design like that
would make destroying the parent resource capable of deleting every child entity for that Pid,
including ones a completely different Terraform module or a manual process created. Managing
one `sapintegrationsuite_partner_string_parameter` never touches any other entity that happens
to share the same `partner_id`.

## Partner IDs (Pid) have no resource

A **Pid** is the internal Partner Directory identifier a partner's entities are scoped to. SAP
documents no confirmed create operation for the `Partners` entity itself, and states that Pid
uniqueness "is ensured by the tenant owner application" — in other words, a Pid comes into
existence implicitly the first time a string parameter, binary parameter, alternative partner,
authorized user, or user credential parameter is created that references it. There is nothing
to `POST` to create a Partner on its own.

SAP additionally documents that deleting a Pid can remove every entity belonging to it in a
single call. Combined with the lack of a confirmed create operation, this is why Partners is
modeled only as read-only data sources:

- `data.sapintegrationsuite_partner` confirms whether a given Pid exists.
- `data.sapintegrationsuite_partners` lists every Pid known to the tenant — the primary
  brownfield discovery mechanism for this provider's Partner Directory support, since there is
  no resource to enumerate instances of otherwise.

A `sapintegrationsuite_partner` resource would need a `Destroy` that either does nothing useful
(no create means nothing to reliably delete just by itself) or cascades to child content this
provider — or another Terraform module — does not own. Neither is an acceptable Terraform
lifecycle, so no such resource exists.

## String and binary parameters

`sapintegrationsuite_partner_string_parameter` and `sapintegrationsuite_partner_binary_parameter`
manage individual named values scoped to a `partner_id`, identified by their own
`parameter_id`. Both support full create/read/update/delete via SAP's documented API (PUT for
updates, replacing the parameter's value in place).

```hcl
resource "sapintegrationsuite_partner_string_parameter" "receiver_address" {
  partner_id   = "PartnerZ"
  parameter_id = "ReceiverAddress"
  value        = "https://receiver.example.com"
}

resource "sapintegrationsuite_partner_binary_parameter" "order_schema" {
  partner_id   = "PartnerZ"
  parameter_id = "OrderSchema"

  content_type = "xsd"
  content      = "${path.module}/partner-directory/order-schema.xsd"
  content_hash = filesha256("${path.module}/partner-directory/order-schema.xsd")
}
```

Import uses `<partner_id>/<parameter_id>`:

```shell
terraform import sapintegrationsuite_partner_string_parameter.receiver_address PartnerZ/ReceiverAddress
terraform import sapintegrationsuite_partner_binary_parameter.order_schema PartnerZ/OrderSchema
```

### Binary parameter content types and size

`content_type` accepts any value — it is not restricted to a fixed list, because SAP's
documentation names common values (`xml`, `xsl`, `xsd`, `json`, `text`, `zip`, `gz`, `zlib`,
`crt`) while also accepting encoding-suffixed variants such as `xml;encoding=UTF-8` that a short
fixed list would incorrectly reject.

SAP documents a **260 KB maximum** for a binary parameter's decoded value. This provider checks
that limit locally before ever sending the request, so an oversized file fails fast with a clear
message rather than an opaque API error. For XML/XSL/XSD content larger than that uncompressed,
SAP recommends storing it as a `zip` — the XML Validator and XSLT Mapping steps automatically
unzip a binary parameter whose `content_type` is `zip`.

### Do not store secrets here

**SAP explicitly documents that ordinary Partner Directory data is stored unencrypted.** Do not
put passwords, private keys, tokens, or other sensitive information into a string or binary
parameter's value. If you need a partner-scoped credential, see
`sapintegrationsuite_partner_user_credential_parameter` below — and even that has real caveats
of its own.

## Alternative partners

`sapintegrationsuite_alternative_partner` maps an external identity tuple — `agency`, `scheme`,
`external_id` — to an internal `partner_id`, letting inbound messages be resolved to a Pid by an
alternate identifier scheme instead of your own.

```hcl
resource "sapintegrationsuite_alternative_partner" "sender" {
  agency      = "Sender_1"
  scheme      = "SenderInterface"
  external_id = "Interface_1"
  partner_id  = "CS_Scenario_1"
}
```

SAP's actual OData entity key for this mapping is not the plain `agency`/`scheme`/`external_id`
strings but their **hex-encoded** form (`Hexagency`/`Hexscheme`/`Hexid`) — confirmed against a
documented SAP example where the agency `"agency1"` becomes the key component
`"6167656e637931"`. This provider computes that encoding internally; you never see or set a hex
value in configuration.

Import uses the three hex-encoded segments joined by `/`, since hex text can never itself
contain a `/` regardless of what characters the plain values contain:

```shell
terraform import sapintegrationsuite_alternative_partner.sender \
  53656e6465725f31/53656e646572496e74657266616365/496e746572666163655f31
```

`partner_id` can be changed in place (SAP documents `PUT` for repointing an existing mapping at
a different Pid); `agency`, `scheme`, and `external_id` together are the mapping's identity and
changing any of them replaces the resource.

## Authorized users

`sapintegrationsuite_partner_authorized_user` maps a communication `user` to the `partner_id` it
is authorized to act as for inbound communication. SAP documents this as many-to-one: a
communication user maps to exactly one Pid, while a Pid can have several authorized users.

```hcl
resource "sapintegrationsuite_partner_authorized_user" "commuser" {
  user       = "commuser1"
  partner_id = "PartnerZ"
}
```

This resource manages only the Partner Directory mapping — never the underlying BTP user,
OAuth client, or communication user credential itself. Whether SAP normalizes `user`'s case
internally was not confirmed against a primary source during this feature's research, so this
provider passes the value through exactly as configured rather than guessing at a normalization
rule. If your tenant does normalize case and you see persistent drift, configure `user` in
whatever case SAP actually stores.

## User credential parameters: a special, security-sensitive case

`sapintegrationsuite_partner_user_credential_parameter` manages a communication
username/password credential scoped to a `partner_id`, consumed by an integration flow through
the generated security artifact alias `pd:<partner_id>:<parameter_id>:UserCredential`. This is
treated differently from every other Partner Directory resource in this provider:

```hcl
variable "receiver_communication_password" {
  type      = string
  sensitive = true
}

resource "sapintegrationsuite_partner_user_credential_parameter" "receiver" {
  partner_id   = "Receiver_1"
  parameter_id = "USER"
  user         = "commuser1"

  password_wo         = var.receiver_communication_password
  password_wo_version = "1"
}
```

- **`password_wo` is write-only** (requires Terraform CLI 1.11+): Terraform never stores it in
  plan or state artifacts. This provider also never requests or reads a password back from SAP,
  which does not document an API for returning one.
- **`password_wo_version`** is a plain, stored marker you change whenever the password itself
  changes. Terraform can only detect a rotation by diffing something it actually keeps in
  state — the write-only value itself never round-trips, so nothing about it alone would ever
  show up in a plan.
- **No in-place update.** No public API for changing an existing credential's password was
  confirmed, so every field — including `password_wo_version` — forces replacement: rotating a
  password deletes the old credential and creates a new one, rather than guessing at a
  `PUT`/`PATCH` this provider could not verify.
- **Never batched.** SAP documents that `UserCredentialParameter` (and `CertificateUserMapping`)
  cannot be combined with other Partner Directory entity types in a single OData ChangeSet
  request; this provider always issues it as a standalone request.
- **Import has a real gap.** Importing recovers `partner_id`, `parameter_id`, and `user`, but
  never the password — there is nothing to recover it from. The first `terraform apply` after
  import, once you supply `password_wo` and `password_wo_version`, plans as a replacement even
  though nothing has actually changed server-side. This is an inherent limitation of adopting a
  write-only-secret resource, not a bug.

This is reflected in the feature catalog as **partial** support, not full: the write and delete
lifecycle is solid, but there is no update and no read-back, and that is a deliberate, permanent
property of this resource's security model — not something a future release is expected to
"complete."

## Permissions

SAP documents the Cloud Foundry role template **`AuthGroup_TenantPartnerDirectoryConfigurator`**
as required for Partner Directory read/write access through the OData API (the
`AuthGroup_Administrator` role also works). This provider does not manage that role or the BTP
role collection assigning it — configure it directly in SAP BTP cockpit, as a prerequisite to
using any Partner Directory resource or data source here.

## Pagination

`data.sapintegrationsuite_partners` and `data.sapintegrationsuite_partner_string_parameters`
follow SAP's server-driven paging (`__next` links) to return the complete result set, not just
the first page — Partner Directory entity sets, String Parameters especially, are documented as
capable of holding large numbers of entries per tenant.

## Known limitations

- No `sapintegrationsuite_partner` resource — see above.
- `sapintegrationsuite_partner_user_credential_parameter` has no in-place update and no
  read-back of its password — see above.
- Whether `AuthorizedUsers.User` is case-normalized by SAP internally is unconfirmed; this
  provider does not normalize it.
- The exact wire-format casing rules for `BinaryParameters.ContentType` beyond SAP's documented
  example values are unconfirmed; this provider does not restrict the attribute to a fixed list
  because of this.
- CSRF token handling for Partner Directory writes (and every other write this provider makes)
  is handled transparently by the shared HTTP client — see `docs/sap-api-references.md` — and
  requires no configuration.
