

resource "aws_s3_bucket" "tfstate" {
  bucket = "aws-hands-on-tfstate-${data.aws_caller_identity.current.account_id}"

  # No force_destroy: deleting this by accident loses the record of every stack.
  # force_destroy = true
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
  # or a bucket policy (newer). Each gets two settings - one that blocks new
  # attempts, one that neutralises anything already set.

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