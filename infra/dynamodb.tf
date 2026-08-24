# DynamoDB single-table design for user profiles.
# PK = "USER#<provider>#<providerID>", SK = "PROFILE".
resource "aws_dynamodb_table" "users" {
  name         = var.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST" # On-Demand: $0 when idle, scales automatically.
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  # Disabled to avoid the (small) additional PITR cost; enable for production
  # workloads that need point-in-time recovery.
  point_in_time_recovery {
    enabled = false
  }
}
