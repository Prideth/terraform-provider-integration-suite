variable "backend_password" {
  type      = string
  sensitive = true
}

resource "sapintegrationsuite_user_credential" "backend" {
  id   = "BACKEND_BASIC"
  user = "integration-user"

  # password_wo is write-only: Terraform never stores it in plan or state.
  # Bump password_wo_version whenever the password itself changes so
  # Terraform redeploys this credential in place with the new password.
  password_wo         = var.backend_password
  password_wo_version = "1"
}

# SuccessFactors credentials additionally require company_id and kind.
resource "sapintegrationsuite_user_credential" "success_factors" {
  id         = "SFSF_BASIC"
  kind       = "SuccessFactors"
  user       = "sfsf-integration-user"
  company_id = "SFPART000123"

  password_wo         = var.backend_password
  password_wo_version = "1"
}
