# SecureString parameters for OAuth2 and JWT secrets. Values are managed
# out-of-band (AWS CLI/Console) after apply, since committing real secrets
# to Terraform state/config is not acceptable. `ignore_changes` on `value`
# prevents Terraform from ever reverting a manually-set secret back to the
# placeholder.
locals {
  oauth_parameter_names = [
    "google_client_id",
    "google_client_secret",
    "github_client_id",
    "github_client_secret",
    "jwt_signing_key",
  ]
}

resource "aws_ssm_parameter" "oauth_secrets" {
  for_each = toset(local.oauth_parameter_names)

  name  = "${var.oauth_ssm_prefix}/${each.value}"
  type  = "SecureString"
  value = "CHANGEME"

  lifecycle {
    ignore_changes = [value]
  }
}
