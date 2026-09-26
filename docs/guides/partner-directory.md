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

A **Pid** is the internal Partner Directory identifier a partner's entities are scoped to. For
the `Partners` entity set, SAP's API offers exactly two operations: "You can read all partners or
delete a partner from the Partner Directory." There is no create. SAP states that Pid uniqueness
"is ensured by the tenant owner application", meaning the caller chooses the value, and a Pid
comes into existence the first time a string parameter, binary parameter, alternative partner,
authorized user or user credential parameter references it. The tenant `$metadata` confirms
this: `Partner` has the key `Pid` and no other property.

Deleting a partner removes it "and all its entities", as SAP describes the *Delete Partner*
operation. That is why Partners is modeled only as read-only data sources:

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

SAP lists these `content_type` values: `xml`, `xsl`, `xsd`, `json`, `text`, `zip`, `gz` (GZIP),
`zlib` and `crt` (a DER-encoded X.509 certificate). For the text types (`xml`, `xsl`, `xsd`,
`json`, `text`) an encoding can be appended after a semicolon, for example `xml;encoding=UTF-8`.
The UI's separate *Encoding* field ends up there as well; the API has no property of its own
for it. The provider does not restrict `content_type` to a fixed list, so these combinations
pass through unchanged.

**Size.** SAP's pages disagree about the limit. The API examples say 260 KB, and the entity type
overview says both 262,144 bytes and 1.5 MB. The tenant's `$metadata` declares
`BinaryParameter.Value` with `MaxLength="1572864"`, exactly 1.5 MiB. The provider follows the
service's own metadata: it rejects files above 1,572,864 bytes before uploading and leaves
anything below to SAP. The contract tests check this number against the `$metadata`. If your
tenant still enforces a lower limit, SAP rejects the upload with its own error.

Larger XML/XSL/XSD content can be stored as `zip`. The XML Validator and XSLT Mapping steps
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
OAuth client, or communication user credential itself.

**Write `user` in lowercase.** SAP stores authorized users lowercased. Its own example creates
`"User": "MyUser"` and gets back `myuser`, and SAP notes that filters on `User` must use
lowercase characters (Locale.English). A mixed-case value would be stored differently from what
the configuration says, and Terraform would report an inconsistent result after apply. The
provider therefore rejects uppercase letters at plan time and names the lowercase form to use.
The same applies to the `user` argument of the data source. Integration flows see the lowercase
form too: SAP's scripting API returns authorized users "with lower case characters".

## User credential parameters: a special, security-sensitive case

`sapintegrationsuite_partner_user_credential_parameter` manages a communication
username/password credential scoped to a `partner_id`, consumed by an integration flow through
the generated security artifact alias `pd:<partner_id>:<parameter_id>:UserCredential`. SAP
shows it in the monitor under *Security Material* by that name. This is treated differently
from every other Partner Directory resource in this provider:

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
  plan or state artifacts. SAP returns `Password` as `null` on reads. It can return a SHA-256
  hash with the query option `returnHashedPassword=SHA256`, but the provider never asks for it:
  storing a hash of the password in state would be a leak of its own.
- **`password_wo_version`** is a plain, stored marker you change whenever the password itself
  changes. Terraform can only detect a rotation by diffing something it actually keeps in
  state — the write-only value itself never round-trips, so nothing about it alone would ever
  show up in a plan.
- **Rotation happens in place.** SAP documents that "you can also use the POST request to update
  a User Credentials parameter with the same values for PID and Id", and that PUT is not
  supported. Changing `password_wo_version` or `user` therefore sends that POST with the
  configured user and password. The credential is never deleted in between, so integration
  flows that use it do not fail during a rotation. Only `partner_id` and `parameter_id` force a
  replacement.
- **Create never overwrites.** Because the same POST also overwrites, the provider first checks
  whether the credential exists. If it does, create stops with an error that asks for an import
  instead of replacing a password this configuration does not own.
- **Never batched.** SAP documents that a change set containing a `UserCredentialParameter`
  request may contain only that one request; this provider always issues it on its own.
- **Import.** Importing recovers `partner_id`, `parameter_id` and `user`, never the password.
  The first `terraform apply` after the import plans an in-place update that sends the
  configured password, which also makes sure SAP holds the value the configuration names.

SAP records the technical user of the last POST in `CreatedBy` and `LastModifiedBy`; for this
entity both fields and both timestamps always describe the most recent write.

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

## Changes reach integration flows with a delay

String parameters, binary parameters and authorized users are cached on the runtime nodes.
SAP invalidates a cache entry when it is updated or deleted, but "this invalidation can take a
few minutes", longer when many entries change at once. An `apply` that changes these values
is complete in the Partner Directory right away; integration flows may read the old value for
a few more minutes. Plan tests and cut-overs accordingly.

## Limits

SAP gives tenant-wide maximums: 3,000,000 string parameters, 400,000 binary parameters,
1,000,000 alternative partners and 500,000 authorized users (sized for 10,000 partners). A
string parameter holds up to 4,000 characters. A Pid, like a parameter ID, may contain `A-Z`,
`a-z`, `0-9` and the special characters `-`, `.`, `_`, `~`, `<`, `>` and `@`.

## Known limitations

- No `sapintegrationsuite_partner` resource — see above.
- `sapintegrationsuite_partner_user_credential_parameter` never reads its password back — see
  above.
- The provider does not use the OData `$batch` mass operations SAP offers. Each resource is one
  request, which keeps failures attributable to a single resource.
- The optional `user` query option SAP describes for audit logging is not sent; SAP records the
  technical user of the OAuth client instead.
- CSRF token handling for Partner Directory writes (and every other write this provider makes)
  is handled transparently by the shared HTTP client — see `docs/sap-api-references.md` — and
  requires no configuration.
