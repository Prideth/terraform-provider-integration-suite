# Provider Boundaries

This document draws the line between the three Terraform-manageable layers that make up a
full SAP Integration Suite landscape, so that module authors know which provider owns which
resource.

```
SAP/btp
    |
    v
BTP Control Plane
    - Global Account / Directory / Subaccount
    - Entitlements
    - Integration Suite subscription
    - Service instances / service bindings
    - Destinations
    - Role collections / role assignments


Prideth/sap-integration-suite
    |
    v
Integration Suite Control Plane + Runtime Configuration
    - Capability discovery / configuration (where a public API exists)
    - Cloud Integration design-time content and deployments
    - Access Policies
    - API Management (classic) and the API Gateway / API Artifact model
    - Integration Cell and Edge Integration Cell control-plane configuration


Kubernetes / Helm providers
    |
    v
Customer-managed Runtime Infrastructure
    - The Kubernetes cluster an Edge Integration Cell runs on
    - Cluster-level workloads, networking, storage
```

## How the layers connect

Terraform modules compose these providers through ordinary inputs and outputs — never
through direct Go-level coupling:

```hcl
module "btp" {
  source = "..." # uses the SAP/btp provider

  # -> produces: integration suite tenant URL, OAuth service key
}

provider "sapintegrationsuite" {
  host = module.btp.integration_suite_api_url

  oauth {
    token_url     = module.btp.oauth_token_url
    client_id     = module.btp.oauth_client_id
    client_secret = module.btp.oauth_client_secret
  }
}

resource "sapintegrationsuite_integration_package" "utilities" {
  # ...
}
```

## Starting point assumption

`Prideth/sap-integration-suite` assumes:

1. A BTP subaccount already exists.
2. An SAP Integration Suite subscription already exists on that subaccount.
3. A service key / OAuth client with the necessary Integration Suite roles already exists.

It never creates a subaccount, assigns entitlements, or creates the Integration Suite
subscription itself — those steps stay in `SAP/btp`.

## Capability activation boundary

"Provisioning" in this provider's context never means creating a BTP subscription. It means,
strictly inside an already-subscribed Integration Suite tenant:

```
Integration Suite already exists
        |
        v
  Capability discovered
        |
        v
  Capability activated   (only where a public API exists — see
        |                 docs/provisioning-capability-matrix.md)
        v
  Capability configured
        |
        v
  Runtime provisioned
        |
        v
  Readiness verified
```

Where no public activation API exists (this is currently true for every capability this
project has investigated — see the provisioning capability matrix), activation remains a
documented manual bootstrap step and is never implemented against an undocumented or
internal endpoint.
