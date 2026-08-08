# Configure Terraform to use the AWS provider
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  # Nothing circular here: this folder creates an OIDC provider and two IAM
  # roles, so its state can live in the bucket like any other. The bucket's own
  # folder is the one that has to stay local.
  backend "s3" {
    bucket       = "aws-hands-on-tfstate-725211237961"
    key          = "00-setup/github-oidc/terraform.tfstate"
    region       = "eu-west-1"
    profile      = "developer"
    use_lockfile = true
  }
}

# Set the AWS region where the S3 bucket will be created
# eu-west-1 = Ireland
provider "aws" {
  region  = "eu-west-1"
  profile = "developer"

  default_tags {
    tags = {
      Project     = "aws-hands-on"
      Environment = "learning"
    }
  }
}
