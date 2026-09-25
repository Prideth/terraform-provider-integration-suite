resource "sapintegrationsuite_integration_package" "adapters" {
  id   = "CUSTOM_ADAPTERS"
  name = "Custom Adapters"
}

resource "sapintegrationsuite_integration_adapter" "sftp_extension" {
  id         = "custom-sftp-extension"
  package_id = sapintegrationsuite_integration_package.adapters.id

  # SAP documents id as unique across the whole tenant, not just this
  # package, and rejects importing a duplicate id as an error - there is no
  # confirmed way to update an existing adapter's content or metadata in
  # place, so every attribute here forces replacement.
  name = "Custom SFTP Extension"

  content      = "${path.module}/adapters/custom-sftp-extension.esa"
  content_hash = filesha256("${path.module}/adapters/custom-sftp-extension.esa")
}
