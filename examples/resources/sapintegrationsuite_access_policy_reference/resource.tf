resource "sapintegrationsuite_access_policy_reference" "utilities_flows" {
  access_policy_id = sapintegrationsuite_access_policy.utilities.id

  artifact_type = "IntegrationFlow"
  attribute     = "Name"
  operator      = "EQUALS"
  value         = "metering"
}
