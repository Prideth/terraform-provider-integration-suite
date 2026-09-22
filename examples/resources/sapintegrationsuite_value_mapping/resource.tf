resource "sapintegrationsuite_value_mapping" "company_codes" {
  package_id = sapintegrationsuite_integration_package.utilities.id
  mapping_id = "company-codes"
  name       = "Company Codes"

  content      = "${path.module}/value-mappings/company-codes.zip"
  content_hash = filesha256("${path.module}/value-mappings/company-codes.zip")
}
