# Configure Terraform to use the AWS provider
#
# This folder creates the state bucket, and its own state lives inside that same
# bucket. Self-referential, and fine day to day - the bucket is just a bucket,
# it does not care that the file describing it is sitting in it.
#
# Two moments where the ordering does matter:
#
# 1. Starting from an empty AWS account. The bucket does not exist yet, so
#    terraform init cannot reach the backend and fails before it can create
#    anything. Comment out the backend block, terraform apply to make the
#    bucket, then uncomment it and run:
#
#      terraform init -migrate-state
#
# 2. Tearing it all down. The state file is in the bucket, so DeleteBucket
#    fails with BucketNotEmpty. Migrate the state back to local first, then
#    remove prevent_destroy and the deny policy in main.tf, then destroy.
#
# Neither is a reason to keep state on one laptop. Losing it would not be a
# disaster either - state is a pointer, not the infrastructure, and these four
# resources could be brought back with terraform import.
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  backend "s3" {
    bucket       = "aws-hands-on-tfstate-725211237961"
    key          = "00-setup/remote-state-aws-buket/terraform.tfstate"
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
