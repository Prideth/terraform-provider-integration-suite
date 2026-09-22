variable "receiver_communication_password" {
  type      = string
  sensitive = true
}

resource "sapintegrationsuite_partner_user_credential_parameter" "receiver" {
  partner_id   = "Receiver_1"
  parameter_id = "USER"
  user         = "commuser1"

  # password_wo is write-only: Terraform never stores it in plan or state.
  # Bump password_wo_version whenever the password itself changes so
  # Terraform detects the rotation (this replaces the resource, since no
  # in-place password update is confirmed for this SAP entity).
  password_wo         = var.receiver_communication_password
  password_wo_version = "1"
}
