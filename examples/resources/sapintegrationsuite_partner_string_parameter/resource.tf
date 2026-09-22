resource "sapintegrationsuite_partner_string_parameter" "receiver_address" {
  partner_id   = "PartnerZ"
  parameter_id = "ReceiverAddress"
  value        = "https://receiver.example.com"
}
