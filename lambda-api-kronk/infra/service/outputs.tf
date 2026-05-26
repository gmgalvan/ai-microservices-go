locals {
  api_base_url = trimsuffix(aws_apigatewayv2_stage.default.invoke_url, "/")
}

output "api_invoke_url" {
  description = "Base invoke URL for the HTTP API."
  value       = local.api_base_url
}

output "hello_url" {
  description = "URL for the hello endpoint."
  value       = "${local.api_base_url}/api/hello"
}

output "bye_url" {
  description = "URL for the bye endpoint."
  value       = "${local.api_base_url}/api/bye"
}

output "hello_curl" {
  description = "curl command for the hello endpoint."
  value       = "curl '${local.api_base_url}/api/hello'"
}

output "bye_curl" {
  description = "curl command for the bye endpoint."
  value       = "curl '${local.api_base_url}/api/bye'"
}
