# AWS IoT Core broker: devices publish sensor readings over MQTT, a Topic
# Rule ingests them straight into DynamoDB (no Lambda in the hot path).

data "aws_iot_endpoint" "this" {
  endpoint_type = "iot:Data-ATS"
}

# Device policy: mTLS-authenticated clients may only connect as their own
# Thing name and publish under their owner's topic prefix, no subscribe/receive.
# `owner_user_id` must be set as a Thing attribute at provisioning time so a
# device can never publish under another user's prefix even with a valid cert.
data "aws_iam_policy_document" "iot_device" {
  statement {
    sid       = "Connect"
    actions   = ["iot:Connect"]
    resources = ["arn:aws:iot:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:client/$${iot:Connection.Thing.ThingName}"]
  }

  statement {
    sid       = "PublishOwnTopic"
    actions   = ["iot:Publish"]
    resources = ["arn:aws:iot:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:topic/evap/$${iot:Connection.Thing.Attributes[owner_user_id]}/$${iot:Connection.Thing.ThingName}/readings"]
  }
}

resource "aws_iot_policy" "device" {
  name   = "${var.project_name}-iot-device-policy"
  policy = data.aws_iam_policy_document.iot_device.json
}

# --- Topic rule: evap/{device_id}/readings -> DynamoDB --------------------

resource "aws_cloudwatch_log_group" "iot_rule_errors" {
  name              = "/aws/iot/${var.project_name}-sensor-ingest-errors"
  retention_in_days = var.log_retention_days

  lifecycle {
    ignore_changes = [tags_all]
  }
}

data "aws_iam_policy_document" "iot_rule_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "iot_rule" {
  name               = "${var.project_name}-iot-rule-role"
  assume_role_policy = data.aws_iam_policy_document.iot_rule_assume_role.json
}

data "aws_iam_policy_document" "iot_rule_permissions" {
  statement {
    sid       = "WriteSensorReadings"
    actions   = ["dynamodb:PutItem"]
    resources = [aws_dynamodb_table.sensor_readings.arn]
  }

  statement {
    sid       = "WriteErrorLogs"
    actions   = ["logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["${aws_cloudwatch_log_group.iot_rule_errors.arn}:*"]
  }
}

resource "aws_iam_role_policy" "iot_rule_permissions" {
  name   = "${var.project_name}-iot-rule-permissions"
  role   = aws_iam_role.iot_rule.id
  policy = data.aws_iam_policy_document.iot_rule_permissions.json
}

resource "aws_iot_topic_rule" "sensor_ingest" {
  name        = "${replace(var.project_name, "-", "_")}_sensor_ingest"
  description = "Ingest evap/{owner_user_id}/{device_id}/readings messages into DynamoDB."
  enabled     = true
  sql         = "SELECT *, topic(2) AS owner_user_id, topic(3) AS device_id, timestamp() AS ingested_at FROM 'evap/+/+/readings'"
  sql_version = "2016-03-23"

  dynamodbv2 {
    role_arn = aws_iam_role.iot_rule.arn
    put_item {
      table_name = aws_dynamodb_table.sensor_readings.name
    }
  }

  error_action {
    cloudwatch_logs {
      role_arn       = aws_iam_role.iot_rule.arn
      log_group_name = aws_cloudwatch_log_group.iot_rule_errors.name
    }
  }
}
