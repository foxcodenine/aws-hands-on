/*
Note: 
Build the binaries before running terraform apply - Terraform only zips what
is already on disk, it does not compile Go:

  cd 05-multi_lambda_crud_with_api_gateway/golang && ./build.sh

That script writes build/<function>/bootstrap for every folder in cmd/, which
is what the for_each below relies on.
*/

# --- Create the Lambda functions ------------------------------------------------------------------

locals {
  table_name = "${var.prefix}05-users"

  # One entry per folder in golang/cmd/. lambda.tf loops over this to build the
  # zip, the function and the invoke permission, so adding a seventh Lambda means
  # adding a cmd/ folder and one line here - not another three resource blocks.
  functions = [
    "create_user",
    "list_users",
    "get_user",
    "update_user",
    "delete_user",
    "echo",
  ]
}

# for_each turns one block into six, keyed by function name. Referring to a
# single one later looks like: aws_lambda_function.functions["get_user"].
data "archive_file" "functions" {
  for_each = toset(local.functions)

  type        = "zip"
  source_file = "${path.module}/../golang/build/${each.key}/bootstrap"
  output_path = "${path.module}/../golang/build/${each.key}.zip"
}

resource "aws_lambda_function" "functions" {
  for_each = toset(local.functions)

  # e.g. aws-hands-on-05-get_user. Kept short deliberately: Lambda caps function
  # names at 64 characters, and var.lesson alone is already 38.
  function_name = "${var.prefix}05-${each.key}"

  role = aws_iam_role.lambda_apigateway.arn

  # must match the compiled binary's filename
  handler = "bootstrap"

  # custom runtime - required for compiled Go binaries
  runtime = "provided.al2023"

  filename = data.archive_file.functions[each.key].output_path

  # redeploys the function whenever the zipped code actually changes
  source_code_hash = data.archive_file.functions[each.key].output_base64sha256

  # How the repository learns which table to talk to, instead of hardcoding it.
  environment {
    variables = {
      TABLE_NAME = aws_dynamodb_table.users.name
    }
  }
}
