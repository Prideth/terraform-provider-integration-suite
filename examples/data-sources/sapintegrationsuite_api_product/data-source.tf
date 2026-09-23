data "sapintegrationsuite_api_product" "sample" {
  name = "SampleProduct"
}

output "sample_product_proxies" {
  value = data.sapintegrationsuite_api_product.sample.api_proxy_names
}
