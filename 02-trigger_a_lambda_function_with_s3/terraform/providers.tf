terraform {
  required_version = ">= 1.12.2"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  backend "s3" {
    bucket = "aws-hands-on-tfstate-725211237961"
    key    = "02-trigger_a_lambda_function_with_s3/terraform.tfstate"
    region = "eu-west-1"
    profile      = "developer"
    use_lockfile = true
  }
}

provider "aws" {
  region  = "eu-west-1"
  profile = "developer"

  default_tags {
    tags = {
      Project     = "aws-hands-on"
      Tutorial    = var.lesson
      Environment = "learning"
    }
  }
}