resource "sapintegrationsuite_value_mapping_deployment" "company_codes" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_value_mapping.company_codes.mapping_id
  mapping_version = sapintegrationsuite_value_mapping.company_codes.version

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
