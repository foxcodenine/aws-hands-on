# GitHub Actions — what I could use it for here

Notes for when I get to this, probably as tutorial 06. Nothing set up yet.

## Why bother, for this repo specifically

Three things from tutorial 05 that CI would have caught or clarified:

- **Unformatted code and HCL typos.** `terraform validate` alone would have caught
  the missing GSI `hash_key` in tutorial 03, which only showed up when I ran a
  plan by hand.
- **"Works on my machine".** My Go toolchain turned out to be 32-bit, which broke
  cgo locally. A GitHub runner is amd64 with a working toolchain, so CI would
  have been green while my laptop failed — that contrast is the fastest way to
  tell a real break from a local one.
- **Deploying stale code.** I edited Go, ran `terraform apply`, got "no changes",
  and the old binary stayed live querying a table that no longer existed.
  Terraform zips but does not compile. If build and apply are one CI step, they
  cannot drift apart.

## Tier 1 — no AWS credentials needed

Start here. All of it runs on a bare runner, because the Go tests use a fake
DynamoDB client and Terraform can validate without a backend.

- `go test ./...` and `go vet ./...`
- `gofmt -l .` — fail if anything is unformatted
- `terraform fmt -check`
- `terraform init -backend=false && terraform validate`
- `./build.sh` — proves all six Lambdas still compile for `linux/amd64`

Sketch (check action versions when I actually do this):

```yaml
name: ci
on: [push, pull_request]

jobs:
  go:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: 05-multi_lambda_crud_with_api_gateway/golang
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - run: go vet ./...
      - run: go test ./... -race
      - run: ./build.sh

  terraform:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: 05-multi_lambda_crud_with_api_gateway/terraform
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
      - run: terraform fmt -check
      - run: terraform init -backend=false
      - run: terraform validate
```

### One workflow per tutorial

Decided: a separate workflow file per tutorial, e.g. `ci-05.yml`, rather than one
clever file covering all of them. Simpler while I am learning, and it matches the
repo rule that each tutorial is self-contained. The tutorials are not uniform
anyway — 01 has no Terraform, 03 has no `build.sh`, 05 has six Lambdas — so a
shared workflow would need conditionals to paper over the differences.

Two things to know about doing it this way:

- Workflow files **must** live in `.github/workflows/` at the repo root. They
  cannot sit inside the tutorial folder, so the folder is not quite as
  self-contained as the rest of the repo. `working-directory` points them at the
  right place.
- Add a `paths:` filter so each workflow only runs when its own folder changes -
  otherwise editing tutorial 02 runs the tests for all of them:

```yaml
on:
  push:
    paths:
      - '05-multi_lambda_crud_with_api_gateway/**'
```

If the copies ever get annoying, the alternative is a *matrix*: list the folder
names once and GitHub runs the same steps for each, the same idea as `for_each`
in Terraform. Worth revisiting only if there end up being a lot of them.

## Tier 2 — needs AWS credentials

`terraform plan` on pull requests, with the plan posted as a PR comment so I can
see what an apply *would* change before merging.

The thing to learn properly here is **OIDC**: GitHub can assume an IAM role
directly via `aws-actions/configure-aws-credentials`, so there are no long-lived
access keys sitting in repo secrets. It needs an IAM OIDC identity provider for
`token.actions.githubusercontent.com` plus a role whose trust policy is scoped to
this repo.

That IAM setup is itself a good Terraform exercise — and a natural follow-on from
the identity-based vs resource-based permissions work in 02 and 05.

## Tier 3 — deploy on merge

`./build.sh` then `terraform apply` when something lands on `main`. This is the
one that removes the build/apply drift problem for good.

### Blocker: my state is local

Every tutorial keeps `terraform.tfstate` on disk, gitignored. A runner starts
empty, sees no state, and would try to create everything from scratch — or worse,
fail halfway on names that already exist.

So tier 3 needs, first:

- an **S3 backend** for state. I already have `terraform-state-s3-bucket-cf12` in
  the account from earlier work, so this is not new ground.
- **locking**, so two runs cannot apply at once. Recent AWS provider versions
  support S3-native locking, so a separate DynamoDB lock table may no longer be
  required — check what the provider wants at the time.

Migrating an existing local state to a backend is `terraform init -migrate-state`.
Worth doing on one tutorial first rather than all five at once.

## Suggested order

1. Tier 1 workflow — no prerequisites, useful immediately
2. S3 backend + locking on one tutorial, then the rest
3. OIDC role in Terraform
4. `plan` on PRs
5. `apply` on merge to `main`
