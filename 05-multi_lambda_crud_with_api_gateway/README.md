# 05 — Multi-Lambda CRUD with API Gateway

Building on 03 (DynamoDB CRUD in Go) and 04 (API Gateway in front of one monolithic `operation`-dispatch Lambda), I am rebuilding the same CRUD API as several single-purpose Lambda functions instead of one function doing everything. Each Lambda handles exactly one operation, wired to its own API Gateway route.

## What I am building

```text
API Gateway (one route per operation)
  ├─ POST   /users        → create_user Lambda  ┐
  ├─ GET    /users        → list_users Lambda    │
  ├─ GET    /users/{id}   → get_user Lambda      ├─→ DynamoDB table
  ├─ PUT    /users/{id}   → update_user Lambda   │
  ├─ DELETE /users/{id}   → delete_user Lambda   │
  └─ POST   /echo         → echo Lambda          ┘
```

## Project structure

- `golang/cmd/` — one Lambda entrypoint per operation, each built and deployed as its own binary
- `golang/internal/` — shared `api`, `db`, `models` and `repository` packages, reused across all six Lambdas
- `golang/build.sh` — compiles every folder in `cmd/` into `build/<name>/bootstrap`
- `golang/http/` — `.http` requests to exercise the deployed API by hand
- `terraform/` — DynamoDB table, IAM, the Lambdas, and the API Gateway wiring
- `docs/` — my notes:
  - [`api-gateway.md`](./docs/api-gateway.md) — how API Gateway fits together
  - [`go-setup.md`](./docs/go-setup.md) — the commands that set the Go module up
  - [`next-steps.md`](./docs/next-steps.md) — what is left to do

## Deploy

When pushed to GitHub, nothing deploys on its own.
A push only builds, tests and plans — it never touches AWS. To deploy, I press a button.

### From GitHub

The normal route to deploy is from the **Run workflow** button on `main`, in the Actions tab.

One job compiles the binaries and then applies, so it cannot forget the build
step.

Pushing code and deploying infrastructure are two separate decisions. A
documentation change or an unfinished commit cannot modify AWS on its own.

### By hand

First compile the Go lambdas.
Then run `terraform apply` to zip the Lambda binaries and apply everything to AWS.

```bash
cd golang && ./build.sh
cd ../terraform && terraform apply
```

If `build.sh` is skipped, Terraform zips the old binary, sees nothing new, and prints
`No changes`. The deploy looks fine and ships nothing.

The base URL for `golang/http/users.http`:

```bash
export AWS_PROFILE=developer   # the profile is commented out for CI
terraform output -raw invoke_url
```

## Test the Go code

Every test uses a fake DynamoDB client, so they run offline and never touch AWS:

```bash
cd golang
go test ./...
go vet ./...
```

Handy variants:

```bash
go test ./... -v                 # each test name, with its own timing
go test ./internal/api/ -v       # one package
go test ./... -run TestList -v   # one test
go test ./... -cover             # coverage per package
go test ./... -count=1           # ignore the cache and really re-run
go test ./... -race              # detect data races
go test ./... -failfast          # stop at the first failure
```

Timing comes for free: the `ok  <package>  0.004s` line is the total, and `-v`
adds a duration per test. Anything slow enough to notice here is a mistake —
these never leave the machine.

`(cached)` instead of a time means nothing changed since the last run, so Go
skipped it. `-count=1` forces a real run.

For coverage line by line, in a browser:

```bash
go test ./internal/api/ -coverprofile=cover.out && go tool cover -html=cover.out
```

The tests cover the parts most likely to break quietly: a missing item returning 404 rather than an empty 200, duplicate emails being rejected *before* the write, pagination cursors round-tripping, and `DELETE` answering 204 with no body. `cmd/` has no tests — those files are three-line `main()`s.

## The workflow

```text
.github/workflows/05-multi_lambda_crud_with_api_gateway.yaml
```

Four jobs. The first two need no AWS credentials at all.

### Go checks

```bash
go vet ./...
gofmt check
go test -race ./...
./build.sh
```

The tests use a fake DynamoDB client, so they run offline.

### Terraform checks

```bash
terraform fmt -check
terraform init -backend=false
terraform validate
```

`-backend=false` skips the state entirely. This only downloads the providers and
checks the configuration against them.

### Terraform plan

Runs against the real S3 state using the read-only IAM role, so this job does
need AWS credentials.

It waits for the two free jobs, so a broken build never spends an AWS
round-trip:

```yaml
needs:
  - go-build
  - terraform-validate
```

### Terraform apply

The only job that changes AWS. It uses the apply role, and runs from the **Run
workflow** button on `main`:

```yaml
if: github.ref == 'refs/heads/main' && github.event_name == 'workflow_dispatch'
```

The branch check matters because the button lets you pick any branch from a
dropdown. Only `main` may apply.

It compiles and applies in the same job:

```bash
./build.sh
terraform apply -auto-approve
```

Terraform zips binaries but never compiles them, so the two steps must stay
together. A job cannot forget the build.

It runs last:

```yaml
needs:
  - go-build
  - terraform-validate
  - terraform-plan
```

## IAM roles

The workflow uses two roles, never one.

| Role | Permissions | Allowed branches |
|---|---|---|
| Plan role | Read-only | Any branch |
| Apply role | Write access | `main` only |

### Plan role

Uses AWS `ReadOnlyAccess`. Terraform plan only needs to:

- Read the current AWS resources
- Read the remote Terraform state
- Compare the current infrastructure with the configuration

It does not need permission to create or update anything.

Its trust policy allows any branch in the repository:

```text
repo:foxcodenine/aws-hands-on:*
```

That is acceptable precisely because the role can only read.

### Apply role

Can create and update:

- Lambda functions
- DynamoDB tables
- API Gateway resources
- CloudWatch log resources

Its trust policy allows one branch only:

```text
refs/heads/main
```

Its permissions are also limited to resources whose names start with
`aws-hands-on-`. It is defined in `00-setup/github-oidc/iam-apply.tf`.

### Why two roles, not one

Separate roles keep broad branch access and write permissions apart.

Pull requests and experimental branches can inspect infrastructure changes, but
they cannot modify AWS.

## Supporting setup

### Remote Terraform state

Tutorial 05 stores its Terraform state in S3, not only on my laptop:

```text
aws-hands-on-tfstate-725211237961
```

The backend uses:

```hcl
use_lockfile = true
```

This prevents two Terraform operations from modifying the state at the same
time, so a separate DynamoDB locking table is not required.

The existing local state was moved with:

```bash
terraform init -migrate-state
```

### GitHub OIDC authentication

OIDC lets GitHub Actions assume AWS IAM roles without permanent access keys, so
there are no AWS credentials stored in repository secrets.

The infrastructure is in `00-setup/github-oidc/` and contains:

- One GitHub OIDC provider for the AWS account
- A read-only plan role
- A restricted apply role

### Workflow concurrency

The workflow has a top-level `concurrency` configuration, so two runs cannot race
each other. One waits for the other to finish.

### Go module caching

GitHub Actions originally displayed:

```text
Dependencies file is not found
```

The Go module is inside a tutorial subfolder rather than at the repository root.
`go-version-file` selects the Go version, but it does not tell `setup-go` where
the dependency file is. Caching needs its own input:

```yaml
go-version-file: 05-multi_lambda_crud_with_api_gateway/golang/go.mod
cache-dependency-path: 05-multi_lambda_crud_with_api_gateway/golang/go.sum
```

Added to every job that installs Go. This avoids re-downloading the AWS SDK on
every run.
