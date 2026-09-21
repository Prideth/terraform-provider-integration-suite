# Brownfield example: bring an existing Integration Suite landscape under
# Terraform management. Write the configuration below to match what already
# exists, then import each resource. `terraform plan` afterwards should show
# no changes once the configuration matches the live objects.
#
#   terraform import sapintegrationsuite_integration_package.utilities UTILITIES
#   terraform import sapintegrationsuite_integration_flow.metering UTILITIES/metering
#   terraform import sapintegrationsuite_integration_flow_deployment.metering UTILITIES/metering
#   terraform import sapintegrationsuite_access_policy.utilities <existing-access-policy-id>
#   terraform import sapintegrationsuite_access_policy_reference.utilities_flows <existing-access-policy-id>/<existing-reference-id>

terraform {
  required_providers {
    sapintegrationsuite = {
      source  = "Prideth/sap-integration-suite"
      version = "~> 0.1"
    }
  }
}

provider "sapintegrationsuite" {
  # host / oauth read from SAP_INTEGRATION_SUITE_* environment variables
}

resource "sapintegrationsuite_integration_package" "utilities" {
  id          = "UTILITIES"
  name        = "Utilities Integration"
  description = "Integration content for the utilities line of business"
}

# content/content_hash are left unset until the first `terraform apply`
# after import; SAP does not return a local file path for an existing
# design-time artifact, so Terraform keeps whatever was last applied until a
# matching local file is supplied (see docs/resources/integration_flow.md).
resource "sapintegrationsuite_integration_flow" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = "metering"
  name       = "Metering"
}

resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = sapintegrationsuite_integration_flow.metering.flow_id
}

resource "sapintegrationsuite_access_policy" "utilities" {
  role_name   = "UTILITIES_ARCHITECT"
  description = "Utilities architecture access"
}

resource "sapintegrationsuite_access_policy_reference" "utilities_flows" {
  access_policy_id = sapintegrationsuite_access_policy.utilities.id

  artifact_type = "IntegrationFlow"
  attribute     = "Name"
  operator      = "MATCHES"
  value         = "UTILITIES_.*"
}
