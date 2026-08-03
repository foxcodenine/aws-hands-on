# ------------------------------------------------------------------------------
# The apply role
#
# main.tf creates a role that can only read. That is enough for terraform plan,
# and it is safe to let any branch use it.
#
# terraform apply needs to create and delete real resources, so it needs a
# second role - and that role must be harder to reach.
#
# Two roles instead of one, because the trust and the permissions have to
# differ:
#
#  ┌─────────────────────────────────────────┬────────────────────────┬──────────────┐
#  |                  role                   │      permissions       │ trusted from │
#  ├─────────────────────────────────────────┼────────────────────────┼──────────────┤
#  │ aws-hands-on-github-actions (exists)    │ ReadOnlyAccess + state │ any branch   │
#  ├─────────────────────────────────────────┼────────────────────────┼──────────────┤
#  │ aws-hands-on-github-actions-apply (new) │ read + write           │ main only    │
#  └─────────────────────────────────────────┴────────────────────────┴──────────────┘
#
# Keeping them apart means the permissive trust and the dangerous permissions
# never sit on the same role.

# ------------------------------------------------------------------------------
# 1. Create the role the apply job will use
#
# Same idea as the role in main.tf: GitHub Actions assumes it and receives
# temporary AWS credentials.
#
# The difference is in the condition below, which decides who may assume it.


resource "aws_iam_role" "github_actions_apply" {
  name = "aws-hands-on-github-actions-apply"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",

        # Trust tokens created by the GitHub identity provider in main.tf.
        # Both roles share the one provider - there is only ever one per
        # account for a given URL.
        "Principal" : {
          "Federated" : aws_iam_openid_connect_provider.github.arn
        },

        "Action" : "sts:AssumeRoleWithWebIdentity",

        "Condition" : {
          "StringEquals" : {
            # The GitHub token must have been created for AWS.
            "token.actions.githubusercontent.com:aud" : "sts.amazonaws.com",

            # This is the whole point of the second role.
            #
            # The role in main.tf ends in :* so any branch, tag or pull request
            # can use it. This one names one exact branch, so a workflow running
            # anywhere else is refused by AWS before it gets credentials.
            #
            # StringEquals, not StringLike: there is no wildcard left to match.
            "token.actions.githubusercontent.com:sub" : "repo:${var.github_owner}/${var.github_repo}:ref:refs/heads/main"
          }
        }
      }
    ]
  })
}

# ------------------------------------------------------------------------------
# 2. Allow the role to read AWS resources

# Attach an existing AWS-managed policy. Two separate things already exist (role, policy); 
# this resource just links them.

# apply reads before it writes: it refreshes the state first, which means
# describing every resource it manages to see what changed.
#
# So this role needs the same read access as the plan role, on top of the write
# permissions further down.


resource "aws_iam_role_policy_attachment" "apply_read_only" {
  role       = aws_iam_role.github_actions_apply.name
  policy_arn = "arn:aws:iam::aws:policy/ReadOnlyAccess"
}

# ------------------------------------------------------------------------------
# 3. Allow the role to access the Terraform state
#
# The same policy body as the plan role in main.tf, repeated for this role.
#
# Note this is an inline policy, not an attachment like step 2 above. An inline
# policy is written into the role rather than existing on its own, so there is
# no policy_arn here and nothing separate to attach. It is created and deleted
# with the role.
#
# Inline suits it: only this role will ever use it. ReadOnlyAccess in step 2 is
# the opposite case - it exists on its own and any role can attach it.
#
# Terraform reads the state file from S3, and because use_lockfile = true it
# also creates a temporary .tflock file before writing and deletes it after.
#
# So the role needs to list the bucket, read the state, and create and delete
# the lock file.


resource "aws_iam_role_policy" "apply_state_access" {
  name = "terraform-state-access"
  role = aws_iam_role.github_actions_apply.id

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
        "Action" : ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
        "Resource" : "${local.state_bucket_arn}/*"
      }
    ]
  })
}

# ------------------------------------------------------------------------------
# 4. Allow the role to create and delete the tutorial resources
#
# This is what the plan role deliberately does not have.
#
# The policy is in two halves, because the two halves need different limits.


resource "aws_iam_role_policy" "apply_write" {
  name = "terraform-apply"
  role = aws_iam_role.github_actions_apply.id

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        # The four services the tutorials actually build with.
        #
        # Broad on purpose: a learning repo adds new resource types often, and a
        # narrower list would mean a failed apply every time. The limit that
        # matters is the one below.
        "Sid" : "ManageTutorialResources",
        "Effect" : "Allow",
        "Action" : [
          "lambda:*",
          "dynamodb:*",
          "apigateway:*",
          "logs:*"
        ],
        "Resource" : "*"
      },
      {
        # IAM is the dangerous one, so it is limited by name.
        #
        # Everything Terraform creates here is called aws-hands-on-something, so
        # the role can manage its own roles and policies and nothing else in the
        # account.
        #
        # Without that limit, a role allowed iam:* could simply create itself an
        # administrator role and step around every other restriction here.
        #
        # iam:PassRole is the one that is easy to miss: creating a Lambda means
        # handing it an execution role, and AWS treats handing over a role as a
        # separate permission from creating one.
        "Sid" : "ManageTutorialIAM",
        "Effect" : "Allow",
        "Action" : [
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:GetRole",
          "iam:PassRole",
          "iam:TagRole",
          "iam:UntagRole",
          "iam:AttachRolePolicy",
          "iam:DetachRolePolicy",
          "iam:ListRolePolicies",
          "iam:ListAttachedRolePolicies",
          "iam:CreatePolicy",
          "iam:DeletePolicy",
          "iam:GetPolicy",
          "iam:GetPolicyVersion",
          "iam:CreatePolicyVersion",
          "iam:DeletePolicyVersion",
          "iam:ListPolicyVersions",
          "iam:TagPolicy",
          "iam:UntagPolicy"
        ],
        "Resource" : [
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-hands-on-*",
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/aws-hands-on-*"
        ]
      }
    ]
  })
}
