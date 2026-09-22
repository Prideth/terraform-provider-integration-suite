resource "sapintegrationsuite_script_collection" "shared" {
  package_id           = sapintegrationsuite_integration_package.utilities.id
  script_collection_id = "shared-scripts"
  name                 = "Shared Scripts"

  content      = "${path.module}/script-collections/shared-scripts.zip"
  content_hash = filesha256("${path.module}/script-collections/shared-scripts.zip")
}
