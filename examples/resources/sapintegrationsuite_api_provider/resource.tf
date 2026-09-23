# Requires provider.api_management to be configured. Only the "Internet"
# connection type is supported — see docs/guides/classic-api-management.md.
# This resource has no Update: every attribute is RequiresReplace.
resource "sapintegrationsuite_api_provider" "backend" {
  name      = "ES5_1"
  title     = "ES5"
  host      = "sapes5.sapdevcenter.com"
  port      = 443
  use_ssl   = true
  trust_all = false

  auth_type   = "BASIC"
  user_name   = "backend-user"
  password_wo = var.backend_password
}

variable "backend_password" {
  type      = string
  sensitive = true
}
