data "sapintegrationsuite_value_mapping" "company_codes" {
  package_id = "UTILITIES"
  mapping_id = "company-codes"
}

output "company_codes_value_mapping_version" {
  value = data.sapintegrationsuite_value_mapping.company_codes.version
}
