# --- Create a DynamoDB table ----------------------------------------------------------------------

resource "aws_dynamodb_table" "lambda_apigatway" {
  # must match `tableName` in golang/main.go exactly
  name = "lambda-apigateway"

  # on-demand pricing - no read/write capacity to configure
  billing_mode = "PAY_PER_REQUEST"

  # the partition key - every item needs one, used to look items up
  hash_key = "id"

  # DynamoDB only requires a type declaration for attributes used as a
  # table key (or index key) - not every field your items will have.
  attribute {
    name = "id"
    # S = string, N = number, B = binary
    type = "S"
  }
}
