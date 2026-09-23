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

  # Optional, and independent of the oauth block above - Classic API
  # Management (API Providers, API Proxies, API Products, Key Value Maps)
  # authenticates with its own API Portal application URL and OAuth 2.0
  # client (the apiportal-apiaccess service plan). Leave this entire block
  # out if you do not use any sapintegrationsuite_api_provider,
  # sapintegrationsuite_api_product, sapintegrationsuite_api_key_value_map,
  # or sapintegrationsuite_api_management_certificate_store_reference
  # resource or data source. All four values (or their
  # SAP_INTEGRATION_SUITE_API_MANAGEMENT_* environment variable
  # equivalents) must be supplied together, or all left unset.
  api_management {
    host          = var.api_management_host
    token_url     = var.api_management_token_url
    client_id     = var.api_management_client_id
    client_secret = var.api_management_client_secret
  }
}
