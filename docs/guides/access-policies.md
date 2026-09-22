---
page_title: "Access Policies"
subcategory: "Security"
description: |-
  What SAP Integration Suite access policies are, how they relate to BTP roles, and how this
  provider manages them.
---

# Access Policies

An SAP Integration Suite **access policy** restricts which artifacts (integration flows,
OData/REST/SOAP APIs, script collections, value mappings, message mappings, message queues,
global data stores, and global variables) a given role can see and operate on. This guide
covers what this provider manages, what it deliberately does not, and the SAP-side concepts
you need to reason about when writing Terraform for access policies. See
`docs/resource-design.md` and `docs/sap-api-references.md` in this repository's source for the
full research trail behind the statements below.

## The two resources

This provider models an access policy as two separate resources:

- [`sapintegrationsuite_access_policy`](../resources/access_policy.md) — the policy itself:
  a `role_name` and a `description`.
- [`sapintegrationsuite_access_policy_reference`](../resources/access_policy_reference.md) —
  one artifact-matching rule attached to a policy. A policy typically has several of these.

They are separate resources, not a nested block, because each reference has its own
server-assigned ID and independent create/delete lifecycle — see `docs/resource-design.md`
§17 for the full suitability analysis. A minimal example:

```hcl
resource "sapintegrationsuite_access_policy" "utilities_architect" {
  role_name   = "UTILITIES_ARCHITECT"
  description = "Access policy for utilities architecture artifacts"
}

resource "sapintegrationsuite_access_policy_reference" "integration_flows" {
  access_policy_id = sapintegrationsuite_access_policy.utilities_architect.id

  artifact_type = "IntegrationFlow"
  attribute     = "Id"
  operator      = "MATCHES"
  value         = "UTIL_.*"
}
```

`operator = "MATCHES"` requires **a valid Java Regular Expression** (the same syntax
`java.util.regex.Pattern` accepts) — not a wildcard or glob pattern. `"UTIL_.*"` above matches
any ID starting with `UTIL_`. A value like `"UTIL_*"` is also valid Java regex syntax, but
means "`UTIL` followed by zero or more trailing underscores" rather than "`UTIL_` followed by
anything" — easy to misread as a glob, so prefer the explicit `.*` form.

## What `role_name` is (and isn't)

`role_name` identifies a role on the SAP BTP side, associated with the access policy through
the **SAP BTP cockpit** — not this provider, and not the SAP Integration Suite UI. SAP's own
documentation describes this as defining a role (for example, a custom role built from a role
template in your subaccount's Security area) and associating it with the policy in BTP
cockpit; the artifacts the policy protects then become visible only to users who hold that
role, typically through a role collection that includes it.

This provider **does not** manage that BTP-side role or the role collection that grants it to
users — see `docs/provider-scope.md`. `role_name` is an opaque string this provider passes
through unchanged; if the underlying BTP role does not exist, SAP will still create the access
policy (the association is separate), and it is your responsibility to keep the two in sync.
`role_name` forces resource replacement on change, since SAP does not document renaming a
policy's role in place.

## Supported artifact types, attributes, and operators

`artifact_type` on a reference accepts exactly the values SAP documents as supported for
access policies: `IntegrationFlow`, `ODataAPI`, `RestAPI`, `SoapAPI`, `ScriptCollection`,
`ValueMapping`, `MessageMapping`, `MessageQueue`, `GlobalDataStore`, `GlobalVariable`. Notably,
`IntegrationPackage` is **not** a valid artifact type here (SAP explicitly documents that
access policies cannot be scoped to a whole package).

`attribute` accepts `Name` or `Id` — match by the artifact's display name or its technical ID.

`operator` accepts `EQUALS` (exact match against `value`) or `MATCHES` (Java-regex match
against `value`, see above).

The exact-set membership of these three enumerations is confirmed against SAP's own
documentation; the precise wire-format casing SAP's OData API expects has not been confirmed
against a live tenant or `$metadata` document, so this provider's validators only accept the
casing shown above. If your tenant's API rejects a value with this exact casing, please open an
issue with the API's error response — that is exactly the kind of primary-source evidence this
project needs to correct it.

## Runtime replication and reconciliation

SAP's "Manage Access Policies" documentation describes replicating a policy to one or more
runtimes — the Cloud Integration runtime, Integration Cell, and Edge Integration Cell are all
named as valid targets — and shows a per-runtime reconciliation status of `Fail`, `Success`,
or `Pending`. This is a real Integration Suite capability, not a documentation gap in this
project's own imagination.

What this provider could **not** confirm is whether that capability is exposed through the
public `AccessPolicies` OData API it uses, as distinct from being specific to the Integration
Suite application UI. Because of that, this provider:

- surfaces `reconciliation_status` on `sapintegrationsuite_access_policy` as a best-effort,
  `Computed`-only field, populated only if and when the API happens to return it;
- does **not** poll it to a terminal state during create or update;
- does **not** offer any resource for managing runtime replication, Integration Cell
  configuration, or Edge Integration Cell configuration — these remain tracked in
  `ROADMAP.md` as blocked on a confirmed public API.

Treat `reconciliation_status`, when present, as informational only. Do not build automation
that depends on its exact values or its presence.

## Ownership boundaries and brownfield drift

- `sapintegrationsuite_access_policy` never deletes or modifies references it does not itself
  manage. If a reference is created outside Terraform (through the UI or another automation),
  managing the parent policy with this provider has no effect on it — Terraform only acts on
  the specific reference IDs it created or that you have separately imported.
- Both resources support `terraform import`:
  - `terraform import sapintegrationsuite_access_policy.utilities_architect <id>`
  - `terraform import sapintegrationsuite_access_policy_reference.integration_flows <access_policy_id>/<reference_id>`
- If a policy or reference is deleted outside Terraform, the next `terraform plan` detects it
  as removed and proposes recreating it, the same as any other resource in this provider.
- If a reference's match condition (`artifact_type`, `attribute`, `operator`, or `value`) is
  changed outside Terraform, `terraform plan` shows the drift as a proposed replacement — SAP
  does not document an in-place update for these fields, so this provider does not attempt
  one.
- `data.sapintegrationsuite_access_policy` and `data.sapintegrationsuite_access_policy_reference`
  let you read a policy or reference this provider does not manage — for example, to attach new
  references to a policy someone else's automation owns.

## Known limitations

- No update-in-place for any field of a reference; every field change replaces the resource.
- `reconciliation_status` is best-effort and not polled; see above.
- No support for listing/pagination of access policies as a data source yet — if your tenant
  has few enough policies that this matters, filter with HCL `for_each`/`locals` over IDs you
  already know, or open an issue describing the use case for a
  `sapintegrationsuite_access_policies` list data source.
- This provider does not manage the BTP-side role or role collection referenced by
  `role_name` — use the official `SAP/btp` Terraform provider for that.
