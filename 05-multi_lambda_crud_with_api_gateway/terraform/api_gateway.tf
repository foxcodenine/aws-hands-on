# ==================================================================================================
# API GATEWAY
#
#   POST   /users            -> create_user
#   GET    /users            -> list_users
#   GET    /users/{userID}   -> get_user
#   PUT    /users/{userID}   -> update_user
#   DELETE /users/{userID}   -> delete_user
#   POST   /echo             -> echo
# ==================================================================================================

resource "aws_api_gateway_rest_api" "api" {
  name = "${var.prefix}${var.lesson}-api"
}

# --- The URL paths --------------------------------------------------------------------------------

resource "aws_api_gateway_resource" "users" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "users"
}

# Nested under /users, so the full path is /users/{userID}. The name in the
# braces is what handlers read as req.PathParameters["userID"].
resource "aws_api_gateway_resource" "user_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "{userID}"
}

resource "aws_api_gateway_resource" "echo" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "echo"
}

# --- The routes -----------------------------------------------------------------------------------

# The route table from the header, as data. Each key is a Lambda name from
# local.functions; each value says which path it sits on and which verb it answers.
locals {
  routes = {
    create_user = { path = aws_api_gateway_resource.users.id, method = "POST" }
    list_users  = { path = aws_api_gateway_resource.users.id, method = "GET" }
    get_user    = { path = aws_api_gateway_resource.user_id.id, method = "GET" }
    update_user = { path = aws_api_gateway_resource.user_id.id, method = "PUT" }
    delete_user = { path = aws_api_gateway_resource.user_id.id, method = "DELETE" }
    echo        = { path = aws_api_gateway_resource.echo.id, method = "POST" }
  }
}

# The three blocks below each run once per route. Written out longhand that
# would be eighteen blocks; for_each keeps it at three.

# 1. Method - attaches an HTTP verb to a path, e.g. GET on /users/{userID}. A
#    path with no method exists but rejects everything, and a method on its own
#    still knows nothing about what handles the request.
resource "aws_api_gateway_method" "routes" {
  for_each = local.routes

  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = each.value.path
  http_method   = each.value.method
  authorization = "NONE"
}

# 2. Integration - gives the method above a Lambda to send the request to.
#    The repeated path and verb just name which method that is.
resource "aws_api_gateway_integration" "routes" {
  for_each = local.routes

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = each.value.path
  http_method = each.value.method

  # AWS_PROXY hands the raw request to the function and expects the
  # events.APIGatewayProxyResponse shape the handlers already return.
  type = "AWS_PROXY"
  uri  = aws_lambda_function.functions[each.key].invoke_arn

  # Always POST - even for the GET and DELETE routes. This is not the caller's
  # verb, it is how API Gateway calls the Lambda service, which is always POST.
  integration_http_method = "POST"

  # AWS rejects an integration for a method that does not exist yet, and nothing
  # above mentions the method - so without this, Terraform is free to create both
  # at once and the apply fails intermittently.
  depends_on = [aws_api_gateway_method.routes]
}

# 3. Permission - lets API Gateway actually invoke that Lambda. 
#    Pairs with the role in iam.tf, but points the other way:
#
#      API Gateway --calls--> Lambda --reads--> DynamoDB
#                   ^                  ^
#                   permission         role (iam.tf)
#
#    Missing role: the Lambda runs, then fails on DynamoDB with AccessDenied.
#    Missing this perm : the Lambda never runs - a 500 with empty logs.
resource "aws_lambda_permission" "routes" {
  for_each = local.routes

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.functions[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

# --- Deployment and stage --------------------------------------------------------------------------

# This is the "Deploy API" button from the console. Everything
# above is only a draft - the URL keeps serving the last published copy until
# something deploys.
#
# The catch: Terraform only rebuilds a resource when one of its arguments
# changes, and this block's only real argument (rest_api_id) never does. So
# adding a route would create the method and integration, skip the deployment,
# and leave the live URL unchanged - with apply reporting success.
#
# triggers exists purely to change. It squashes every route into one short
# string, so editing any route changes the string, which makes Terraform deploy.
resource "aws_api_gateway_deployment" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_method.routes,
      aws_api_gateway_integration.routes,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

# The stage is what actually gets a URL - see the invoke_url output.
resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.api.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = "test"
}
