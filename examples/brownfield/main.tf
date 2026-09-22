# Brownfield example: bring an existing Integration Suite landscape under
# Terraform management. Write the configuration below to match what already
# exists, then import each resource. `terraform plan` afterwards should show
# no changes once the configuration matches the live objects.
#
#   terraform import sapintegrationsuite_integration_package.utilities UTILITIES
#   terraform import sapintegrationsuite_integration_flow.metering UTILITIES/metering
#   terraform import sapintegrationsuite_integration_flow_deployment.metering UTILITIES/metering
#   terraform import sapintegrationsuite_value_mapping.company_codes UTILITIES/company-codes
#   terraform import sapintegrationsuite_value_mapping_deployment.company_codes UTILITIES/company-codes
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
  package_id   = sapintegrationsuite_integration_package.utilities.id
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version
}

# content/content_hash are left unset until the first `terraform apply`
# after import, for the same reason as the integration flow above. Unlike
# the integration flow above, though, sapintegrationsuite_value_mapping has
# no confirmed in-place update path (see docs/sap-api-references.md): once
# you do supply content/content_hash here to bring the mapping's content
# under management, that first apply replaces the artifact (creates a new
# one, deletes the old one) rather than updating it in place. Every
# subsequent apply with an unchanged hash is a no-op, same as any other
# resource here.
resource "sapintegrationsuite_value_mapping" "company_codes" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "company-codes"
  name       = "Company Codes"
}

resource "sapintegrationsuite_value_mapping_deployment" "company_codes" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_value_mapping.company_codes.mapping_id
  mapping_version = sapintegrationsuite_value_mapping.company_codes.version
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
