# Requires provider.api_management to be configured. This resource has no
# Update and does not support encrypted maps — see
# docs/guides/classic-api-management.md. Changing entries replaces the
# whole Key Value Map.
resource "sapintegrationsuite_api_key_value_map" "oc_instance_token" {
  name     = "apim.oc.instance.token"
  scope    = "APIPROXY"
  scope_id = "SampleAPI"

  entries = [
    {
      key   = "default"
      value = var.open_connectors_instance_token
    },
  ]
}

variable "open_connectors_instance_token" {
  type      = string
  sensitive = true
}
