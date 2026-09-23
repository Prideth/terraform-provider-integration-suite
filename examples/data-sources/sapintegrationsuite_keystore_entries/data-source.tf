# Lists both tenant-administrator-owned and SAP-owned entries - SAP's API
# exposes no field to filter by ownership. Useful for brownfield discovery
# before importing individual sapintegrationsuite_certificate or
# sapintegrationsuite_key_pair resources.
data "sapintegrationsuite_keystore_entries" "all" {}

output "entries_expiring_soon" {
  value = [
    for e in data.sapintegrationsuite_keystore_entries.all.entries :
    e.alias if e.valid_not_after != null
  ]
}
