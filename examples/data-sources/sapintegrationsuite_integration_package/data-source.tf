data "sapintegrationsuite_integration_package" "utilities" {
  id = "UTILITIES"
}

output "utilities_package_name" {
  value = data.sapintegrationsuite_integration_package.utilities.name
}
