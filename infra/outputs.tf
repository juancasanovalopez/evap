output "api_endpoint" {
  description = "Invoke URL of the HTTP API (no custom domain configured)."
  value       = aws_apigatewayv2_api.http_api.api_endpoint
}

output "lambda_function_name" {
  description = "Name of the deployed Lambda function."
  value       = aws_lambda_function.api.function_name
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table storing user profiles."
  value       = aws_dynamodb_table.users.name
}

output "sensor_readings_table_name" {
  description = "Name of the DynamoDB table storing IoT sensor readings."
  value       = aws_dynamodb_table.sensor_readings.name
}

output "iot_endpoint" {
  description = "AWS IoT Core data-plane endpoint devices must connect to."
  value       = data.aws_iot_endpoint.this.endpoint_address
}

output "lambda_role_arn" {
  description = "ARN of the Lambda execution role."
  value       = aws_iam_role.lambda_exec.arn
}

output "github_actions_role_arn" {
  description = "ARN to configure as the AWS_DEPLOY_ROLE_ARN GitHub secret/variable."
  value       = aws_iam_role.github_actions_deploy.arn
}
