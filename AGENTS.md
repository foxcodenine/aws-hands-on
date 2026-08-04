# AGENTS.md

## What this repo is

Hands-on AWS learning — one numbered tutorial folder per topic (`01-`, `02-`, …). Each folder is self-contained with its own source code, deploy scripts, and notes.

## Conventions

- Folders are numbered in order: `01-create_your_first_lambda_function/`, next will be `02-…`, etc.
- `README.md` stays at the tutorial root.
- Tutorial docs live in `docs/`, unnumbered by default.
- `docs/steps/` with `01-` prefixes is only for a walkthrough that must be followed in order. A number is a promise that order matters — alternatives and reference notes stay unnumbered in `docs/`.
- Runnable scripts go in `scripts/`, not in `docs/`.
- Source code sits at the top level of its runtime folder (e.g. `golang/main.go`).
- Terraform configuration sits in a `terraform/` subdirectory inside the relevant tutorial.

## AWS / tooling notes

- Go Lambdas use the `provided.al2023` runtime. Build: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go`, zip the `bootstrap` binary, set handler to `bootstrap`.
- AWS CLI profile in use: `developer`.
- Terraform uses the `developer` AWS profile and must not commit state files, `.terraform/`, plans, or local `.tfvars` files.
- Tutorial 05 keeps state in S3 (`aws-hands-on-tfstate-<account-id>`, `use_lockfile = true`). Every other tutorial still has local state — moving them is the next planned job.
- 05's `terraform/providers.tf` has `profile` commented out on purpose: CI runners have no named profile. Do not uncomment it.
- Invoking Lambda via CLI requires `--cli-binary-format raw-in-base64-out` for raw JSON payloads.
- IAM execution role: `lambda-basic-execution` with `AWSLambdaBasicExecutionRole` attached.

## CI/CD

- Workflows live in `.github/workflows/`, one file per tutorial, each with a `paths:` filter so it only runs for its own folder.
- **Pushing to `main` deploys tutorial 05.** The workflow builds the Go binaries and runs `terraform apply -auto-approve` in the same job — that pairing is deliberate, since Terraform zips the binaries but never compiles them. Any push touching `05-multi_lambda_crud_with_api_gateway/**` changes real AWS resources.
- Auth is OIDC — no AWS keys in the repo. Two roles, both in `00-setup/github-oidc/`: `aws-hands-on-github-actions` (read-only, any branch, used by plan) and `aws-hands-on-github-actions-apply` (write, `refs/heads/main` only).
- Current state of the work and what is planned next: `05-multi_lambda_crud_with_api_gateway/docs/next-steps.md`.

## Git

- **Never run `git commit`, `git push`, or create branches.** I do all of that
  myself. Propose the commits instead: a one-line message plus the files that
  belong in each, and I will run them.

## Preferences

- Docs written from a first-person learner perspective (what I built, what I learnt, what tripped me up).
- Keep READMEs short — this is a learning repo, not a product.
- No unnecessary comments in code; only add one when the why is non-obvious.
- Do not write tutorial implementation code unless explicitly asked. Prefer reviewing the learner's work, explaining errors, and providing guidance.
