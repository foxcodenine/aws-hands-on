# Next steps

Things I have decided to do, roughly in order. Nothing here is started yet.

## 1. Remote state

Move `terraform.tfstate` off my laptop and into S3. I have done this in the
Udemy course, so it is mostly a case of applying it here.

Why it matters for this repo:

- state is currently local and gitignored, so nothing else can ever run
  `terraform plan` or `apply` - including GitHub Actions
- if I lose the file, Terraform forgets everything it built and would try to
  create it all again

What it needs:

- an S3 bucket for the state (I already have `terraform-state-s3-bucket-cf12`)
- locking, so two runs cannot apply at once. Newer AWS provider versions support
  S3-native locking, so a separate DynamoDB lock table may not be needed - check
  what the provider wants at the time
- `terraform init -migrate-state` to move the existing local state across

Do it on one tutorial first, not all five at once.

## 2. CloudWatch alarm on the error rate

Get told when something breaks, instead of finding out by calling the API.

Lambda already publishes an `Errors` metric to CloudWatch on its own - no code
needed. The alarm just watches it:

- `aws_cloudwatch_metric_alarm` on the Lambda `Errors` metric
- an SNS topic + email subscription so it can actually reach me

This is worth doing now that the handlers split `reject` (4xx, logged WARN) from
`fail` (5xx, logged ERROR). An alarm is only useful if ERROR means something is
genuinely wrong - if bad user input counted, it would fire constantly and I
would learn to ignore it.

## 3. X-Ray tracing

Shows where the *time* goes in a request, rather than why it failed:

```
API Gateway    5ms
  Lambda     420ms   (cold start 380ms)
    DynamoDB  28ms
```

Needs `tracing_config { mode = "Active" }` on the Lambda plus
`xray:PutTraceSegments` in the IAM policy.

Lower value than the alarm at this scale - one Lambda calling one table makes a
shallow trace. Worth doing once so I have seen it, but not urgent.

## 4. OIDC, then plan and apply in CI

Once state is in S3, GitHub Actions can actually run Terraform.

OIDC lets GitHub assume an IAM role directly, so there are no long-lived AWS
keys stored as repo secrets. Needs an IAM OIDC provider for
`token.actions.githubusercontent.com` and a role whose trust policy is scoped to
this repo.

Then: `plan` on pull requests, `apply` on merge to main. That last one also
fixes the build/apply drift that bit me in 05 - building and applying become one
step that cannot get out of sync.

See `notes/github-actions-ideas.md` for the fuller version.

## Later, not now

- **Modules.** Extracting the repeated "Lambda + IAM role + permission" pattern
  into a reusable module. Commonly asked about in interviews, and a natural fit
  since it is the same abstraction instinct as any other code. Parking it until
  the tutorials above are done.
