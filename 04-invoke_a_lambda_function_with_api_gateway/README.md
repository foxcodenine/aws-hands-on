# 04 — Invoke a Lambda Function with API Gateway

In this tutorial I am learning how to expose a Lambda function as an HTTP API using API Gateway, with the function performing CRUD operations on a DynamoDB table. I will follow the AWS tutorial's console walkthrough first, then redo the deployment in Go with Terraform.

Tutorial: [Using Lambda with API Gateway](https://docs.aws.amazon.com/lambda/latest/dg/services-apigateway-tutorial.html)

## What I am building

```text
HTTP request → API Gateway → Lambda function → DynamoDB table
```

## Project structure

- `python/` — Python Lambda source code
- `node/` — Node.js Lambda source code
- `golang/` — Go Lambda source code
- `terraform/` — AWS infrastructure
- `docs/` — my notes:
  - [`apigateway-tutorial.md`](./docs/apigateway-tutorial.md) — the walkthrough
  - [`go-setup.md`](./docs/go-setup.md) — the commands that set the Go module up

## What I want to learn

- How API Gateway routes HTTP requests to a Lambda function
- How to deploy a REST API (resources, methods, stages) via the console vs. Terraform
- How the tutorial's single `operation`-dispatch Lambda pattern translates from Python/Node into Go
- How IAM execution roles grant a Lambda function DynamoDB access


## Steps

1. Create a permissions policy
2. Create an execution role
3. Create the Lambda function
4. Test the function
5. Create a REST API using API Gateway
6. Create a resource on your REST API
7. Create an HTTP POST method
8. Create a DynamoDB table
9. Test the integration of API Gateway, Lambda, and DynamoDB
10. Deploy the API
11. Use curl to invoke your function using HTTP requests
12. Clean up your resources (optional)

## Go / Terraform — what I did

- Wrote the IAM policy (DynamoDB CRUD actions + CloudWatch Logs) and execution role (trust policy for `lambda.amazonaws.com`) in `iam.tf`.
- Created the `lambda-apigateway` DynamoDB table (on-demand billing, `id` partition key) in `dynamodb.tf`.
- Deployed the Go binary via `data "archive_file"` + `aws_lambda_function` in `lambda.tf` — same zip-and-deploy pattern as the 02 tutorial.
- Wired up API Gateway in `api_gatway.tf`: REST API, a `/DynamoDBManager` resource, a `POST` method, an `AWS_PROXY` integration to the Lambda, a resource-based `aws_lambda_permission` letting API Gateway invoke it, and a deployment + `test` stage.
- Split what started as one `main.tf` into `iam.tf`/`lambda.tf`/`dynamodb.tf`, matching the per-resource file layout from 03 instead of one big file.
- Added `golang/http/dynamodb-manager.http` to exercise `echo`/`create`/`read`/`update`/`delete` by hand against the deployed `invoke_url`.

## Go / Terraform — difficulties and solutions

- **`encoding/json` couldn't populate `types.AttributeValue`** — that type is DynamoDB's internal typed wire format (`{"id": {"S": "..."}}`), not what a plain HTTP JSON body sends (`{"id": "..."}`). Python's `boto3` converts this for you automatically; the Go SDK doesn't. Fixed by reading `Item`/`Key`/`ExpressionAttributeValues` as plain `map[string]interface{}` and converting with a small `toAttrs()` helper (wrapping `attributevalue.MarshalMap`) right before each DynamoDB call.
- **`terraform apply` failed with `InvalidParameterValueException: ... must have length less than or equal to 64`** — the Lambda function name built from `${var.prefix}${var.lesson}-function` came out to 68 characters; Lambda function names cap at 64 (IAM role/policy names allow up to 128, which is why those didn't hit this). Fixed by giving the function a short, fixed name (`04-lambda-apigateway-function`) instead of deriving it from the full lesson slug.
- **`echo` returned `{}` instead of the payload sent** — the `Payload` struct is typed specifically for the CRUD fields (`Item`, `Key`, etc.), so arbitrary keys like `{"somekey1": "somevalue1"}` get silently dropped when unmarshaled into it - unlike Python, where `payload` is just a generic dict. Fixed by re-parsing the raw request body generically just for the `echo` case, instead of relying on the typed struct.
