# Provider Scope

This document defines what `Prideth/sap-integration-suite` manages and what it deliberately
does not manage. It exists so that users, reviewers, and contributors can tell at a glance
whether a feature request belongs in this provider or in a different one.

## In Scope

The provider administers the parts of an **already provisioned** SAP Integration Suite tenant
that are reachable through officially documented, SAP-supported public APIs:

- Integration Suite capability discovery and (where a public API exists) capability
  configuration
- Cloud Integration design-time content: integration packages, integration flows, value
  mappings, script collections, message mappings
- Cloud Integration runtime: deployments of the artifacts above, deployment status
- Access Policies and their artifact references
- Partner Directory entries, where a public API exists
- Classic API Management objects (API Providers, API Proxies, API Products, key value maps),
  where a public API exists
- The newer API-artifact-centric API Gateway model (API Artifacts, API Resources, API
  Policies, runtime profiles, deployments), where a public API exists
- Integration Cell: capability status/read, and any configuration that has a public API
- Edge Integration Cell: SAP-side registration, runtime association, and control-plane
  configuration that has a public API (never the Kubernetes workloads themselves)
- Integration-Suite-specific runtime configuration that is not already owned by the SAP BTP
  control plane

Every resource in this list is only implemented once a concrete, currently documented public
API has been identified for it (see [`sap-api-references.md`](sap-api-references.md) and the
capability/API matrices in this directory). A feature being visible in the Integration Suite
UI is not sufficient justification for a resource.

## Out of Scope

The following belong to the SAP BTP control plane, to Kubernetes/Helm, or to other existing
Terraform providers, and are intentionally **not** implemented here:

- BTP Global Accounts, Directories, Subaccounts
- BTP Entitlements and generic BTP Subscriptions (including subscribing to Integration Suite
  itself)
- Generic BTP Service Instances and Service Bindings
- BTP Role Collections and Role Assignments
- Generic BTP Destinations
- Cloud Foundry Organizations and Spaces
- Kyma, Kubernetes objects, Helm releases
- General BTP account/identity management

For all BTP control-plane concerns, use the official
[`SAP/btp`](https://registry.terraform.io/providers/SAP/btp/latest) Terraform provider. For
the customer-managed Kubernetes cluster that an Edge Integration Cell runs on, use the
Kubernetes and Helm providers directly.

## One-sentence summary

`Prideth/sap-integration-suite` is a Terraform provider for configuring, provisioning, and
administering **content and capabilities inside an already-provisioned SAP Integration
Suite tenant** — it is not, and will not become, a second SAP BTP provider.
