terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  # Local state by default to keep costs at true $0 in a solo/dev setup.
  # For team use, replace with an S3 backend + DynamoDB lock table (small
  # fixed cost) - see README for the commented-out example.
}
