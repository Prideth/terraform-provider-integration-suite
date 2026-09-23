---
page_title: "Security Content"
subcategory: "Security"
description: |-
  What SAP Integration Suite Security Content (security material) is, which artifact types this
  provider manages — credentials, keystore entries, certificates, and SAP-generated key pairs —
  its write-only secret model, alias encoding, SAP-owned entry handling, and why several
  documented artifact types are deliberately not implemented.
---

# Security Content

SAP Cloud Integration's **Security Content** area (also called **Security Material** in the
tenant UI, under *Monitor* > *Manage Security* > *Security Material*) holds the credentials,
certificates, and keys integration flow adapters use for outbound and inbound authentication. It
is backed by a public OData V2 API, documented on SAP Business Accelerator Hub as the "Security
Content" API package, sharing the same `/api/v1` host as the Cloud Integration content APIs this
provider already manages.

This guide covers what this provider implements today, the security model behind every secret
attribute, and — just as importantly — which documented Security Content artifact types this
provider deliberately does not implement yet, and why. See `internal/features/catalog.go` (the
`security.*` entries) for the single source of truth this guide is generated from, and
`docs/feature-support.md` for the full, auto-generated support matrix.

## Implemented: User Credentials and OAuth2 Client Credentials

- [`sapintegrationsuite_user_credential`](../resources/user_credential.md) /
  [`data.sapintegrationsuite_user_credential`](../data-sources/user_credential.md) — a
  username/password credential (SAP's "User Credentials" artifact type), for outbound basic or
  username-token authentication, with optional SuccessFactors/OpenConnectors system binding.
- [`sapintegrationsuite_oauth2_client_credential`](../resources/oauth2_client_credential.md) /
  [`data.sapintegrationsuite_oauth2_client_credential`](../data-sources/oauth2_client_credential.md)
  — a client ID/client secret/token service URL credential (SAP's "OAuth2 Client Credentials"
  artifact type), for the OAuth2 client credentials grant (RFC 6749) on outbound requests.

Both follow the same shape: an identity (`id`, the artifact's name/alias), readable metadata
(`description`, `user`/`client_id`, and so on), and a **write-only** secret plus a plain version
marker that drives rotation.

## The write-only secret model

Sensitive != not stored. Terraform's ordinary `Sensitive: true` attribute flag only masks a
value in CLI output and logs — the value is still written to the Terraform state file and to
plan files, in the clear, for anyone with file access to read. For a credential's password or an
OAuth2 client secret, that is not an acceptable place for the value to live.

Instead, both resources use Terraform's **write-only attributes**
(`password_wo`/`password_wo_version` and `client_secret_wo`/`client_secret_wo_version`):

- The `_wo` attribute is never persisted to plan or state. Terraform sends it to the provider
  only at apply time, from configuration, and the provider must not — and in this provider's
  case, structurally cannot, since the Go client types have no field to decode one into — copy it
  into anything that gets written back to state.
- The paired `_wo_version` attribute *is* stored in state, as an ordinary string. It carries no
  secret material; it exists purely so Terraform has something to compare between plans. Changing
  it is how you tell Terraform "the secret changed, redeploy this credential" — Terraform cannot
  infer that from the write-only value itself, since state never remembers what it was.

This is why every example in this guide bumps `password_wo_version`/`client_secret_wo_version`
whenever `password_wo`/`client_secret_wo` changes, and why a config that keeps referencing the
same `var.something` with the version left unchanged does **not** re-send the secret on every
apply beyond what SAP's own "re-enter the secret on every edit" requirement already forces.

**Requires Terraform CLI 1.11 or later.** Write-only attributes are a Terraform Core / provider
protocol feature that only ships from Terraform CLI 1.11 onward. This provider does not enforce a
`required_version` constraint itself (Terraform module authors set that in their own
configuration), but both write-only resources will fail to plan on an older Terraform CLI. If
your organization pins an older Terraform version, either upgrade it or do not use
`sapintegrationsuite_user_credential`/`sapintegrationsuite_oauth2_client_credential` yet.

## Rotation is an in-place update, not a replacement

SAP's Manage Security Material UI documents an explicit **Edit** action for Credentials
artifacts ("You can also edit and redeploy an existing artifact"), and states that a secret must
be re-entered on every edit. This provider follows that lifecycle: changing
`password_wo_version`/`client_secret_wo_version` (or any other mutable attribute) triggers
Terraform's Update, which this provider implements as a full `PUT` redeploy — not a delete/create
replacement. Only `id` (the artifact's name/alias) and, for user credentials, `kind` (which
system-specific sub-type the credential is) force replacement, since SAP does not document
changing either of those via Edit.

## Drift detection is limited for secrets

Terraform can never detect that a password or client secret changed outside Terraform: SAP's
Security Content API does not return secret values on `GET`, by design, and this provider's Go
client types have no field to receive one even if a future API version did. `Read` only compares
the metadata SAP does return (username, description, token service URL, scope, and so on).
Treat a stored credential's secret as **owned by whoever last set `_wo_version`** — Terraform,
your CI pipeline, or a human editing it directly in the SAP UI — and rotate deliberately by
bumping the version, rather than expecting `terraform plan` to notice an externally-rotated
secret.

## Importing an existing credential

```shell
terraform import sapintegrationsuite_user_credential.backend BACKEND_BASIC
terraform import sapintegrationsuite_oauth2_client_credential.backend BACKEND_OAUTH
```

Import recovers `id` and every readable metadata field. It cannot recover the secret — SAP never
returns one, and this provider's state model has no field for one regardless. Until you add
`password_wo`/`client_secret_wo` and a `_wo_version` to your configuration, Terraform has no
opinion on the secret at all: the imported resource plans cleanly with no password/secret
drift. The first time you add both to your configuration (to take ownership of rotation), that
plans as an ordinary in-place Update — not a replacement — and redeploys the credential with the
password/secret you supplied. This is a deliberate design, not an accident: it means importing a
credential never silently rotates its secret, and taking ownership of rotation is always an
explicit, visible step in a plan.

## Implemented: Keystore Entries, Certificates, and SAP-generated Key Pairs

The tenant keystore (*Monitor* > *Manage Security* > *Keystore*) holds certificates and key
pairs, distinct from the credentials above. This provider manages it through three types:

- [`data.sapintegrationsuite_keystore_entry`](../data-sources/keystore_entry.md) /
  [`data.sapintegrationsuite_keystore_entries`](../data-sources/keystore_entries.md) — read-only
  discovery of any keystore entry (certificate, SAP-generated key pair, or other RSA/DSA/EC-keyed
  entry) by alias, or every entry in the tenant keystore.
- [`sapintegrationsuite_certificate`](../resources/certificate.md) — manages a standalone X.509
  certificate (for example a partner's or CA's public certificate you trust).
- [`sapintegrationsuite_key_pair`](../resources/key_pair.md) — generates an SAP-managed key pair;
  the private key is created and retained by SAP and never enters this provider at all.

There is deliberately **no generic `sapintegrationsuite_keystore_entry` resource**: that entity
represents fundamentally different object types with different lifecycles, so a mutable resource
covering all of them would either be too vague to be safe or would have to reimplement
Certificate's and Key Pair's logic behind one confusing interface. Use the specific resource for
what you are actually managing.

### Alias encoding — you never touch it

Every keystore entry's real OData identity is a lowercase hex encoding of its alias's UTF-8
bytes (SAP's own documentation states this explicitly, and explains why: "the server doesn't
allow slashes or backslashes in a URI, even if they're percent encoded"). You never compute or
supply this yourself — every resource and data source here takes a plain `alias` and encodes it
internally (`internal/client/odata/v2/hexkey.go`), the same shared helper Partner Directory's
`AlternativePartners` key encoding was refactored to use once this provider confirmed both
follow the identical rule. It is asserted against every alias SAP's own documentation uses as a
worked example, plus Unicode, punctuation, semicolons, slashes, and backslashes.

### `sapintegrationsuite_certificate`

```hcl
resource "sapintegrationsuite_certificate" "backend_ca" {
  alias       = "backend-root-ca"
  certificate = file("${path.module}/backend-root-ca.pem")
}
```

`certificate` is PEM content and is **never marked `Sensitive`** — a public X.509 certificate is
not a secret, and conflating it with private key material (which this resource never handles at
all) would be a mistake. Create and Update both use the same confirmed SAP operation (`PUT
CertificateResources('<hexalias>')/$value`, which SAP's own documentation explicitly notes
creates a new entity despite the PUT verb).

**Drift detection compares a canonical fingerprint, not raw PEM text.** Two PEM encodings of the
identical certificate can differ in line endings, wrapping, or a trailing newline without
representing any real change. On every `Read`, this provider parses both the certificate
currently on the tenant and whatever is already in your Terraform state using Go's own
`crypto/x509`, and only replaces your state's PEM text with SAP's own re-serialization when the
SHA-256 fingerprint of the certificate's actual DER bytes has genuinely changed. `subject_dn`,
`issuer_dn`, `serial_number`, and `certificate_sha256` are all derived this same way — locally,
from certificate bytes this provider can confirmedly retrieve — never from a guessed SAP field
name (SAP's own documented example response for a keystore entry is truncated before these
properties are shown).

Delete uses SAP's documented keystore mass-deletion operation
(`KeystoreResources('system')?deleteEntries=true`) with **exactly the one alias this resource
owns** — never a caller-assembled list, so a `terraform destroy` here can never accidentally
reach an unrelated alias.

### `sapintegrationsuite_key_pair`

```hcl
resource "sapintegrationsuite_key_pair" "client_auth" {
  alias        = "client-auth-key"
  key_type     = "RSA"
  key_size     = 2048
  common_name  = "client.example.com"
  country      = "DE"
  organization = "Example GmbH"
}
```

SAP generates the private key internally; **it is never downloadable through this API, and this
resource has no field for it anywhere** — not write-only, not sensitive, simply absent, because
no code path in this provider ever asks SAP for one.

`key_type` is one of `RSA` (default), `DSA`, `EC` — SAP's complete, fixed enum. `key_size` is
required for `RSA`/`DSA`; for `EC`, either `key_size` (112–571) or `key_algorithm_parameter` (a
named curve, for example `secp256r1`) is required — enforced at plan time, since the rule is
conditional on `key_type`. `signature_algorithm`, if set, is validated against the exact
SAP-documented enum for your chosen `key_type` (RSA/DSA/EC each have their own list).

**No update operation is documented** for a generated key pair's material — every attribute that
defines it (`key_type`, `key_size`, the subject DN fields, validity dates, and so on) is
`RequiresReplace`. Changing any of them generates an entirely new key pair under the same alias
lifecycle (replace), never an in-place mutation of existing key material.

**Read is only partially confirmed.** SAP's `KeystoreEntries` GET confirms `key_type`,
`key_size`, `valid_not_before`, and `valid_not_after` back; `signature_algorithm`,
`key_algorithm_parameter`, and every subject DN field are not confirmed returned by any
documented GET. This provider refreshes what it can confirm on every plan and trusts the rest
from the last successful write — which is why this resource's support status is `partial`, not a
statement that anything about it is unsafe to use.

**There is no separate "SSH Key" resource.** SAP's own Security Content API overview lists no
independent SSH Key entity, and the tenant keystore UI's own "Creating a Key Pair/SSH Key Pair"
documentation uses the identical field set for both — "Create > Key Pair" and "Create > SSH Key"
are the same underlying mechanism with a different label. An RSA or DSA
`sapintegrationsuite_key_pair`'s public key is available in OpenSSH format directly:

```hcl
output "client_auth_ssh_public_key" {
  value = sapintegrationsuite_key_pair.client_auth.public_key_openssh
}
```

`public_key_openssh` is populated via SAP's confirmed
`KeystoreEntries('<hexalias>')/Sshkey/$value` export whenever `key_type` is `RSA` or `DSA` (SAP
documents EC as unsupported for this export); it stays `null` for an EC key pair.

Delete uses the same single-alias mass-deletion mechanism as `sapintegrationsuite_certificate`.

### SAP-owned keystore entries

A tenant keystore typically contains entries the tenant administrator owns and entries SAP owns
(for example SAP's own root certificates). **No API field distinguishes the two** — this project
searched thoroughly and found none, only prose confirming the distinction exists. This provider
does not guess at ownership with a heuristic (matching alias name patterns, for example); instead
`sapintegrationsuite_certificate` and `sapintegrationsuite_key_pair` simply attempt the operation
you asked for, and SAP's own server-side protection rejects an Update or Delete against a
protected entry with an ordinary API error, which this provider surfaces to you exactly as it
does any other SAP error. If you import an SAP-owned alias into either resource, the first
Update or Delete you attempt against it will fail with SAP's own rejection — this provider cannot
warn you sooner, since there is nothing in the API to warn from.

### Whole-keystore management stays out of scope

SAP's `KeystoreResources` entity also supports importing an entire keystore (`POST
KeystoreResources`, a base64-encoded JKS/JCEKS file plus password) and backing up/restoring sets
of entries. This provider deliberately does not implement a `sapintegrationsuite_keystore`
resource around that operation, confirmed contract or not: a single import call can create,
update, leave unchanged, or remove many entries at once, based entirely on the uploaded file's
contents, with no way for this provider to know whether any given affected entry belongs to a
different Terraform module, a different administrator, or SAP itself. This is the same
blast-radius concern this provider already avoids for the confirmed mass-delete operation, which
is why Certificate and Key Pair only ever call it with exactly the one alias they own, never an
open-ended list. If you need whole-keystore import/export/backup, use the SAP Integration Suite
UI.

## Deliberately not implemented

SAP documents several more Security Content artifact types this provider does not manage. Each
is recorded in `internal/features/catalog.go` with a specific reason, summarized here:

- **Certificate Chain** (`security.certificate_chain`) — reverified for this feature family: SAP
  documents certificate chain import/export as a *capability of* the Key Pair resource ("create a
  certificate signing request, or import and export the related certificate chain"), not an
  independently exampled entity — no `CertificateChainResources` example request or field
  contract was found documented anywhere. If a concrete contract is confirmed later, this would
  likely be scoped by key-pair alias (`sapintegrationsuite_key_pair_certificate_chain`), not a
  standalone global resource.
- **Certificate-User Mapping** (`security.certificate_user_mapping`) — reverified for this
  feature family and **corrected**: SAP's own documentation for certificate-to-user mapping
  ("Managing Certificate-to-User Mappings", "Client Certificate Authentication and
  Certificate-to-User Mapping (Inbound)") exists only for the **Neo** environment. No Cloud
  Foundry equivalent was found. Since this provider targets Cloud Foundry, this feature is marked
  `unsupported` / `no_public_api` — not merely unimplemented, but out of reach for this
  provider's target environment as things stand today.
- **Secure Parameter** (`security.secure_parameter`) — reverified for this feature family: SAP's
  own Security Content API overview conceptually lists "Secure Parameter" as a resource, but the
  curated, authoritative example-requests index for this API — the same page that correctly
  enumerates every other confirmed operation — lists zero worked examples for it, and its own
  deployment documentation describes only the Eclipse/Node-Explorer design-time wizard, not a
  REST contract. Combined with prior third-party evidence of an OData error resolving a
  `SecureParameters` entity set, this stays unconfirmed. If a public contract is ever confirmed,
  this would be a strong write-only-attribute candidate (`value_wo`/`value_wo_version`), the same
  shape as the credential resources' passwords above.
- **Known Hosts (SSH)** (`security.known_hosts`) — reverified and **strengthened**: unlike Secure
  Parameter, Known Hosts does not appear in SAP's Security Content API overview's resource table
  at all. Its own deployment documentation describes only the Manage Security Material UI, with
  no REST endpoint mentioned anywhere. Corrected from "research required" to "no public API
  found."
- **OAuth2 Authorization Code** — requires an interactive human authorization step by its nature
  (SAP's UI has a dedicated *Authorize* action with a status of *Unauthorized* until a person
  completes it), which does not fit Terraform's non-interactive plan/apply model. This is
  considered out of scope on safety grounds, not a research gap.
- **OAuth2 SAML Bearer Assertion** — investigated only briefly; this project could not confirm a
  public, safe-to-automate lifecycle for it and did not want to generalize from OAuth2 Client
  Credentials without evidence.
- **PGP keyrings** (public and secret) — a PGP *secret* keyring is private key material; this
  project deliberately did not pursue implementing it without a much higher bar of confirmed API
  safety than the artifact types above, consistent with this provider's general stance that
  security takes priority over completeness.

If you need any of these today, the SAP Integration Suite UI remains the correct tool; this
provider will not guess at an unconfirmed API contract for a security-sensitive artifact.

## Security notes

- Passwords and client secrets never appear in this provider's logs, error diagnostics, or debug
  output: the Go client types for both credential resources structurally have no field to decode
  a secret into, even from a response body that happened to contain one.
- `sapintegrationsuite_key_pair` never requests, stores, or exposes a private key: no code path
  in this provider calls an operation that could return one, and no field exists anywhere in the
  schema or the Go client types to receive one.
- Error diagnostics for certificate and key pair operations identify the alias and SAP's own
  status/error, never a full certificate dump or the request/response body verbatim — a public
  certificate is not confidential, but there is no reason to put its full content in an error
  message either.
- Every acceptance test for these resources uses unmistakably synthetic secrets/certificates and
  a `tf-acc-`-prefixed alias, never a real credential or a customer certificate, and never
  touches a pre-existing tenant artifact (`sap_*`, `hcicertificate*`, or any alias not created by
  the test itself).
