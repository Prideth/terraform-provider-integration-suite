data "sapintegrationsuite_api_provider" "backend" {
  name = "ES5_1"
}

output "backend_host" {
  value = data.sapintegrationsuite_api_provider.backend.host
}
