# This file has two jobs:
#   1. The `terraform` block: pins WHICH VERSIONS of Terraform + the AWS
#      provider plugin this config is known to work with. Without this,
#      Terraform will happily use whatever latest version is installed,
#      which can silently break your config months later when AWS releases
#      a new provider version with breaking changes.
#   2. The `provider "aws"` block: configures HOW Terraform talks to AWS
#      (which region, which credentials).


terraform {
  required_version = ">= 1.12.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region  = var.aws_region
  profile = "developer"

  default_tags {
    tags = {
      Project     = "aws-hands-on"
      Tutorial    = "03-connect_a_lambda_function_to_dynamodb"
      Environment = "learning"
    }
  }
}