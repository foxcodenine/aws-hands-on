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
  region  = "eu-west-1"
  profile = "developer"

  default_tags {
    tags = {
      Project     = "aws-hands-on"
      Tutorial    = "02-trigger_a_lambda_function_with_s3"
      Environment = "learning"
    }
  }
}