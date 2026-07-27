# --- Create policy --------------------------------------------------------------------------------

resource "aws_iam_policy" "lambda_apigatway" {
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
        "Resource" : aws_dynamodb_table.lambda_apigatway.arn
      },
      {
        "Sid" : "AllowCloudWatchLogs",

        "Action" : [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],

        "Effect" : "Allow",
        "Resource" : "*"
      }
    ]
  })
}


# --- Create an execution role ---------------------------------------------------------------------

resource "aws_iam_role" "lambda_apigatway" {
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
    }
  )
}

# Links the role to the policy above - a role has no permissions on its own
# until a policy is attached to it.
resource "aws_iam_role_policy_attachment" "lambda_apigatway" {
  role       = aws_iam_role.lambda_apigatway.name
  policy_arn = aws_iam_policy.lambda_apigatway.arn
}
