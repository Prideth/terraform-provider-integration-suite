terraform {
  required_providers {
    sapintegrationsuite = {
      source  = "Prideth/sap-integration-suite"
      version = "~> 0.1"
    }
  }
}

# Credentials can also be supplied via the SAP_INTEGRATION_SUITE_HOST,
# SAP_INTEGRATION_SUITE_TOKEN_URL, SAP_INTEGRATION_SUITE_CLIENT_ID, and
# SAP_INTEGRATION_SUITE_CLIENT_SECRET environment variables instead of the
# attributes below.
provider "sapintegrationsuite" {
  host = var.integration_suite_host

  oauth {
    token_url     = var.integration_suite_token_url
    client_id     = var.integration_suite_client_id
    client_secret = var.integration_suite_client_secret
  }
}
