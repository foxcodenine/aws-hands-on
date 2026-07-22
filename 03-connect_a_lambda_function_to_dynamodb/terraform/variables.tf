variable "aws_region" {
  description = "AWS region to deploy the DynamoDB table into"
  type        = string
  default     = "eu-west-1"
}

# ----------------------------------------------------------------------

variable "environment" {
  description = "Deployment environment name, used as a prefix for resource names (e.g. dev, staging, prod)"
  type        = string
  default     = "learning"

  validation {
    condition     = contains(["learning", "dev", "staging", "prod"], var.environment)
    error_message = "environment must be one of: learning, dev, staging, prod."
  }
}

# ----------------------------------------------------------------------

variable "table_base_name" {
  description = "Base name of the table, WITHOUT environment prefix (the prefix is added automatically in locals.tf)"
  type        = string
  default     = "Users"
}

# ----------------------------------------------------------------------

variable "billing_mode" {
  description = "DynamoDB billing mode: PAY_PER_REQUEST (on-demand) or PROVISIONED"
  type        = string
  default     = "PAY_PER_REQUEST"

  validation {
    condition     = contains(["PAY_PER_REQUEST", "PROVISIONED"], var.billing_mode)
    error_message = "billing_mode must be PAY_PER_REQUEST or PROVISIONED."
  }
}

# ----------------------------------------------------------------------

variable "read_capacity" {
  description = "Provisioned RCUs (only used when billing_mode = PROVISIONED)"
  type        = number
  default     = 5
}

variable "write_capacity" {
  description = "Provisioned WCUs (only used when billing_mode = PROVISIONED)"
  type        = number
  default     = 5
}

# ----------------------------------------------------------------------

variable "enable_point_in_time_recovery" {
  description = "Enable continuous backups / point-in-time recovery"
  type        = bool
  default     = true
}

# ----------------------------------------------------------------------

variable "extra_tags" {
  description = "Additional tags to merge into every resource, on top of the common tags in locals.tf"
  type        = map(string)
  default     = {}
}

# ----------------------------------------------------------------------