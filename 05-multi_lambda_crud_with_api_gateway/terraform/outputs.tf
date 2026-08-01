output "invoke_url" {
  description = "Base URL for the API - append /users, /users/{id} or /echo"
  value       = aws_api_gateway_stage.test.invoke_url
}

output "table_name" {
  description = "DynamoDB table the Lambdas read TABLE_NAME as"
  value       = aws_dynamodb_table.users.name
}
