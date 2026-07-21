# 02 — Trigger a Lambda Function with S3

In this tutorial I am learning how to invoke a Lambda function automatically when a file is uploaded to an S3 bucket. I will first compare Python, Node.js, and Go, then deploy the Go version with Bash and Terraform.

Tutorial: [Using an Amazon S3 trigger to invoke a Lambda function](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html)

## What I am building

```text
file upload → S3 bucket → Lambda function → CloudWatch Logs
```

## Project structure

- `python/` — Python Lambda source code
- `node/` — Node.js Lambda source code
- `golang/` — Go Lambda source code
- `terraform/` — AWS infrastructure
- `steps/` — my walkthrough and learning notes

## Guide

I am building each implementation progressively. The `steps/` folder separates each version of the tutorial.

## What I want to learn

- How S3 event payloads are handled in Python, Node.js, and Go
- How manual console deployment compares with the AWS CLI
- How Terraform tracks AWS resources
- How Lambda execution roles differ from resource-based permissions

## Python — difficulties and solutions

- **`AccessDenied` on `s3:ListBucket`** — execution role had no S3 permissions at all (`AWSLambdaBasicExecutionRole` only covers CloudWatch Logs); fixed by adding a policy granting `s3:GetObject` and `s3:ListBucket`.
- **Same `AccessDenied` error persisted after adding the policy** — the policy was correct but attached to the wrong role. The Lambda function was actually running under an auto-generated role (`s3-trigger-tutorial-role-y3hyiajl`), not the role I had manually created and edited. Root cause: the Lambda console defaults to auto-creating a new execution role unless you explicitly pick "Use an existing role" during function setup.
- **`NoSuchKey` once permissions were fixed** — the test event's object key didn't include the folder prefix the file was actually uploaded under; fixed by adding the prefix to the key `"key": "02-trigger_a_lambda_function_with_s3/_prily.jpg",`.

## Go / Terraform — difficulties and solutions

- **No CloudWatch log entries after `terraform apply`, even after deleting and re-uploading the test file** — the notification config and the Lambda's resource-based policy both looked correct via the CLI, and a direct `aws lambda invoke` with a synthetic S3 event worked fine, which ruled out the code and IAM permissions. A real upload eventually did trigger it — best guess is a short propagation delay between setting a bucket's notification configuration and S3 actually starting to deliver events for it. Lesson: don't trust the first failed test right after `apply`; retry a bit later before assuming the wiring is wrong.
- Used the AWS docs' example S3 event JSON to test the handler directly from the Lambda console/CLI without needing a real upload — just swap in the real bucket name, region, and a URL-encoded object key.
