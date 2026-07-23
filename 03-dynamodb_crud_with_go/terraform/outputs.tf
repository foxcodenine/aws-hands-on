

output "table_name" {
  description = "Name of the created DynamoDB table"
  value       = aws_dynamodb_table.users.name
}

output "table_arn" {
  description = "ARN of the table - needed when writing IAM policies that grant Lambda/app access to it"
  value       = aws_dynamodb_table.users.arn
}

output "endpoint" {
  description = "Regional DynamoDB service endpoint the SDK calls - same for every table in the region; the table name goes in the request body, not the URL"
  value       = "https://dynamodb.${data.aws_region.current.region}.amazonaws.com"
}

output "console_url" {
  description = "Clickable AWS Console deep-link to view this specific table"
  value       = "https://${data.aws_region.current.region}.console.aws.amazon.com/dynamodbv2/home?region=${data.aws_region.current.region}#table?name=${aws_dynamodb_table.users.name}"
}

output "gsi_name" {
  description = "Name of the status Global Secondary Index, matches IndexName used in the Go QueryByStatus function"
  value       = "status-index"
}

output "region" {
  description = "AWS region the table was deployed into"
  value       = data.aws_region.current.region
}

output "account_id" {
  description = "AWS account ID this table was deployed into (sanity check you're in the right account)"
  value       = data.aws_caller_identity.current.account_id
}
