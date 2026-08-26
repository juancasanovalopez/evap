# GitHub Actions OIDC provider + a deploy role scoped to this repository's
# master branch, so CI never needs long-lived AWS access keys.

data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]
}

data "aws_iam_policy_document" "github_actions_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
        "repo:${var.github_repository}:ref:refs/heads/master",
        # GitHub can append immutable owner/repo IDs to the subject
        # (repo:owner@ownerId/repo@repoId:ref:...) to prevent namespace
        # squatting after a rename/transfer; match that shape too.
        "repo:${split("/", var.github_repository)[0]}@*/${split("/", var.github_repository)[1]}@*:ref:refs/heads/master",
      ]
    }
  }
}

resource "aws_iam_role" "github_actions_deploy" {
  name               = "${var.project_name}-github-actions-deploy"
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json
}

# Scoped to exactly the resources this project's Terraform manages - never
# AdministratorAccess.
data "aws_iam_policy_document" "github_actions_permissions" {
  statement {
    sid = "ManageLambda"
    actions = [
      "lambda:GetFunction",
      "lambda:CreateFunction",
      "lambda:UpdateFunctionCode",
      "lambda:UpdateFunctionConfiguration",
      "lambda:DeleteFunction",
      "lambda:AddPermission",
      "lambda:RemovePermission",
      "lambda:GetPolicy",
      "lambda:TagResource",
      "lambda:ListTags",
    ]
    resources = [aws_lambda_function.api.arn]
  }

  statement {
    sid       = "ManageApiGateway"
    actions   = ["apigateway:*"]
    resources = ["arn:aws:apigateway:${data.aws_region.current.name}::/apis/*"]
  }

  statement {
    sid       = "ManageDynamoDBTable"
    actions   = ["dynamodb:DescribeTable", "dynamodb:CreateTable", "dynamodb:UpdateTable", "dynamodb:DeleteTable", "dynamodb:TagResource"]
    resources = [aws_dynamodb_table.users.arn]
  }

  statement {
    sid       = "ManageSSMParameters"
    actions   = ["ssm:GetParameter", "ssm:GetParameters", "ssm:PutParameter", "ssm:DescribeParameters", "ssm:AddTagsToResource"]
    resources = [for p in aws_ssm_parameter.oauth_secrets : p.arn]
  }

  statement {
    sid       = "ManageLambdaRoleAndLogs"
    actions   = ["iam:GetRole", "iam:PassRole", "iam:GetRolePolicy", "iam:PutRolePolicy"]
    resources = [aws_iam_role.lambda_exec.arn]
  }

  statement {
    sid       = "ManageLogGroup"
    actions   = ["logs:DescribeLogGroups", "logs:PutRetentionPolicy", "logs:TagResource"]
    resources = [aws_cloudwatch_log_group.lambda.arn]
  }
}

resource "aws_iam_role_policy" "github_actions_permissions" {
  name   = "${var.project_name}-github-actions-permissions"
  role   = aws_iam_role.github_actions_deploy.id
  policy = data.aws_iam_policy_document.github_actions_permissions.json
}
