# Look the policy up by role name: the numeric ID differs in every tenant,
# the role name does not.
data "sapintegrationsuite_access_policy" "utilities" {
  role_name = "UTILITIES_ARCHITECT"
}

# Attach an extra rule to a policy that another team's configuration owns.
resource "sapintegrationsuite_access_policy_reference" "billing_flow" {
  access_policy_id = data.sapintegrationsuite_access_policy.utilities.id

  name          = "Billing flow"
  artifact_type = "INTEGRATION_FLOW"
  attribute     = "Name"
  operator      = "exactString"
  value         = "Billing"
}
