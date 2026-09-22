---
page_title: "Security Content"
subcategory: "Security"
description: |-
  What SAP Integration Suite Security Content (security material) is, which artifact types this
  provider manages, its write-only secret model, and why several documented artifact types are
  deliberately not implemented yet.
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

## Deliberately not implemented yet

SAP documents several more Security Content artifact types this provider does not manage today.
Each is recorded in `internal/features/catalog.go` with a specific reason, summarized here:

- **Keystore entries, certificates, key pairs, SSH keys, certificate chains**
  (`security.keystore_entry`, `security.certificate`, `security.key_pair`, `security.ssh_key`,
  `security.certificate_chain`) — SAP's Keystore Monitor UI and the public API catalog confirm
  these exist (`KeystoreEntries`, `Keystores`, `KeystoreResources`, `HistoryKeystoreEntries`,
  `KeyPairGenerationRequests`, `KeyPairResources`, `SSHKeyGenerationRequests`,
  `CertificateChainResources`), but this project could not confirm their exact OData field names
  and casing against a live tenant's `$metadata`. Shipping a guessed field mapping for a security
  artifact was judged a worse outcome than shipping nothing; a read-only keystore entry data
  source is the most likely next step (see `ROADMAP.md`), since read-only guesses fail safely
  (an empty field) rather than risk a bad write to a tenant's keystore.
- **Certificate-User Mapping** (`security.certificate_user_mapping`) — reverified for this
  feature family and **corrected**: SAP's own documentation for certificate-to-user mapping
  ("Managing Certificate-to-User Mappings", "Client Certificate Authentication and
  Certificate-to-User Mapping (Inbound)") exists only for the **Neo** environment. No Cloud
  Foundry equivalent was found. Since this provider targets Cloud Foundry, this feature is marked
  `unsupported` / `no_public_api` — not merely unimplemented, but out of reach for this
  provider's target environment as things stand today.
- **Secure Parameter** (`security.secure_parameter`) — SAP's UI documents creating and deploying
  a Secure Parameter artifact (an opaque confidential value with no associated username), which
  would be a strong write-only-attribute candidate. This project found third-party evidence of a
  practitioner receiving an OData error resolving a `SecureParameters` entity set through the
  Security Content API, which is treated as a signal to reverify before implementing, not as a
  green light to guess.
- **Known Hosts (SSH)** (`security.known_hosts`) — SAP's UI documents uploading/downloading a
  `known_hosts` file for SFTP host key validation, but no confirmed public OData entity set or
  REST endpoint was found for it.
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
  output: the Go client types for both resources structurally have no field to decode a secret
  into, even from a response body that happened to contain one.
- Every acceptance test for these resources uses unmistakably synthetic secrets and a
  `tf-acc-`-prefixed alias, never a real credential, and never touches a pre-existing tenant
  artifact.
