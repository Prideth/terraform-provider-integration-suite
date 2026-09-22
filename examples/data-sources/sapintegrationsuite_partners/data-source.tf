data "sapintegrationsuite_partners" "all" {}

output "known_partner_ids" {
  value = data.sapintegrationsuite_partners.all.pids
}
