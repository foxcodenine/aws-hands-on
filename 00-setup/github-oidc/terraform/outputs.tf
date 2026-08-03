# The workflow needs this - it goes in role-to-assume on
# aws-actions/configure-aws-credentials. Not a secret: it is useless without a
# token from the repo named in the trust policy.
output "github_actions_role_arn" {
  description = "Role ARN for the GitHub Actions workflow to assume"
  value       = aws_iam_role.github_actions.arn
}


output "github_actions_apply_role_arn" {
  description = "Role ARN for the apply job - only assumable from main"
  value       = aws_iam_role.github_actions_apply.arn
}