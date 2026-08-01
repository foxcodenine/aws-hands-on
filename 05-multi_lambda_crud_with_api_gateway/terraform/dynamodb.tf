# --- Create a DynamoDB table ----------------------------------------------------------------------

# Unlike tutorial 04's single-key table, this one matches what the repository in
# golang/internal/repository actually queries: a user_id partition key plus two
# GSIs. Without email-index every POST /users fails, because Create looks up the
# email first to reject duplicates.
resource "aws_dynamodb_table" "users" {
  name = local.table_name

  # on-demand pricing - no read/write capacity to configure
  billing_mode = "PAY_PER_REQUEST"

  # the partition key, matching `dynamodbav:"user_id"` on models.User
  hash_key = "user_id"

  # Only key attributes get declared - user_id for the table, email and status
  # for the two indexes below. Every other field on the item is schemaless.
  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "email"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  # Used by UserRepository.QueryByEmail - lets Create check for a duplicate
  # email without scanning the whole table.
  global_secondary_index {
    name            = "email-index"
    hash_key        = "email"
    projection_type = "ALL"
  }

  # Used by UserRepository.QueryByStatus.
  global_secondary_index {
    name            = "status-index"
    hash_key        = "status"
    projection_type = "ALL"
  }
}
