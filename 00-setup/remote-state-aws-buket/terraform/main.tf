

resource "aws_s3_bucket" "tfstate" {
  bucket = "aws-hands-on-tfstate-${data.aws_caller_identity.current.account_id}"

  # No force_destroy: deleting this by accident loses the record of every stack.
  # force_destroy = true

  # And Terraform refuses to even plan a destroy of this bucket. To delete it I
  # have to come here and remove this block first, which turns an accident into
  # a decision.
  #
  # It blocks replacement too, so a change that would force a new bucket errors
  # instead of quietly recreating one - and every other tutorial's state lives
  # in here.
  #
  # Only binds Terraform. The console and the CLI do not know about it.
  lifecycle {
    prevent_destroy = true
  }
}

# Keeps old versions, so a corrupted or truncated state can be rolled back.
resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id

  versioning_configuration {
    status = "Enabled"
  }
}


# State files can contain secrets in plain text.
resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id

  # There are two ways to make an S3 bucket public: an ACL (older, per-object)
  # or a bucket policy (newer). 

  # Each gets two settings - one that blocks new attempts, one that neutralises 
  # anything already set.

  # Reject any NEW request that would grant public access via an ACL.
  block_public_acls = true

  # Reject any NEW bucket policy that would grant public access.
  block_public_policy = true

  # Ignore public ACLs that are ALREADY on the bucket or its objects, rather
  # than only stopping new ones.
  ignore_public_acls = true

  # Stop honouring an EXISTING public bucket policy - only this account can use
  # the bucket, whatever the policy says.
  restrict_public_buckets = true
}


# prevent_destroy above only binds Terraform. This binds AWS itself, so the
# console and the CLI are covered too.
#
# An explicit Deny always wins in IAM - it cannot be overridden by any Allow, on
# any role, including an administrator.
#
# The deliberate gap: nothing here denies PutBucketPolicy, so an admin can
# remove this policy and then delete the bucket. That is the escape hatch, not
# an oversight. Denying the policy actions as well would lock the bucket down
# permanently, with no way back.
#
# Only DeleteBucket is denied. DeleteObject stays allowed, because Terraform
# creates and removes a .tflock file on every run.
resource "aws_s3_bucket_policy" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "DenyBucketDeletion",
        "Effect" : "Deny",
        "Principal" : "*",
        "Action" : "s3:DeleteBucket",
        "Resource" : aws_s3_bucket.tfstate.arn
      }
    ]
  })
}