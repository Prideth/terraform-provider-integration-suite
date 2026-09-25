# Environment-specific values for the externalized parameters of the flow.
# Only these keys are managed; any other parameter keeps its value.
resource "sapintegrationsuite_integration_flow_configuration" "metering" {
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version

  parameters = {
    receiver_host       = "meter-gateway.prod.example.invalid"
    receiver_credential = "METER_GATEWAY_OAUTH"
    batch_size          = "250"
  }
}

# Parameters take effect at runtime only after a redeploy. Listing them in
# redeploy_triggers redeploys the flow in place whenever a value changes.
resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id   = sapintegrationsuite_integration_package.utilities.id
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version

  redeploy_triggers = sapintegrationsuite_integration_flow_configuration.metering.parameters
}
