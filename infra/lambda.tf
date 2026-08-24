resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.project_name}-api"
  retention_in_days = var.log_retention_days
}

resource "aws_lambda_function" "api" {
  function_name = "${var.project_name}-api"
  role          = aws_iam_role.lambda_exec.arn

  # Built and packaged by CI (see backend/Makefile `package` target) into
  # backend/dist/function.zip before `terraform apply` runs.
  filename         = var.lambda_package_path
  source_code_hash = filebase64sha256(var.lambda_package_path)

  handler       = "bootstrap"
  runtime       = "provided.al2023" # custom runtime for the Go binary, go1.x is deprecated.
  architectures = ["arm64"]         # cheaper and more efficient than x86_64.

  memory_size = var.lambda_memory_size
  timeout     = var.lambda_timeout

  environment {
    variables = {
      DYNAMODB_TABLE_NAME     = aws_dynamodb_table.users.name
      OAUTH_SSM_PREFIX        = var.oauth_ssm_prefix
      OAUTH_REDIRECT_BASE_URL = var.oauth_redirect_base_url
      ALLOWED_CORS_ORIGINS    = join(",", var.allowed_cors_origins)
    }
  }

  depends_on = [aws_cloudwatch_log_group.lambda]
}

resource "aws_lambda_permission" "apigw_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}
