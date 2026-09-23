data "sapintegrationsuite_custom_tag_configuration" "current" {}

output "required_tags" {
  value = [
    for tag in data.sapintegrationsuite_custom_tag_configuration.current.tags :
    tag.name
    if tag.mandatory
  ]
}
