# Reads a reference exactly as SAP stores it. Useful for learning the wire
# constants of a reference that was created in the Integration Suite UI, for
# example one of artifact type "Integration Package" or operator "Matches".
data "sapintegrationsuite_access_policy_reference" "created_in_ui" {
  access_policy_id = "1901"
  reference_id     = "56"
}

output "ui_reference_constants" {
  value = {
    artifact_type = data.sapintegrationsuite_access_policy_reference.created_in_ui.artifact_type
    attribute     = data.sapintegrationsuite_access_policy_reference.created_in_ui.attribute
    operator      = data.sapintegrationsuite_access_policy_reference.created_in_ui.operator
  }
}
