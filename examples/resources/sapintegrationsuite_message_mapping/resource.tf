resource "sapintegrationsuite_message_mapping" "customer" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "customer-mapping"
  name       = "Customer Mapping"

  content      = "${path.module}/message-mappings/customer-mapping.zip"
  content_hash = filesha256("${path.module}/message-mappings/customer-mapping.zip")
}
