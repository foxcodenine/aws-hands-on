resource "aws_dynamodb_table" "users" {
  name = local.table_name

  billing_mode = var.billing_mode

  # Only meaningful when billing_mode = PROVISIONED; ignored by AWS when
  # PAY_PER_REQUEST, but Terraform still wants *something* here - null is
  # valid and means "not applicable".
  read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null

  # The partition key - determines how data is spread across storage
  # partitions, and every read/write needs it. Must match the
  # `dynamodbav:"user_id"` tag in the Go struct exactly (case-sensitive
  # string, no magic binding between Terraform/Go/DynamoDB).
  #
  # No sort key here since each user is looked up as a standalone item,
  # not as part of a one-to-many collection under a shared partition key.
  hash_key = "user_id"

  # DynamoDB only requires you to declare the TYPE (S/N/B) for attributes
  # used as a table key or GSI key - not every field your items will have.
  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  # Global Secondary Index: lets you query by "status" without a full
  # table Scan, same as the QueryByStatus function in the Go repo.
  # Note: "status" is low-cardinality (most users likely "active"), so
  # this GSI partition could get disproportionate traffic at real scale
  # ("hot partition"). Fine for learning; production would usually add a
  # sort key (e.g. created_at) or shard the key to spread the load.
  global_secondary_index {
    name     = "status-index"
    hash_key = "status"

    # ALL = copy every item attribute into the index, so a query against
    # this GSI doesn't need a round-trip back to the main table.
    projection_type = "ALL"

    read_capacity  = local.gsi_read_capacity
    write_capacity = local.gsi_write_capacity
  }

  # Point-in-time recovery: continuous backups, restore to any second in
  # the last 35 days. This is a nested block, not a top-level argument -
  # that's just how the AWS provider models it.
  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  # Encrypts the table at rest with an AWS-owned key. Set explicitly to
  # document intent rather than relying on the implicit platform default.
  server_side_encryption {
    enabled = true
  }

  # DynamoDB Streams (change feed, e.g. to trigger a Lambda on writes) is
  # off by default - not needed for a basic CRUD table.

  # `lifecycle` controls how Terraform itself behaves toward this
  # resource, not the resource's AWS configuration.
  lifecycle {
    # Guards against `terraform destroy` accidentally deleting a table
    # with real data - flip to true once this holds data you care about.
    prevent_destroy = false
  }

  tags = merge(
    { Name = local.table_name },
    var.extra_tags
  )
}
