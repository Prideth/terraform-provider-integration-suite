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

## Branch model

This repository uses a permanent two-branch model:

```
master
  ^
 dev
  ^
feature/<name>
```

- **`master`** is the stable/release branch, and remains the GitHub default
  branch. It only moves forward as an explicit release/stabilization step
  — never automatically after a feature merges.
- **`dev`** is the permanent integration branch. All feature work lands
  here first. `dev` is not the GitHub default branch.
- **`feature/<name>`** branches are always created from `dev`, and always
  merge back into `dev`, never directly into `master`. Name them after the
  feature (for example `feature/access-policy-completion`), never with a
  tooling/vendor/author prefix such as `claude/`, `ai/`, `bot/`, or
  `anthropic/`.

Workflow for a feature:

```shell
git switch dev
git pull --ff-only origin dev
git switch -c feature/<name>
# ... do the work, commit ...
git switch dev
git pull --ff-only origin dev
git merge --ff-only feature/<name>
git push origin dev
```

Prefer `--ff-only` while development is a single sequential stream. If a
fast-forward merge fails, stop and inspect why rather than forcing a merge
commit, rebasing, or force-pushing a published branch.

Promoting `dev` to `master` is a separate, deliberate release step,
performed only when explicitly requested — never automatically after a
feature merges into `dev`.

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
8. **Update the feature support catalog.** This is mandatory, not optional
   — no feature implementation is complete until this step is done:
   1. Add or update the feature's entry in `internal/features/catalog.go`
      (`support_status`, `support_reason`, `resource_types`,
      `data_source_types`, and every `Operations` flag).
   2. Only set an `Operations` flag to `true`, or move `support_status`
      toward `supported`, once the corresponding operation is actually
      implemented and covered by a test proving the claimed SAP API
      behavior — an `Update` method existing in Go is not by itself
      evidence that `operations.update` should be `true`; a resource
      whose `Update` exists to satisfy the Terraform Plugin Framework
      interface but never gets called for a real diff (`RequiresReplace`
      on everything mutable) is not `operations.update = true` either.
   3. Regenerate `docs/feature-support.md` with `go run ./cmd/gendocs > docs/feature-support.md`
      (also done by `make docs`).
   4. Run the catalog consistency tests
      (`go test ./internal/features/... ./internal/provider/... -run 'TestCatalog|TestFeatureCatalog'`)
      — they fail if a registered resource/data source has no catalog
      entry, or if a catalog entry claims a resource/data source type that
      does not exist.

## Commit and PR expectations

- Pull requests target `dev`, not `master` — see the branch model above.
- Run `go build ./...`, `go vet ./...`, `go test ./...`, `gofmt -l .`, and
  `golangci-lint run ./...` before opening a PR; CI enforces all of these
  on pull requests and on pushes to both `dev` and `master`.
- Keep commits logically scoped; a single PR can contain multiple commits.
- Fill in the PR template, including the SAP API reference for any new
  capability.

## Code of Conduct

This project follows the Code of Conduct in `CODE_OF_CONDUCT.md`.
