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
#
# Constructed ARNs are used instead of resource references to break circular
# Terraform dependencies: on first apply the target resources do not exist in
# state yet, so referencing their .arn attributes would force Terraform to
# create them before attaching the policy - but creating them requires the
# policy to already be attached.
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
    resources = ["arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:${var.project_name}-api"]
  }

  statement {
    sid     = "ManageApiGateway"
    actions = ["apigateway:*"]
    # Wildcard covers /v2/apis/*, /tags/*, integrations, routes, stages, etc.
    # The previous /apis/* pattern missed /v2/apis/* and /tags/* (used for
    # tag operations), causing AccessDenied on CreateApi with tags.
    resources = ["arn:aws:apigateway:${data.aws_region.current.name}::/*"]
  }

  statement {
    sid = "ManageDynamoDBTable"
    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:CreateTable",
      "dynamodb:UpdateTable",
      "dynamodb:DeleteTable",
      "dynamodb:TagResource",
    ]
    resources = ["arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/${var.dynamodb_table_name}"]
  }

  statement {
    sid = "ManageSSMParameters"
    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:PutParameter",
      "ssm:DeleteParameter",
      "ssm:DescribeParameters",
      "ssm:AddTagsToResource",
    ]
    resources = ["arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${var.oauth_ssm_prefix}/*"]
  }

  statement {
    sid = "ManageLambdaRole"
    actions = [
      "iam:GetRole",
      "iam:CreateRole",
      "iam:DeleteRole",
      "iam:PassRole",
      "iam:GetRolePolicy",
      "iam:PutRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:TagRole",
    ]
    resources = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.project_name}-lambda-exec"]
  }

  statement {
    sid = "ManageLogGroup"
    actions = [
      "logs:CreateLogGroup",
      "logs:DeleteLogGroup",
      "logs:DescribeLogGroups",
      "logs:PutRetentionPolicy",
      "logs:TagResource",
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.project_name}-api",
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.project_name}-api:*"
      ]
  }

  statement {
    sid = "ManageOIDCProvider"
    actions = [
      "iam:CreateOpenIDConnectProvider",
      "iam:GetOpenIDConnectProvider",
      "iam:DeleteOpenIDConnectProvider",
      "iam:TagOpenIDConnectProvider",
    ]
    resources = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"]
  }

  # Allows subsequent applies to update this role's own policy document.
  statement {
    sid = "ManageDeployRole"
    actions = [
      "iam:CreateRole",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:PutRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:TagRole",
    ]
    resources = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.project_name}-github-actions-deploy"]
  }
}

resource "aws_iam_role_policy" "github_actions_permissions" {
  name   = "${var.project_name}-github-actions-permissions"
  role   = aws_iam_role.github_actions_deploy.id
  policy = data.aws_iam_policy_document.github_actions_permissions.json
}
