variable "aws_region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Short project name used to prefix resource names."
  type        = string
  default     = "evap"
}

variable "environment" {
  description = "Deployment environment name (e.g. dev, prod)."
  type        = string
  default     = "prod"
}

variable "lambda_memory_size" {
  description = "Memory (MB) allocated to the Lambda function."
  type        = number
  default     = 256
}

variable "lambda_timeout" {
  description = "Timeout (seconds) for the Lambda function."
  type        = number
  default     = 10
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table storing user profiles."
  type        = string
  default     = "evap_users"
}

variable "allowed_cors_origins" {
  description = "Origins allowed to call the API with credentials."
  type        = list(string)
  default     = []
}

variable "oauth_ssm_prefix" {
  description = "SSM Parameter Store path prefix holding OAuth2/JWT secrets."
  type        = string
  default     = "/evap/prod"
}

variable "oauth_redirect_base_url" {
  description = "Base URL used to build OAuth2 redirect URIs (set to the API Gateway invoke URL once known)."
  type        = string
}

variable "log_retention_days" {
  description = "CloudWatch Logs retention in days, kept short to minimize cost."
  type        = number
  default     = 14
}

variable "github_repository" {
  description = "GitHub repository allowed to assume the deploy role, in the form org/repo."
  type        = string
}

variable "lambda_package_path" {
  description = "Path to the zipped Lambda deployment package built by `make package`."
  type        = string
  default     = "../backend/dist/function.zip"
}
