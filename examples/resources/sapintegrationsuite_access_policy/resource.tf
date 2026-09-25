# The role name must match the Values attribute of a BTP custom role; only
# users holding that role see the artifacts this policy protects.
resource "sapintegrationsuite_access_policy" "utilities" {
  role_name   = "UTILITIES_ARCHITECT"
  description = "Integration flows owned by the utilities architecture team"
}
