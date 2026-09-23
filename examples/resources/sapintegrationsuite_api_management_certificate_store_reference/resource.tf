# Requires provider.api_management to be configured. The referenced
# keystore/truststore must already exist (created through the SAP
# Integration Suite UI — this provider does not manage certificate/keystore
# content for Classic API Management, only this pointer to it).
resource "sapintegrationsuite_api_management_certificate_store_reference" "backend_trust" {
  name                   = "backend-trust-reference"
  certificate_store_name = "backend-truststore"
}

# Rotating a certificate: point the reference at a newly created store
# without touching any virtual host configuration that uses the reference.
