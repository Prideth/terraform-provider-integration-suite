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
#   terraform import sapintegrationsuite_message_mapping.customer UTILITIES/customer-mapping
#   terraform import sapintegrationsuite_message_mapping_deployment.customer UTILITIES/customer-mapping
#   terraform import sapintegrationsuite_script_collection.shared UTILITIES/shared-scripts
#   terraform import sapintegrationsuite_script_collection_deployment.shared UTILITIES/shared-scripts
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
  short_text  = "Utilities integration content"
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

# content/content_hash are left unset until the first `terraform apply`
# after import, for the same reason as the integration flow above.
# sapintegrationsuite_message_mapping does have a confirmed in-place update
# path (unlike sapintegrationsuite_value_mapping): once you supply
# content/content_hash here, that first apply updates the artifact's
# content rather than replacing the resource, and every subsequent apply
# with an unchanged hash is a no-op.
resource "sapintegrationsuite_message_mapping" "customer" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "customer-mapping"
  name       = "Customer Mapping"
}

resource "sapintegrationsuite_message_mapping_deployment" "customer" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_message_mapping.customer.mapping_id
  mapping_version = sapintegrationsuite_message_mapping.customer.version
}

# content/content_hash are left unset until the first `terraform apply`
# after import, for the same reason as the other file-based resources
# above. sapintegrationsuite_script_collection has a confirmed in-place
# update path, the same as sapintegrationsuite_message_mapping.
resource "sapintegrationsuite_script_collection" "shared" {
  package_id           = sapintegrationsuite_integration_package.utilities.id
  script_collection_id = "shared-scripts"
  name                 = "Shared Scripts"
}

resource "sapintegrationsuite_script_collection_deployment" "shared" {
  package_id                = sapintegrationsuite_integration_package.utilities.id
  script_collection_id      = sapintegrationsuite_script_collection.shared.script_collection_id
  script_collection_version = sapintegrationsuite_script_collection.shared.version
}

resource "sapintegrationsuite_access_policy" "utilities" {
  role_name   = "UTILITIES_ARCHITECT"
  description = "Utilities architecture access"
}

resource "sapintegrationsuite_access_policy_reference" "utilities_flows" {
  access_policy_id = sapintegrationsuite_access_policy.utilities.id

  name          = "Metering flow"
  artifact_type = "INTEGRATION_FLOW"
  attribute     = "Name"
  operator      = "exactString"
  value         = "Metering"
}
