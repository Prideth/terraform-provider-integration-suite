# SAP documents exactly one Custom Tag Configuration per tenant, addressed
# by the fixed key "CustomTags" - this resource is a tenant-wide singleton.
# Every apply sends the complete desired tag list; there is no per-tag
# add/remove operation.
resource "sapintegrationsuite_custom_tag_configuration" "governance" {
  tags = [
    {
      name      = "Owner"
      mandatory = true
    },
    {
      name      = "BusinessUnit"
      mandatory = true

      permitted_values = [
        "Network",
        "Sales",
        "Water",
        "Transport",
      ]
    },
    {
      name      = "Criticality"
      mandatory = false

      permitted_values = [
        "Low",
        "Medium",
        "High",
      ]
    },
  ]
}
