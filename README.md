# Terraform Provider for SAP Integration Suite

A Terraform provider for configuring, provisioning, and administering
content and capabilities **inside an already-provisioned** SAP Integration
Suite tenant — Cloud Integration, Access Policies, API Management, and (as
their public APIs are confirmed) the newer API Gateway model, Integration
Cell, and Edge Integration Cell.

> **This project is an independent open-source Terraform provider and is
> not an official SAP product**, unless and until SAP formally adopts or
> publishes it.

## Status

Pre-release, under active development toward `v0.1.0`. Schemas may still
change. See `ROADMAP.md` for what is planned and `CHANGELOG.md` for what
has landed.

**Repository naming**: the GitHub repository, Go module, provider binary
name, and registry manifest all use the final naming
(`github.com/Prideth/terraform-provider-sap-integration-suite`,
`Prideth/sap-integration-suite`, `sapintegrationsuite`). `master` is this
repository's permanent stable/default branch; feature work happens on
short-lived branches off `master` (for example `provider-hardening`).

## Provider scope

This provider manages **content and capabilities inside** an Integration
Suite tenant. It does **not** create BTP subaccounts, entitlements, the
Integration Suite subscription, generic BTP destinations, role collections,
or anything else owned by the SAP BTP control plane — use the official
[`SAP/btp`](https://registry.terraform.io/providers/SAP/btp/latest)
provider for that. See [`docs/provider-scope.md`](docs/provider-scope.md)
and [`docs/provider-boundaries.md`](docs/provider-boundaries.md) for the
full picture.

```
SAP/btp                              Prideth/sap-integration-suite
  Subaccount                           Access Policies
  Entitlements                         Integration Packages
  Integration Suite subscription  -->  Integration Flows
  Service instances / bindings         Integration Flow Deployments
  Destinations                         (API Gateway, Integration Cell,
  Role collections / assignments        Edge Integration Cell as their
                                         public APIs are confirmed)
```

## Supported capabilities (v0.1.0)

| Resource / Data Source | Purpose |
|---|---|
| `sapintegrationsuite_integration_package` | Cloud Integration content package |
| `sapintegrationsuite_integration_flow` | Integration flow design-time content (file-based) |
| `sapintegrationsuite_integration_flow_deployment` | Integration flow runtime deployment |
| `sapintegrationsuite_access_policy` | Access policy |
| `sapintegrationsuite_access_policy_reference` | A single artifact reference on an access policy |
| `sapintegrationsuite_value_mapping` | Value mapping design-time content (file-based) |
| `sapintegrationsuite_value_mapping_deployment` | Value mapping runtime deployment |
| `sapintegrationsuite_message_mapping` | Reusable message mapping design-time content (file-based) |
| `sapintegrationsuite_message_mapping_deployment` | Message mapping runtime deployment |
| `data.sapintegrationsuite_message_mapping` | Read-only lookup of an existing message mapping |

Every resource here is backed by a currently documented, SAP-supported
public API — see [`docs/sap-api-references.md`](docs/sap-api-references.md)
for the exact API behind each one, and
[`docs/api-capability-matrix.md`](docs/api-capability-matrix.md) /
[`docs/provisioning-capability-matrix.md`](docs/provisioning-capability-matrix.md)
for the full discovery behind what is and is not implemented yet.

## Feature support

This provider ships a machine-readable feature support catalog, queryable
directly from Terraform with no SAP host or credentials — it describes
what this provider *version* implements, not what is active in any
particular tenant:

```hcl
data "sapintegrationsuite_provider_features" "all" {}

data "sapintegrationsuite_provider_feature" "value_mapping" {
  key = "cloud_integration.value_mapping"
}
```

See [`docs/feature-support.md`](docs/feature-support.md) for the full,
generated table of every feature this project has evaluated — supported,
partial, or not implemented, and why.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5
- An SAP Integration Suite tenant with Cloud Integration activated
- An OAuth 2.0 client credentials service key with the Integration Content /
  Security Content API scopes

## Installation

```hcl
terraform {
  required_providers {
    sapintegrationsuite = {
      source  = "Prideth/sap-integration-suite"
      version = "~> 0.1"
    }
  }
}
```

## Authentication

```hcl
provider "sapintegrationsuite" {
  host = var.integration_suite_host

  oauth {
    token_url     = var.integration_suite_token_url
    client_id     = var.integration_suite_client_id
    client_secret = var.integration_suite_client_secret
  }
}
```

Or via environment variables:

```shell
export SAP_INTEGRATION_SUITE_HOST="https://<tenant>.it-cpi<...>.cfapps.<region>.hana.ondemand.com"
export SAP_INTEGRATION_SUITE_TOKEN_URL="https://<tenant>.authentication.<region>.hana.ondemand.com/oauth/token"
export SAP_INTEGRATION_SUITE_CLIENT_ID="..."
export SAP_INTEGRATION_SUITE_CLIENT_SECRET="..."
```

## Example

```hcl
resource "sapintegrationsuite_integration_package" "utilities" {
  id          = "UTILITIES"
  name        = "Utilities Integration"
  description = "Integration content for the utilities line of business"
}

resource "sapintegrationsuite_integration_flow" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = "metering"
  name       = "Metering"

  content      = "${path.module}/iflows/metering.zip"
  content_hash = filesha256("${path.module}/iflows/metering.zip")
}

resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = sapintegrationsuite_integration_flow.metering.flow_id
}
```

See [`examples/greenfield`](examples/greenfield) for a full new-landscape
example and [`examples/brownfield`](examples/brownfield) for importing an
existing one.

## Import

Every resource is importable using its documented ID format, for example:

```shell
terraform import sapintegrationsuite_integration_package.utilities UTILITIES
terraform import sapintegrationsuite_integration_flow.metering UTILITIES/metering
terraform import sapintegrationsuite_access_policy.utilities <access-policy-id>
```

See each resource's page under `docs/resources/` for its exact ID format.

## Development

```shell
go build ./...
go test ./...
make lint   # golangci-lint
make docs   # regenerate docs/ from schema + examples
```

## Testing

- **Unit tests** (`go test ./...`) run without credentials and cover the
  HTTP client, OAuth token handling, OData V2 request/pagination/error
  handling, and the API clients, using `httptest`.
- **Acceptance tests** exercise a real tenant and only run with `TF_ACC=1`
  and `SAP_INTEGRATION_SUITE_*` credentials set:

  ```shell
  TF_ACC=1 go test -v -timeout 60m ./...
  ```

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md). Every new resource must trace back
to an officially documented, SAP-supported public API — see
[`docs/sap-api-references.md`](docs/sap-api-references.md) for the pattern.

