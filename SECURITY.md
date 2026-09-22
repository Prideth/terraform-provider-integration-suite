# Security Policy

## Reporting a Vulnerability

Please do not open a public GitHub issue for security vulnerabilities.
Instead, use GitHub's private vulnerability reporting for this repository
(the "Security" tab → "Report a vulnerability"). If that is not available,
open a draft security advisory or contact the repository owner directly
through their GitHub profile.

Please include:

- A description of the vulnerability and its impact
- Steps to reproduce, or a minimal Terraform configuration that triggers it
- The provider version and Terraform version

## Scope

This provider talks to SAP Integration Suite tenants using credentials the
operator supplies. In scope for a security report:

- Credential, token, or secret material appearing in logs, diagnostics, or
  state in plaintext where it should be protected
- TLS verification being bypassed
- Server-side request forgery, path traversal, zip-slip, XXE, or unbounded
  decompression in how the provider handles file content or API responses
- Dependency vulnerabilities affecting this provider's runtime behavior

Out of scope: vulnerabilities in SAP Integration Suite itself (report those
to SAP directly) and vulnerabilities in Terraform Core.

## Supported Versions

Security fixes are made against the latest minor release on the `0.x` line
until `1.0.0` is released, after which the latest major version will be
supported.

## Known dependency findings without an available upstream fix

`govulncheck` runs on every push and pull request (see
`.github/workflows/lint.yml`) and is not suppressed. When it reports a
finding with no upstream fix release, that finding is recorded here rather
than hidden, until a fixed version can be adopted through the normal
dependency-upgrade process.

- **GO-2026-6443** (`google.golang.org/grpc`, CVE-2026-84445): a server
  panic when an xDS-routed request arrives with neither an `:authority`
  nor a `Host` header. The only release with a fix is a `v1.85.0-dev`
  pseudo-version; no stable `v1.85.0` (or later) had shipped as of this
  writing. `go.mod` pins `google.golang.org/grpc` to `v1.83.2` — a real,
  stable release that predates the vulnerability's reintroduction in
  `v1.84.0-dev` and onward — rather than adopting a prerelease build. This
  also means the provider is not running the *latest* grpc release; that
  is a deliberate, documented trade-off, not an oversight. Separately,
  this specific vulnerability requires xDS routing, which this provider's
  gRPC usage (the Terraform plugin protocol server `providerserver.Serve`
  spins up, talked to only by the local `terraform` CLI process that
  launched it) never configures, so it is understood to be unreachable
  here regardless. Revisit this pin once grpc-go ships a stable release
  containing the fix from
  [grpc/grpc-go#9365](https://github.com/grpc/grpc-go/pull/9365).
