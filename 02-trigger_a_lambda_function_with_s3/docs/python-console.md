# Python with the AWS Console

Follow the [AWS S3 trigger tutorial](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html) using Python and the AWS Console.

## What I did

- Created the `aws-hands-on-<account-id>` bucket manually in the console.
- Created an IAM role with an S3 read policy, but the Lambda creation wizard defaulted to auto-creating its own role instead — ended up running under `s3-trigger-tutorial-role-y3hyiajl`, not the role I made.
- Deployed the sample `lambda_handler` code from the AWS docs, which reads the object back with `s3.get_object` to print its `ContentType`.
- Configured the bucket as an S3 event trigger for `ObjectCreated` events.
- Tested with the console's dummy event first, then a real upload.

See the [README](../README.md#python--difficulties-and-solutions) for the errors I hit along the way and how I fixed them.
