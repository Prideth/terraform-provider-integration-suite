data "sapintegrationsuite_access_policy_reference" "utilities_flows" {
  access_policy_id = "1"
  reference_id     = "ref-1"
}

output "utilities_flows_reference_value" {
  value = data.sapintegrationsuite_access_policy_reference.utilities_flows.value
}
