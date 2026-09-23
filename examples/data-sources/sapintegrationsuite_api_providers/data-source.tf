data "sapintegrationsuite_api_providers" "all" {}

output "internet_providers" {
  value = [
    for p in data.sapintegrationsuite_api_providers.all.providers :
    p.name if p.dest_type == "INTERNET"
  ]
}
