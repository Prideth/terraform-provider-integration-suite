# Architecture

## Layering

```
                    Terraform Core
                          |
               Terraform Plugin Protocol (v6)
                          |
              sapintegrationsuite Provider
                          |
        +-----------------------------------+
        |           Provider Layer          |   internal/provider
        |  schemas, plan modifiers,         |
        |  validators, Configure/Import     |
        +-----------------------------------+
                          |
        +-----------------------------------+
        |            Domain Layer           |   internal/domain
        |  Terraform <-> SAP object mapping, |
        |  flatten/expand, ID composition   |
        +-----------------------------------+
                          |
        +-----------------------------------+
        |            API Clients            |   internal/client/*
        +-----------------------------------+
             |          |          |
             v          v          v
     Cloud Integration  API       Edge
        Client         Gateway  Integration
                        Client   Cell Client
                          |
              +----------------------+
              |   Shared Infrastructure   |  internal/client/{auth,http,odata}
              +----------------------+
```

Terraform resource and data source code never performs raw HTTP calls. Every resource talks
to a domain-layer mapper, which talks to a typed API client, which talks to the shared HTTP
client. This keeps schema concerns (Terraform types, plan modifiers, diagnostics) fully
separate from SAP wire-format concerns (OData batch/query syntax, pagination, error bodies).

## Package layout

```
internal/
├── client/
│   ├── auth/            OAuth2 client-credentials token source, cache, refresh
│   ├── http/            shared *http.Client wrapper: retry, backoff, jitter, User-Agent
│   ├── odata/
│   │   ├── v2/          OData V2 request/query building, "d"/"results" envelopes, paging
│   │   └── v4/          OData V4 request/query building, "value"/"@odata.nextLink" paging
│   ├── controlplane/    tenant/capability discovery (read-only where no activation API exists)
│   ├── cloudintegration/  Integration Content, Security Content, Partner Directory clients
│   ├── apimanagement/   classic API Management + API Gateway / API Artifact clients
│   ├── integrationcell/ Integration Cell status/config client
│   └── edgeintegrationcell/  Edge Integration Cell control-plane client
│
├── domain/              SAP <-> Terraform model mapping, shared by resources/data sources
│
└── provider/            terraform-plugin-framework provider, resources, data sources
```

## Authentication

Every SAP Integration Suite API area this provider talks to supports OAuth 2.0 client
credentials. The provider configuration accepts one base host plus one OAuth client
configuration by default, and allows overriding host/credentials per API area for tenants
that separate Cloud Integration, API Management, and Edge Integration Cell endpoints.

Clients for capability-specific APIs (Cloud Integration, API Management, Integration Cell,
Edge Integration Cell) are **initialized lazily**: `Configure` only builds a base HTTP
transport and a token source. A client is only exercised — and only fails — when a resource
that actually needs it is planned or applied. This avoids forcing every user to have every
capability activated before `terraform plan` can even run.

## OData V2 vs. V4

Cloud Integration's Integration Content, Security Content, and Partner Directory APIs are
OData V2. Some newer Integration Suite and Edge Integration Cell APIs use OData V4. The two
protocols differ enough (response envelope, pagination link shape, error body shape) that a
shared abstraction would either leak protocol details or hide real differences. The client
package therefore keeps `odata/v2` and `odata/v4` as separate, protocol-correct
implementations, sharing only generic helpers: URL encoding, `$filter` escaping, composite
key formatting.

## Retry and eventual consistency

The shared HTTP client (`internal/client/http`) retries `429`, `502`, `503`, and `504`
responses with exponential backoff and jitter, honoring a `Retry-After` header when present.
`401` triggers exactly one controlled token refresh and retry, never a blind retry loop.
`400`, `403`, and `404` are never retried.

Resources whose underlying SAP operation is asynchronous (for example, artifact deployment)
poll a status field using `context.Context`-aware backoff with a caller-supplied timeout,
never a fixed `time.Sleep`.

## Error handling

All API clients return a typed `apierror.Error` (status code, SAP error code, message,
details). The provider layer turns this into a Terraform diagnostic that names the resource,
the operation, and the SAP-reported code/message, with any credential material stripped
before it reaches a diagnostic or a log line.

## State, drift, and identity

Every resource is designed against a stable, SAP-assigned or user-assigned business key
(never a randomly generated Terraform UUID). `Read` always re-fetches the object from SAP; it
never simply echoes back prior state. A `404` on `Read` removes the resource from state; a
`404` on `Delete` is treated as success.
