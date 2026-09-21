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

**Repository naming**: the Go module, provider binary name, registry
manifest, and every path in this repository already use the final naming
(`github.com/Prideth/terraform-provider-sap-integration-suite`,
`Prideth/sap-integration-suite`, `sapintegrationsuite`). The GitHub
repository itself is still hosted at
`Prideth/terraform-provider-integration-suite`; renaming it to
`Prideth/terraform-provider-sap-integration-suite` is the one remaining
step (GitHub repository rename tooling was not available in the
environment this codebase was developed in). GitHub automatically
redirects the old URL after a rename, so this is a safe, low-risk,
purely administrative step — Settings → repository name, in the GitHub UI
or `gh repo rename terraform-provider-sap-integration-suite`.

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

Every resource here is backed by a currently documented, SAP-supported
public API — see [`docs/sap-api-references.md`](docs/sap-api-references.md)
for the exact API behind each one, and
[`docs/api-capability-matrix.md`](docs/api-capability-matrix.md) /
[`docs/provisioning-capability-matrix.md`](docs/provisioning-capability-matrix.md)
for the full discovery behind what is and is not implemented yet.

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
- `sapintegrationsuite_integration_flow`'s `content`/`content_hash` cannot
  be populated by `terraform import`, since SAP does not return a local
  file path for an existing design-time artifact; apply a matching
  configuration after import to bring content under management.
- API Gateway / API Artifacts, classic API Management, Integration Cell,
  and Edge Integration Cell resources are not yet implemented — see
  `ROADMAP.md`.

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
