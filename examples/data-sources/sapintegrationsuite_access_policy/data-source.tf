data "sapintegrationsuite_access_policy" "utilities" {
  id = "1"
}

output "utilities_access_policy_role" {
  value = data.sapintegrationsuite_access_policy.utilities.role_name
}
