# ==================================================================================================
# S3 BUCKET
# ==================================================================================================

# --- Create bucket --------------------------------------------------------------------------------
resource "aws_s3_bucket" "hands_on" {

  bucket = "aws-hands-on-${data.aws_caller_identity.current.account_id}"

  # Lets `terraform destroy` delete this bucket even if it still has objects in it.
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "hands_on" {
  bucket = aws_s3_bucket.hands_on.id

  versioning_configuration {
    status = "Enabled"
  }
}

# --- Upload object --------------------------------------------------------------------------------

# aws_s3_object with a key ending in "/" and no content creates an empty
# "folder" placeholder — S3 has no real directories, this just makes one show up in the console.
resource "aws_s3_object" "trigger_dir" {
  bucket = aws_s3_bucket.hands_on.id
  key    = "${var.lesson}/"
}

resource "aws_s3_object" "test_upload" {
  bucket = aws_s3_bucket.hands_on.id
  key    = "${aws_s3_object.trigger_dir.key}_me.jpg"
  source = "${path.module}/../data/_me.jpg"

  # etag (MD5 hash of the file) lets Terraform detect if the local file changed
  # and needs re-uploading — without it, Terraform won't notice edits to _me.jpg.
  etag = filemd5("${path.module}/../data/_me.jpg")

  #  content_type matters here specifically because your Lambda's whole job is to
  # read back ContentType — S3 won't guess it correctly without this set explicitly.
  content_type = "image/jpeg"
}

output "lambda_trigger_bucket" {
  value = aws_s3_bucket.hands_on.bucket
}

# ==================================================================================================
# FUNCTION
# ==================================================================================================

# --- Create policy --------------------------------------------------------------------------------

# What the role is ALLOWED TO DO once AWS lets it act (identity-based permissions).
resource "aws_iam_policy" "s3_trigger" {
  name = "${var.lesson}-policy"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        # Lets the function write logs to CloudWatch — same as AWSLambdaBasicExecutionRole.
        "Effect" : "Allow",
        "Action" : [
          "logs:PutLogEvents",
          "logs:CreateLogGroup",
          "logs:CreateLogStream"
        ],
        "Resource" : "arn:aws:logs:*:*:*"
      },
      {
        # ListBucket is a bucket-level action (no "/*"), GetObject is object-level ("/*") —
        # they need separate resource ARNs, one won't cover the other.
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:ListBucket"
        ],
        "Resource" : [
          "arn:aws:s3:::aws-hands-on-${data.aws_caller_identity.current.account_id}",
          "arn:aws:s3:::aws-hands-on-${data.aws_caller_identity.current.account_id}/*"
        ]
      }
    ]
  })
}

# --- Create role ----------------------------------------------------------------------------------

# WHO is allowed to assume this role (trust policy) — different from the policy above,
# which says what the role can do once assumed.
resource "aws_iam_role" "ns3_trigger" {
  name = "${var.lesson}-role"
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      }
    ]
    }
  )
}

# Links the role to the policy — the role stays empty of permissions until attached.
resource "aws_iam_role_policy_attachment" "s3_trigger" {
  role       = aws_iam_role.ns3_trigger.name
  policy_arn = aws_iam_policy.s3_trigger.arn
}

# --- Create function / Deploy code ----------------------------------------------------------------

# Zips the compiled Go binary (built separately with `go build`) so it can be uploaded.
# Terraform does not build the Go code — "bootstrap" must already exist before apply.
data "archive_file" "s3_trigger" {
  type        = "zip"
  source_file = "${path.module}/../golang/bootstrap"
  output_path = "${path.module}/../golang/bootstrap.zip"
}

resource "aws_lambda_function" "s3_trigger" {
  function_name = "${var.lesson}-function"
  role          = aws_iam_role.ns3_trigger.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"

  filename = data.archive_file.s3_trigger.output_path

  # Tells Terraform to redeploy the function whenever the zipped code actually changes.
  source_code_hash = data.archive_file.s3_trigger.output_base64sha256
}

# ==================================================================================================
# TRIGGER
# ==================================================================================================

# --- Configure trigger -----------------------------------------------------------------------------

# Resource-based permission (on the Lambda side): allows the S3 SERVICE itself to invoke
# this function. This is separate from the IAM role above, which controls what the
# function is allowed to do, not who is allowed to call it.
resource "aws_lambda_permission" "allow_s3" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.s3_trigger.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.hands_on.arn
}

# The actual trigger: tells the bucket to call the Lambda on new uploads under
# the tutorial folder prefix. depends_on ensures the permission above exists first —
# otherwise S3 would reject the notification for lacking invoke permission.
resource "aws_s3_bucket_notification" "s3_trigger" {
  bucket = aws_s3_bucket.hands_on.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.s3_trigger.arn
    events              = ["s3:ObjectCreated:*"]
    filter_prefix       = "${var.lesson}/"
  }

  depends_on = [aws_lambda_permission.allow_s3]
}
