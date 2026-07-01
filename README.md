# aws-hands-on

A collection of hands-on AWS tutorials, one numbered folder per topic.

## Tutorials

| # | Topic | Runtimes |
|---|-------|----------|
| 01 | [Create your first Lambda function](./01-create_your_first_lambda_function/) | [Python](./01-create_your_first_lambda_function/python/) · [Node](./01-create_your_first_lambda_function/node/) · [Go](./01-create_your_first_lambda_function/golang/) |

## Structure

Each tutorial lives in a numbered folder (`01-`, `02-`, …) and is self-contained — its own source code, deploy scripts, and notes.

The Go folder includes a `steps/` subdirectory with numbered scripts and docs (`01-` through `05-`) that walk through creating the IAM role, deploying the function, and invoking it via the CLI.
