# Go with Terraform

Recreate the S3 trigger tutorial for Go, using Terraform instead of the console. Skipping a separate Go console pass — going straight from Node.js to Terraform.

## What I did

- Set up the provider (`eu-west-1`, `developer` profile) with `default_tags` (`Project`, `Tutorial`, `Environment`) so everything created gets tagged automatically.
- Created the `aws-hands-on-<account-id>` S3 bucket with versioning enabled and `force_destroy = true`, via `aws_s3_bucket` + `aws_s3_bucket_versioning`.
- Added an `aws_s3_object` resource (`trigger_dir`) to create the `02-trigger_a_lambda_function_with_s3/` folder placeholder in the bucket — a zero-byte object with a trailing `/` key.
- Added another `aws_s3_object` resource (`test_upload`) to upload a real test file (`data/_me.jpg`) into that folder, using `filemd5()` for the `etag` so Terraform detects local file changes, and an explicit `content_type` since the Lambda's job is to read that back.
- Wrote the Go handler (`golang/main.go`) — reads the bucket/key off the S3 event, URL-decodes the key, then calls `HeadObject` to read back `ContentType`, same pattern as the Node version. Built with `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go` for the `provided.al2023` runtime.
- Wrote the IAM role + policy in Terraform this time instead of the console: a trust policy letting `lambda.amazonaws.com` assume the role, plus a permissions policy for CloudWatch logging and `s3:GetObject`/`s3:ListBucket` on the bucket.
- Added `aws_lambda_function`, using a `data "archive_file"` to zip the compiled `bootstrap` binary — Terraform doesn't build the Go code, so `bootstrap` has to exist before `apply`.
- Wired up the trigger with `aws_lambda_permission` (lets the S3 *service* invoke the function) and `aws_s3_bucket_notification` (tells the bucket to actually call it on `ObjectCreated:*`, filtered to the tutorial's key prefix).

## What tripped me up

After `apply`, deleting and re-uploading the test file produced nothing in CloudWatch Logs — no invocation at all. Checked the notification config and the Lambda's resource policy via the CLI and both looked correct; a direct `aws lambda invoke` with a synthetic S3 event also worked fine and returned `image/jpeg`, which confirmed the function and its IAM permissions were fine — the problem was specifically the S3→Lambda event delivery, not the code.

Eventually a real upload did fire it. Best guess: notification-configuration changes on a bucket aren't instant — there can be a short propagation delay before S3 actually starts delivering events for a newly created (or newly changed) notification config, even though the API immediately reports it as set. Worth remembering for next time: don't trust the first failed test right after `apply` — retry a bit later before assuming the wiring is wrong.

The AWS docs' example test event (from the [S3 trigger tutorial](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html)) is handy for testing the handler directly in the Lambda console without needing a real S3 event — just swap in the real bucket name/region and URL-encode the object key.

## Wrap-up

Destroyed the stack with `terraform destroy` once the trigger was confirmed working, since this tutorial's resources are billable and no longer needed running.