## Roadmap

See [`ROADMAP.md`](ROADMAP.md).

## Known limitations

- No public API for Integration Suite capability activation (Cloud
  Integration, API Management, Integration Cell, Edge Integration Cell) was
  found; activation stays a manual, one-time bootstrap step. See
  `docs/provisioning-capability-matrix.md`.
- `sapintegrationsuite_integration_flow`, `sapintegrationsuite_value_mapping`,
  and `sapintegrationsuite_message_mapping`'s `content`/`content_hash`
  cannot be populated by `terraform import`, since SAP does not return a
  local file path for an existing design-time artifact; apply a matching
  configuration after import to bring content under management.
- `sapintegrationsuite_value_mapping` has no in-place update: changing
  `name`, `content`, or `content_hash` replaces the resource (creates a new
  artifact, then deletes the old one) rather than calling an unverified
  `PUT`. SAP separately documents a distinct
  `ValueMappingDesigntimeArtifactSaveAsVersion` action this provider does
  not yet use; implementing true in-place update through it is deferred to
  v0.2.x. See `docs/sap-api-references.md`.
  `sapintegrationsuite_message_mapping` does not share this limitation — it
  has a confirmed in-place update via `PUT`, on different, entity-specific
  evidence (see `docs/sap-api-references.md`).
- `sapintegrationsuite_message_mapping` is the reusable, package-level
  message mapping artifact, not the inline/local message mapping step
  configurable directly inside an integration flow — see
  `docs/resource-design.md` for the distinction.
- Individual value mapping entries (`UpsertValMaps`, `UpdateDefaultValMap`,
  `DeleteValMaps`) are not yet manageable through this provider — only the
  design-time artifact as a whole. See `docs/resource-design.md` for why.
- Whether Delete removes only the active version or every version of the
  artifact is unconfirmed for both `sapintegrationsuite_value_mapping` and
  `sapintegrationsuite_message_mapping` — see `docs/sap-api-references.md`.
- Script collections, API Gateway / API Artifacts, classic API Management,
  Integration Cell, and Edge Integration Cell resources are not yet
  implemented — see `ROADMAP.md`.

## API support matrix

See [`docs/api-capability-matrix.md`](docs/api-capability-matrix.md) and
[`docs/provisioning-capability-matrix.md`](docs/provisioning-capability-matrix.md).

## Disclaimer

This project is an independent open-source Terraform provider for SAP
Integration Suite. It is not an official SAP product, and SAP has not
endorsed or certified it, unless SAP formally adopts or publishes it in the
future. It only uses officially documented, SAP-supported public APIs — see
`docs/sap-api-references.md`.

## License

[Apache License 2.0](LICENSE)
