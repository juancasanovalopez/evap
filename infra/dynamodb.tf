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

# Sensor readings ingested from AWS IoT Core (see iot.tf).
# PK = device_id, SK = timestamp (ISO 8601, set by the publishing device).
# GSI lets a future API query strictly by the authenticated user's own
# owner_user_id, so one user's readings are never reachable by another.
resource "aws_dynamodb_table" "sensor_readings" {
  name         = var.sensor_readings_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "device_id"
  range_key    = "timestamp"

  attribute {
    name = "device_id"
    type = "S"
  }

  attribute {
    name = "timestamp"
    type = "S"
  }

  attribute {
    name = "owner_user_id"
    type = "S"
  }

  global_secondary_index {
    name            = "owner_user_id-timestamp-index"
    hash_key        = "owner_user_id"
    range_key       = "timestamp"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = false
  }
}
