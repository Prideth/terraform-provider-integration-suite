data "sapintegrationsuite_script_collection" "shared" {
  package_id           = "UTILITIES"
  script_collection_id = "shared-scripts"
}

output "shared_script_collection_version" {
  value = data.sapintegrationsuite_script_collection.shared.version
}
