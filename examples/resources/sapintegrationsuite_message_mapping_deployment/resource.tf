resource "sapintegrationsuite_message_mapping_deployment" "customer" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_message_mapping.customer.mapping_id
  mapping_version = sapintegrationsuite_message_mapping.customer.version

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
