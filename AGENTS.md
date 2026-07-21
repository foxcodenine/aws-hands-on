# AGENTS.md

## What this repo is

Hands-on AWS learning — one numbered tutorial folder per topic (`01-`, `02-`, …). Each folder is self-contained with its own source code, deploy scripts, and notes.

## Conventions

- Folders are numbered in order: `01-create_your_first_lambda_function/`, next will be `02-…`, etc.
- Inside each tutorial, scripts and docs go in a `steps/` subdirectory, numbered from `01-`.
- Source code sits at the top level of its runtime folder (e.g. `golang/main.go`), not inside `steps/`.
- Terraform configuration sits in a `terraform/` subdirectory inside the relevant tutorial.

## AWS / tooling notes

- Go Lambdas use the `provided.al2023` runtime. Build: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go`, zip the `bootstrap` binary, set handler to `bootstrap`.
- AWS CLI profile in use: `developer`.
- Terraform uses the `developer` AWS profile and must not commit state files, `.terraform/`, plans, or local `.tfvars` files.
- Invoking Lambda via CLI requires `--cli-binary-format raw-in-base64-out` for raw JSON payloads.
- IAM execution role: `lambda-basic-execution` with `AWSLambdaBasicExecutionRole` attached.

## Preferences

- Docs written from a first-person learner perspective (what I built, what I learnt, what tripped me up).
- Keep READMEs short — this is a learning repo, not a product.
- No unnecessary comments in code; only add one when the why is non-obvious.
- Do not write tutorial implementation code unless explicitly asked. Prefer reviewing the learner's work, explaining errors, and providing guidance.
