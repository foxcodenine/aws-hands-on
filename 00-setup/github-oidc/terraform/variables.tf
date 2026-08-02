variable "github_owner" {
  description = "GitHub user or org that owns the repo"
  type        = string
  default     = "foxcodenine"
}

variable "github_repo" {
  description = "Repo allowed to assume the role"
  type        = string
  default     = "aws-hands-on"
}
