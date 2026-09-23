# Requires provider.api_management to be configured. api_proxy_names
# references already-existing API Proxies by name (created through the SAP
# Integration Suite UI — sapintegrationsuite_api_proxy is not yet
# implemented, see docs/guides/classic-api-management.md) and is
# RequiresReplace, since no confirmed way exists to change it after
# creation.
resource "sapintegrationsuite_api_product" "sample" {
  name         = "SampleProduct"
  version      = "1"
  title        = "SampleProduct"
  description  = "Sample product bundling the SampleAPI proxy"
  status_code  = "PUBLISHED"
  is_published = true

  api_proxy_names = ["SampleAPI"]

  additional_properties = [
    {
      name  = "team"
      value = "integration-platform"
    },
  ]
}
