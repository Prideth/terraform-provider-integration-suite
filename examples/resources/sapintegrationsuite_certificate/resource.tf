# certificate is public X.509 content, never a secret - it is never marked
# Sensitive. Drift detection compares a canonical SHA-256 fingerprint of
# the certificate's DER bytes, not raw PEM text, so re-wrapped or
# CRLF-converted PEM content never causes a spurious diff.
resource "sapintegrationsuite_certificate" "backend_ca" {
  alias       = "backend-root-ca"
  certificate = file("${path.module}/backend-root-ca.pem")
}

output "backend_ca_fingerprint" {
  value = sapintegrationsuite_certificate.backend_ca.certificate_sha256
}
