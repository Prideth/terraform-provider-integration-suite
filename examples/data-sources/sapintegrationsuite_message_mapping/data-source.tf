data "sapintegrationsuite_message_mapping" "customer" {
  package_id = "UTILITIES"
  mapping_id = "customer-mapping"
}

output "customer_message_mapping_version" {
  value = data.sapintegrationsuite_message_mapping.customer.version
}
