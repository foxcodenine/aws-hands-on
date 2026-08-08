# Tutorial 05 — next steps

What is left to do. Everything already built is in the [README](../README.md).

## 1. Add a CloudWatch Lambda error alarm

Lambda automatically publishes an `Errors` metric.

I can create:

- An `aws_cloudwatch_metric_alarm`
- An SNS topic
- An email subscription

This would notify me when Lambda functions start returning unexpected errors.

The handlers already distinguish between:

```text
reject → expected client error, such as HTTP 4xx
fail   → unexpected server error, such as HTTP 5xx
```

This distinction is important.

Expected user errors should not trigger operational alarms. Otherwise, the alarm would fire too often and become easy to ignore.

---

## 2. Set CloudWatch log retention

The Lambda log groups are currently created automatically by AWS.

By default:

- Logs never expire.
- Terraform does not manage the log groups.
- `terraform destroy` may leave the log groups behind.

I should declare each log group in Terraform:

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/example"
  retention_in_days = 14
}
```

This gives Terraform control over the log groups and prevents logs from being stored forever.

---

## 3. Enable AWS X-Ray tracing

CloudWatch logs explain what happened.

X-Ray helps show where request time was spent.

Example:

```text
API Gateway      5 ms
Lambda         420 ms
Cold start     380 ms
DynamoDB        28 ms
```

To enable it, the Lambda requires:

```hcl
tracing_config {
  mode = "Active"
}
```

The IAM role also needs permission such as:

```text
xray:PutTraceSegments
xray:PutTelemetryRecords
```

X-Ray has limited value for this small project because each request currently uses only one Lambda and one DynamoDB table.

It is still useful to implement once for learning and interview preparation.

---

## Later

### Terraform modules

The Lambda resources currently repeat the same pattern:

- Lambda function
- IAM role
- IAM policy
- API Gateway permission
- CloudWatch configuration

This could later be extracted into a reusable Terraform module.

Modules reduce duplication and create a consistent interface for deploying new Lambda functions.

I am leaving this until the more important deployment, monitoring and state-management work is complete.
