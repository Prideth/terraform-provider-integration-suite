# Greenfield example: SAP BTP provisioning (subaccount, entitlements, the
# Integration Suite subscription) is intentionally out of scope for
# sapintegrationsuite and is handled by the official SAP BTP provider in a
# separate module. This example starts from the outputs of that module.
#
# terraform {
#   required_providers {
#     btp = {
#       source = "SAP/btp"
#     }
#   }
# }
#
# module "btp" {
#   source = "./modules/btp-integration-suite"
#   # ... subaccount, entitlements, Integration Suite subscription, service key ...
# }

terraform {
  required_providers {
    sapintegrationsuite = {
      source  = "Prideth/sap-integration-suite"
      version = "~> 0.1"
    }
  }
}

variable "integration_suite_host" {
  type        = string
  description = "Base URL of the Integration Suite tenant, e.g. from module.btp.integration_suite_api_url."
}

variable "integration_suite_token_url" {
  type        = string
  description = "OAuth token URL, e.g. from module.btp.oauth_token_url."
}

variable "integration_suite_client_id" {
  type        = string
  description = "OAuth client ID, e.g. from module.btp.oauth_client_id."
}

variable "integration_suite_client_secret" {
  type        = string
  sensitive   = true
  description = "OAuth client secret, e.g. from module.btp.oauth_client_secret."
}

provider "sapintegrationsuite" {
  host = var.integration_suite_host

  oauth {
    token_url     = var.integration_suite_token_url
    client_id     = var.integration_suite_client_id
    client_secret = var.integration_suite_client_secret
  }
}

resource "sapintegrationsuite_access_policy" "utilities" {
  role_name   = "UTILITIES_ARCHITECT"
  description = "Utilities architecture access"
}

resource "sapintegrationsuite_integration_package" "utilities" {
  id          = "UTILITIES"
  name        = "Utilities Integration"
  description = "Integration content for the utilities line of business"
}

resource "sapintegrationsuite_access_policy_reference" "utilities_flows" {
  access_policy_id = sapintegrationsuite_access_policy.utilities.id

  artifact_type = "IntegrationFlow"
  attribute     = "Name"
  operator      = "MATCHES"
  value         = "UTILITIES_.*"
}

resource "sapintegrationsuite_integration_flow" "metering" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  flow_id    = "metering"
  name       = "Metering"

  content      = "${path.module}/iflows/metering.zip"
  content_hash = filesha256("${path.module}/iflows/metering.zip")
}

resource "sapintegrationsuite_integration_flow_deployment" "metering" {
  package_id   = sapintegrationsuite_integration_package.utilities.id
  flow_id      = sapintegrationsuite_integration_flow.metering.flow_id
  flow_version = sapintegrationsuite_integration_flow.metering.version
}

resource "sapintegrationsuite_value_mapping" "company_codes" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "company-codes"
  name       = "Company Codes"

  content      = "${path.module}/value-mappings/company-codes.zip"
  content_hash = filesha256("${path.module}/value-mappings/company-codes.zip")
}

resource "sapintegrationsuite_value_mapping_deployment" "company_codes" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_value_mapping.company_codes.mapping_id
  mapping_version = sapintegrationsuite_value_mapping.company_codes.version
}

# A reusable message mapping artifact: not wired up as a dependency of the
# metering flow above, since this sample flow's own content does not
# actually reference it. In a real landscape, an integration flow that does
# reference a message mapping artifact from a message mapping step still
# does not need Terraform to express that reference explicitly — the flow's
# own content owns it, and SAP does not automatically deploy a referenced
# message mapping when the referencing flow is deployed, so its deployment
# stays a separate resource either way.
resource "sapintegrationsuite_message_mapping" "customer" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "customer-mapping"
  name       = "Customer Mapping"

  content      = "${path.module}/message-mappings/customer-mapping.zip"
  content_hash = filesha256("${path.module}/message-mappings/customer-mapping.zip")
}

resource "sapintegrationsuite_message_mapping_deployment" "customer" {
  package_id      = sapintegrationsuite_integration_package.utilities.id
  mapping_id      = sapintegrationsuite_message_mapping.customer.mapping_id
  mapping_version = sapintegrationsuite_message_mapping.customer.version
}

# A reusable script collection: also not wired up as a dependency of the
# metering flow, for the same reason as the message mapping above.
resource "sapintegrationsuite_script_collection" "shared" {
  package_id           = sapintegrationsuite_integration_package.utilities.id
  script_collection_id = "shared-scripts"
  name                 = "Shared Scripts"

  content      = "${path.module}/script-collections/shared-scripts.zip"
  content_hash = filesha256("${path.module}/script-collections/shared-scripts.zip")
}

resource "sapintegrationsuite_script_collection_deployment" "shared" {
  package_id                = sapintegrationsuite_integration_package.utilities.id
  script_collection_id      = sapintegrationsuite_script_collection.shared.script_collection_id
  script_collection_version = sapintegrationsuite_script_collection.shared.version
}
