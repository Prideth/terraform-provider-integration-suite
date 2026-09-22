resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id   = sapintegrationsuite_integration_package.utilities.id
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
