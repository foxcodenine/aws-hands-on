# aws-hands-on

A collection of hands-on AWS tutorials, one numbered folder per topic.

## Tutorials

| # | Topic | Runtimes |
|---|-------|----------|
| 01 | [Create your first Lambda function](https://docs.aws.amazon.com/lambda/latest/dg/getting-started.html) | [Python](./01-create_your_first_lambda_function/python/) · [Node](./01-create_your_first_lambda_function/node/) · [Go](./01-create_your_first_lambda_function/golang/) |
| 02 | [Trigger a Lambda function with S3](https://docs.aws.amazon.com/lambda/latest/dg/with-s3-example.html) | [Python](./02-trigger_a_lambda_function_with_s3/python/) · [Node](./02-trigger_a_lambda_function_with_s3/node/) · [Go](./02-trigger_a_lambda_function_with_s3/golang/) · [Terraform](./02-trigger_a_lambda_function_with_s3/terraform/) |
| 03 | [DynamoDB CRUD with Go](https://oneuptime.com/blog/post/2026-02-12-use-dynamodb-with-go/view) | [Go](./03-dynamodb_crud_with_go/golang/) · [Terraform](./03-dynamodb_crud_with_go/terraform/) |
| 04 | [Invoke a Lambda function with API Gateway](https://docs.aws.amazon.com/lambda/latest/dg/services-apigateway-tutorial.html) | [Python](./04-invoke_a_lambda_function_with_api_gateway/python/) · [Node](./04-invoke_a_lambda_function_with_api_gateway/node/) · [Go](./04-invoke_a_lambda_function_with_api_gateway/golang/) · [Terraform](./04-invoke_a_lambda_function_with_api_gateway/terraform/) |
| 05 | [Multi-Lambda CRUD with API Gateway](./05-multi_lambda_crud_with_api_gateway/) | [Go](./05-multi_lambda_crud_with_api_gateway/golang/) · [Terraform](./05-multi_lambda_crud_with_api_gateway/terraform/) |

## Structure

Each tutorial lives in a numbered folder (`01-`, `02-`, …) and is self-contained — its own source code, deploy scripts, infrastructure, and notes.

Each tutorial's `docs/` directory records the learning path. Where the order matters, the walkthrough is numbered under `docs/steps/`; runnable scripts live in `scripts/`.

## CI

CI lives in `.github/workflows/`, one file per tutorial. Tutorial 05 also deploys itself: a push to `main` builds the Lambdas and runs `terraform apply`, authenticating with OIDC so there are no AWS keys stored in the repo.
