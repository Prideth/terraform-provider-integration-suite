# Entry values are marked sensitive in this data source's schema, even
# though only unencrypted maps are supported.
data "sapintegrationsuite_api_key_value_map" "oc_instance_token" {
  name     = "apim.oc.instance.token"
  scope    = "APIPROXY"
  scope_id = "SampleAPI"
}
