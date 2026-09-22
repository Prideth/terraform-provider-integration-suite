resource "sapintegrationsuite_partner_binary_parameter" "order_schema" {
  partner_id   = "PartnerZ"
  parameter_id = "OrderSchema"

  content_type = "xsd"
  content      = "${path.module}/partner-directory/order-schema.xsd"
  content_hash = filesha256("${path.module}/partner-directory/order-schema.xsd")
}
