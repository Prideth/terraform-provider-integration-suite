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
