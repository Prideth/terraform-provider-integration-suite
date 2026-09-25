resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id   = sapintegrationsuite_integration_package.utilities.id
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version

  # Redeploy in place when these values change, for example the parameters of
  # a sapintegrationsuite_integration_flow_configuration.
  redeploy_triggers = {
    release = "2026-09"
  }

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
