resource "sapintegrationsuite_access_policy_reference" "metering_flow" {
  access_policy_id = sapintegrationsuite_access_policy.utilities.id

  name        = "Metering flow"
  description = "The metering integration flow of the utilities package"

  artifact_type = "INTEGRATION_FLOW"
  attribute     = "Name"
  operator      = "exactString"
  value         = "Metering"
}
