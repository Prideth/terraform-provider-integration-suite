data "sapintegrationsuite_keystore_entry" "backend_ca" {
  alias = "backend-root-ca"
}

output "backend_ca_valid_until" {
  value = data.sapintegrationsuite_keystore_entry.backend_ca.valid_not_after
}
