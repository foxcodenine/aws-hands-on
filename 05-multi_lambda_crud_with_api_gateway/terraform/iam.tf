# --- Create policy --------------------------------------------------------------------------------

# All six functions share one role. That keeps the tutorial simple, but it does
# mean the read-only Lambdas can write too - a per-function role would be the
# least-privilege version, at the cost of six roles and six attachments.
resource "aws_iam_policy" "lambda_apigateway" {
  name = "${var.prefix}${var.lesson}-policy"

  policy = jsonencode({
    "Version" : "2012-10-17",

    "Statement" : [
      {
        "Sid" : "AllowDynamoDBAccess",

        "Action" : [
          "dynamodb:DeleteItem",
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:UpdateItem"
        ],

        "Effect" : "Allow",

        # Two ARNs, for the same reason ListBucket and GetObject needed separate
        # ones in tutorial 02: querying a GSI is permissioned on the *index*, not
        # the table. Without the /index/* entry, QueryByEmail fails with
        # AccessDenied even though the table itself is allowed.
        "Resource" : [
          aws_dynamodb_table.users.arn,
          "${aws_dynamodb_table.users.arn}/index/*"
        ]
      },
      {
        "Sid" : "AllowCloudWatchLogs",

        "Action" : [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],

        "Effect" : "Allow",
        "Resource" : "arn:aws:logs:*:*:*"
      }
    ]
  })
}

# --- Create an execution role ---------------------------------------------------------------------

resource "aws_iam_role" "lambda_apigateway" {
  name = "${var.prefix}${var.lesson}-role"

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
  })
}

# Links the role to the policy - a role has no permissions on its own until a
# policy is attached to it.
resource "aws_iam_role_policy_attachment" "lambda_apigateway" {
  role       = aws_iam_role.lambda_apigateway.name
  policy_arn = aws_iam_policy.lambda_apigateway.arn
}
