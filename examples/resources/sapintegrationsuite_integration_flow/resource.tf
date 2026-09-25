resource "sapintegrationsuite_integration_flow" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = "metering"
  name       = "Metering"

  content      = "${path.module}/iflows/metering.zip"
  content_hash = filesha256("${path.module}/iflows/metering.zip")

  # Optional: save each release under an explicit version. Bump it together
  # with the content; a new version is saved only when this value changes.
  save_as_version = "1.0.3"
}
