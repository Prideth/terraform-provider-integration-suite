## Summary

## SAP API reference

Link to the official SAP documentation for the public API this change relies on (see `docs/sap-api-references.md`).

## Checklist

- [ ] `go build ./...`, `go vet ./...`, and `go test ./...` pass
- [ ] `gofmt -l .` reports nothing
- [ ] `golangci-lint run ./...` is clean
- [ ] `make docs` was run if any resource/data source schema or description changed, and the diff is committed
- [ ] New or changed resources document import, drift detection, and any immutable fields in `docs/resource-design.md`
- [ ] No secrets, tokens, or tenant-specific values were committed
