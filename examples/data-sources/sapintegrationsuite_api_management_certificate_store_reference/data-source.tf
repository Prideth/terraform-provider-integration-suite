data "sapintegrationsuite_api_management_certificate_store_reference" "backend_trust" {
  name = "backend-trust-reference"
}

output "backend_trust_store" {
  value = data.sapintegrationsuite_api_management_certificate_store_reference.backend_trust.certificate_store_name
}
