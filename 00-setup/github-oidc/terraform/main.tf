# ------------------------------------------------------------------------------
# 1. The Terraform state bucket ARN
#
# The state bucket was created in:
# 00-setup/remote-state-aws-buket
#
# That setup has its own Terraform state file, so we cannot reference the bucket
# resource directly from here.
#
# Instead, we rebuild the bucket ARN using the current AWS account ID.


locals {
  state_bucket_arn = "arn:aws:s3:::aws-hands-on-tfstate-${data.aws_caller_identity.current.account_id}"
}

# ------------------------------------------------------------------------------
# 2. Register GitHub as an identity provider
#
# This tells AWS that it can trust identity tokens created by GitHub Actions.
#
# This resource does not give GitHub any AWS permissions.
# It only allows AWS to recognise GitHub as a valid identity provider.
#
# The IAM role below decides:
#
# - which GitHub repository can connect;
# - which AWS permissions it receives.


resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  # GitHub includes this value in tokens that are created for AWS.
  #
  # AWS checks it to confirm that the token was meant to be used with AWS.
  client_id_list = ["sts.amazonaws.com"]

  # thumbprint_list is not needed with AWS provider version 6.
  #
  # AWS now validates GitHub's certificate automatically.
  # Older Terraform examples may still include it.
}

# ------------------------------------------------------------------------------
# 3. Create the IAM role used by GitHub Actions - this role can only read
#
# A GitHub Actions workflow will temporarily assume this role.
#
# AWS then gives the workflow temporary AWS credentials for this role.
#
# The assume_role_policy below controls who is allowed to assume the role.


resource "aws_iam_role" "github_actions" {
  name = "aws-hands-on-github-actions"

  # This policy allows GitHub Actions to assume the role, but only when all the
  # conditions below are correct.
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",

        # Trust tokens created by the GitHub identity provider above.
        "Principal" : {
          "Federated" : aws_iam_openid_connect_provider.github.arn
        },

        # Exchange the GitHub token for temporary AWS credentials.
        "Action" : "sts:AssumeRoleWithWebIdentity",

        "Condition" : {
          # The GitHub token must have been created for AWS.
          "StringEquals" : {
            "token.actions.githubusercontent.com:aud" : "sts.amazonaws.com"
          },

          # The token must come from the GitHub repository defined in the
          # github_owner and github_repo variables.
          #
          # The :* allows workflows from any branch, tag or pull request in
          # that repository.
          #
          # It could later be restricted to the main branch with:
          #
          # repo:owner/name:ref:refs/heads/main
          "StringLike" : {
            "token.actions.githubusercontent.com:sub" : "repo:${var.github_owner}/${var.github_repo}:*"
          }
        }
      }
    ]
  })
}

# ------------------------------------------------------------------------------
# 4. Allow the role to read AWS resources
#
# terraform plan compares the Terraform configuration with the resources that
# already exist in AWS.
#
# To do this, it needs to read resources such as:
#
# - Lambda;
# - DynamoDB;
# - IAM;
# - API Gateway.
#
# ReadOnlyAccess allows the workflow to inspect these resources, but it does not
# allow it to create, update or delete them.


resource "aws_iam_role_policy_attachment" "read_only" {
  role       = aws_iam_role.github_actions.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

# ------------------------------------------------------------------------------
# 5. Allow the role to access the Terraform state
#
# Terraform needs to read the state file stored in the S3 bucket.
#
# Because use_lockfile = true is enabled, Terraform also creates a temporary
# .tflock file before using the state.
#
# When Terraform finishes, it deletes the .tflock file.
#
# The role therefore needs permission to:
#
# - list the bucket;
# - read the state and lock files;
# - create the lock file;
# - delete the lock file.


resource "aws_iam_role_policy" "state_access" {
  name = "terraform-state-access"
  role = aws_iam_role.github_actions.id

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        # Allow Terraform to list the files inside the state bucket.
        "Sid" : "ListStateBucket",
        "Effect" : "Allow",
        "Action" : ["s3:ListBucket"],
        "Resource" : local.state_bucket_arn
      },
      {
        # Allow Terraform to read the state file and create or delete the
        # temporary lock file.
        "Sid" : "ReadWriteStateAndLock",
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ],
        "Resource" : "${local.state_bucket_arn}/*"
      }
    ]
  })
} 