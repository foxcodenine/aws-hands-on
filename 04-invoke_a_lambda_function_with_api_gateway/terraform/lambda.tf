# --- Create the Lambda function -------------------------------------------------------------------

# Build the binary before running terraform apply (Terraform only zips
# whatever file already exists here, it does not compile Go for you):
#   cd 04-invoke_a_lambda_function_with_api_gateway/golang
#   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

# Zips the compiled binary into the file Lambda actually accepts as a deployment package.
data "archive_file" "lambda_apigatway" {
  # only "zip" is supported
  type = "zip"

  # the compiled binary to package
  source_file = "${path.module}/../golang/bootstrap"

  # where to write the resulting .zip
  output_path = "${path.module}/../golang/bootstrap.zip"
}

resource "aws_lambda_function" "lambda_apigatway" {
  # name shown in the Lambda console - kept short since Lambda caps
  # function names at 64 characters (var.lesson alone is too long)
  function_name = "${var.prefix}04-lambda-apigateway-function"

  # execution role - what the function is allowed to do
  role = aws_iam_role.lambda_apigatway.arn

  # must match the compiled binary's filename
  handler = "bootstrap"

  # custom runtime - required for compiled Go binaries
  runtime = "provided.al2023"

  # the zip to upload
  filename = data.archive_file.lambda_apigatway.output_path

  # lets Terraform detect code changes and redeploy
  source_code_hash = data.archive_file.lambda_apigatway.output_base64sha256
}
