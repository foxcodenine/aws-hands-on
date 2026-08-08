# AGENTS.md

## What this repo is

Hands-on AWS learning — one numbered tutorial folder per topic (`01-`, `02-`, …). Each folder is self-contained with its own source code, deploy scripts, and notes.

## Conventions

- Folders are numbered in order, `01-` to `05-` so far; the next one is `06-`.
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
- All state lives in S3 (`aws-hands-on-tfstate-<account-id>`, `use_lockfile = true`) — every tutorial and both `00-setup/` folders. `00-setup/remote-state-aws-buket` creates that bucket and keeps its own state in it, which only matters when bootstrapping a fresh account or tearing everything down; the ordering for both is commented in its `provider.tf`.
- The state bucket is protected two ways: `prevent_destroy` in the Terraform config, and a bucket policy denying `s3:DeleteBucket`. Both have to be removed by hand before it can ever be deleted.
- 05's `terraform/providers.tf` has `profile` commented out on purpose: CI runners have no named profile. Do not uncomment it.
- **Terraform zips Go binaries but never compiles them.** Deploying by hand is always `./build.sh` then `terraform apply`, in that order. Skip the build and Terraform re-zips the old binary, sees an unchanged `source_code_hash`, reports "No changes" and ships nothing — a deploy that looks successful and is not.
- Invoking Lambda via CLI requires `--cli-binary-format raw-in-base64-out` for raw JSON payloads.
- IAM execution role: `lambda-basic-execution` with `AWSLambdaBasicExecutionRole` attached.

## CI/CD

- Workflows live in `.github/workflows/`, one file per tutorial, each with a `paths:` filter so it only runs for its own folder.
- **Deploying tutorial 05 is manual** — the `workflow_dispatch` button on `main`. A push runs build, validate and plan only, and never changes AWS. The apply job builds the Go binaries and runs `terraform apply -auto-approve` in the same job — that pairing is deliberate, since Terraform zips the binaries but never compiles them.
- Auth is OIDC — no AWS keys in the repo. Two roles, both in `00-setup/github-oidc/`: `aws-hands-on-github-actions` (read-only, any branch, used by plan) and `aws-hands-on-github-actions-apply` (write, `refs/heads/main` only).
- The pipeline is documented in `05-multi_lambda_crud_with_api_gateway/README.md` — the four jobs, the IAM roles, and the supporting setup. OIDC itself is explained in `00-setup/github-oidc/README.md`.
- What is still planned: `05-multi_lambda_crud_with_api_gateway/docs/next-steps.md`. That file is to-do only; finished work moves into the README.

## Git

- **Never run `git commit`, `git push`, or create branches.** I do all of that
  myself. Propose the commits instead: a one-line message plus the files that
  belong in each, and I will run them.
- Give the message as **bare text**, not wrapped in a `git commit -m "…"`
  command. I paste it into my own `-m '…'`, so a full command ends up nested
  inside the quotes and becomes the commit message.

## Preferences

- Docs written from a first-person learner perspective (what I built, what I learnt, what tripped me up).
- Keep READMEs short — this is a learning repo, not a product.
- No unnecessary comments in code; only add one when the why is non-obvious.
- Do not write tutorial implementation code unless explicitly asked. Prefer reviewing the learner's work, explaining errors, and providing guidance.
