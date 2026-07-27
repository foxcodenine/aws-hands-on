# ------------------------------------------------------------------------------
# 1. The REST API itself


resource "aws_api_gateway_rest_api" "lambda_apigatway" {
  name = "${var.prefix}${var.lesson}-api"
}

# ------------------------------------------------------------------------------
# 2. A resource (the URL path segment) — matches the tutorial's /DynamoDBManager


resource "aws_api_gateway_resource" "lambda_apigatway" {
  rest_api_id = aws_api_gateway_rest_api.lambda_apigatway.id
  parent_id   = aws_api_gateway_rest_api.lambda_apigatway.root_resource_id
  path_part   = "DynamoDBManager"
}

# ------------------------------------------------------------------------------
# 3. The POST method on that resource


resource "aws_api_gateway_method" "lambda_apigatway" {
  rest_api_id   = aws_api_gateway_rest_api.lambda_apigatway.id
  resource_id   = aws_api_gateway_resource.lambda_apigatway.id
  http_method   = "POST"
  authorization = "NONE"
}

# ------------------------------------------------------------------------------
# 4. The integration — wires the method to your Lambda (proxy integration means 
# API Gateway forwards the raw request and expects the 
# events.APIGatewayProxyResponse shape your Go code already returns)


resource "aws_api_gateway_integration" "lambda_apigatway" {
  rest_api_id             = aws_api_gateway_rest_api.lambda_apigatway.id
  resource_id             = aws_api_gateway_resource.lambda_apigatway.id
  http_method             = aws_api_gateway_method.lambda_apigatway.http_method
  integration_http_method = aws_api_gateway_method.lambda_apigatway.http_method
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda_apigatway.invoke_arn
}

# ------------------------------------------------------------------------------
# 5. Permission — same concept as aws_lambda_permission in your 02 tutorial 
# (S3 → Lambda), except this time it's API Gateway that needs invoke rights


resource "aws_lambda_permission" "apigatway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_apigatway.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.lambda_apigatway.execution_arn}/*/*"
}

# ------------------------------------------------------------------------------
# 6. Deploy it — a deployment + stage, which is what actually gives you 
# the invoke URL (matches "Deploy the API" in your steps list)


resource "aws_api_gateway_deployment" "lambda_apigatway" {
  rest_api_id = aws_api_gateway_rest_api.lambda_apigatway.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.lambda_apigatway.id,
      aws_api_gateway_method.lambda_apigatway.id,
      aws_api_gateway_integration.lambda_apigatway.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.lambda_apigatway.id
  rest_api_id   = aws_api_gateway_rest_api.lambda_apigatway.id
  stage_name    = "test"
}

output "invoke_url" {
  value = "${aws_api_gateway_stage.test.invoke_url}${aws_api_gateway_resource.lambda_apigatway.path}"
}