data "sapintegrationsuite_partner_string_parameters" "all" {
  partner_id = "PartnerZ"
}

output "partner_string_parameter_ids" {
  value = [for v in data.sapintegrationsuite_partner_string_parameters.all.values : v.parameter_id]
}
