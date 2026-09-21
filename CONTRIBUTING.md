# Contributing

Thank you for considering a contribution to the Terraform Provider for SAP
Integration Suite.

## Ground rules

1. **Every resource and data source must trace back to an officially
   documented, SAP-supported public API.** See `docs/sap-api-references.md`
   for the pattern each resource follows, and never implement against an
   internal/UI-only endpoint discovered through browser network traffic.
2. **Respect the provider boundary.** BTP control-plane concerns (accounts,
   subaccounts, entitlements, subscriptions, generic destinations, role
   collections) belong in the official `SAP/btp` provider, not here. See
   `docs/provider-scope.md` and `docs/provider-boundaries.md`.
3. **Quality over quantity.** A resource is only added once it passes the
   suitability checklist in `docs/resource-design.md` (desired state,
   identity, read-back, create, update, delete, drift, import).

## Development setup

Requirements: Go (version pinned in `go.mod`), Terraform CLI (for
documentation generation and acceptance tests), `golangci-lint`.

```shell
go build ./...
go test ./...
make lint
make docs   # regenerate docs/ after any schema or description change
```

## Adding or changing a resource

1. Update the relevant matrix in `docs/api-capability-matrix.md` and/or
   `docs/provisioning-capability-matrix.md` with what you found and where.
2. Document the resource in `docs/resource-design.md` and
   `docs/sap-api-references.md`.
3. Implement the API client method(s) in `internal/client/...` — Terraform
   resource code must never perform raw HTTP calls directly.
4. Implement the resource/data source in `internal/provider/...` with
   Create/Read/Update/Delete/ImportState, using `RequiresReplace()` for any
   field SAP does not support updating in place.
5. Add unit tests (`httptest`-based) for the client method(s) and a basic
   schema/import sanity test for the resource.
6. Add an example under `examples/resources/<name>/` or
   `examples/data-sources/<name>/`, then run `make docs`.
7. If the change is acceptance-testable, add an acceptance test gated on
   `TF_ACC=1`, using `tf-acc-` prefixed names for any object created in a
   real tenant.

## Commit and PR expectations

- Run `go build ./...`, `go vet ./...`, `go test ./...`, `gofmt -l .`, and
  `golangci-lint run ./...` before opening a PR; CI enforces all of these.
- Keep commits logically scoped; a single PR can contain multiple commits.
- Fill in the PR template, including the SAP API reference for any new
  capability.

## Code of Conduct

This project follows the Code of Conduct in `CODE_OF_CONDUCT.md`.
