# No SAP host or OAuth credentials are needed for this data source either.
data "sapintegrationsuite_provider_feature" "value_mapping" {
  key = "cloud_integration.value_mapping"
}

output "value_mapping_support" {
  value = {
    status      = data.sapintegrationsuite_provider_feature.value_mapping.support_status
    reason      = data.sapintegrationsuite_provider_feature.value_mapping.support_reason
    public_api  = data.sapintegrationsuite_provider_feature.value_mapping.public_api
    limitations = data.sapintegrationsuite_provider_feature.value_mapping.limitations
  }
}
