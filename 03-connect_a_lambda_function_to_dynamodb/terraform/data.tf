# Fetches info about the AWS account/credentials Terraform is currently
# using - most commonly used to build ARNs or to sanity-check you're
# deploying into the account you think you are.

data "aws_caller_identity" "current" {}

# Fetches info about the region the provider is configured for (from
# var.aws_region).

data "aws_region" "current" {}
