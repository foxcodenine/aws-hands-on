locals {

  table_name = "${var.environment}-${var.table_base_name}"

  gsi_read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  gsi_write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
}