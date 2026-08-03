# Next steps

What I have built so far around tutorial 05, and what is left.

## Done

**CI on every push** — `.github/workflows/05-multi_lambda_crud_with_api_gateway.yaml`

- Go: `go vet`, `gofmt` check, `go test -race`, `./build.sh`
- Terraform: `fmt -check`, `init -backend=false`, `validate`

None of that needs AWS credentials: the Go tests use a fake DynamoDB client, and
`validate` only needs the providers downloaded.

**Remote state in S3** — `terraform/providers.tf`

State for 05 lives in `aws-hands-on-tfstate-725211237961` instead of on my
laptop. `use_lockfile = true` stops two runs applying at once, so no separate
DynamoDB lock table is needed. Moving existing state across was
`terraform init -migrate-state`.

**OIDC** — `00-setup/github-oidc/`

GitHub Actions assumes an IAM role directly, so there are no AWS keys sitting in
repo secrets. One identity provider for the whole account, plus a role whose
trust policy is scoped to this repo.

**terraform plan in CI**

Runs against the real S3 state with the OIDC role, so I can see what an apply
would change. The role is read-only, so plan is as far as it can go.

`needs: [build, terraform]` gates it behind the two free jobs, so a broken build
never spends an AWS round-trip.

**One workflow file per tutorial**

Rather than one clever file for all of them. The tutorials are not uniform — 01
has no Terraform, 03 has no `build.sh` — so a shared workflow would need
conditionals. A `paths:` filter keeps each workflow to its own folder.

Two things I learnt doing it: workflow files must live in `.github/workflows/`
at the repo root, so a tutorial folder is not quite as self-contained as the
rest of the repo; and `working-directory` is what points them back at the right
place.

## To do

### 1. Move the other tutorials to S3 state

Only 05 is on the S3 backend. 02, 03, 04 and both `00-setup/` folders still keep
state on disk. Same recipe as 05: add the `backend "s3"` block, then
`terraform init -migrate-state`.

### 2. apply on merge to main

#### The problem it solves

Terraform **zips** my Go binaries, it never **compiles** them. `lambda.tf` points
`archive_file` at `golang/build/<name>/bootstrap`, and that file only exists
because `build.sh` put it there.

So deploying is two commands, in this order:

```bash
./build.sh          # compile
terraform apply     # zip and upload
```

Forget the first one and this happens:

1. I edit a handler
2. I run `terraform apply`
3. Terraform zips the *old* `bootstrap` - unchanged
4. `source_code_hash` is identical, so Terraform reports **"no changes"**
5. My edit never reaches AWS, and the stale binary keeps serving

That is exactly what bit me: apply said everything was fine while the deployed
code was querying a table that no longer existed.

A workflow cannot forget a step. If CI runs both commands on every merge, they
cannot drift apart. The `terraform-plan` job already has this shape - it builds
the binaries before Terraform runs, because a fresh runner has no `build/`
folder.

#### What has to happen first

**The role cannot write.** `aws-hands-on-github-actions` has `ReadOnlyAccess`,
which is enough for `plan` (it only reads) but not for `apply`, which creates and
changes Lambdas, IAM, DynamoDB and API Gateway.

**Then: who is allowed to use it.** The trust policy currently accepts
`repo:foxcodenine/aws-hands-on:*` - any branch. That is fine while the role is
read-only, but a role that can `apply`, reachable from any branch, means an
unfinished experiment can change real infrastructure.

Cleaner than tightening the existing role is to have two:

| role  | permissions | who may use it                        |
| ----- | ----------- | ------------------------------------- |
| plan  | read-only   | any branch, so PRs can show a diff    |
| apply | write       | `main` only                           |

That keeps the permissive trust and the dangerous permissions off the same role.

### 3. CloudWatch alarm on the error rate

Get told when something breaks, instead of finding out by calling the API.
Lambda publishes an `Errors` metric on its own — no code needed — so this is an
`aws_cloudwatch_metric_alarm` plus an SNS topic with my email on it.

Worth doing now that the handlers split `reject` (4xx, logged WARN) from `fail`
(5xx, logged ERROR). An alarm is only useful if ERROR means something is really
wrong — if bad user input counted, it would fire constantly and I would learn to
ignore it.

### 4. Log group retention

The six Lambdas have no `aws_cloudwatch_log_group` in Terraform, so AWS creates
them automatically with **never expire**, and `terraform destroy` leaves them
behind. Declaring them with `retention_in_days` is a few lines.

### 5. X-Ray tracing

Shows where the *time* goes, rather than why something failed:

```
API Gateway    5ms
  Lambda     420ms   (cold start 380ms)
    DynamoDB  28ms
```

Needs `tracing_config { mode = "Active" }` on the Lambda plus
`xray:PutTraceSegments` in the IAM policy. Low value at this size — one Lambda
and one table makes a shallow trace — but worth seeing once.

## Later, not now

**Modules.** Extracting the repeated "Lambda + IAM role + permission" pattern
into a reusable module. Commonly asked about in interviews, and the same
abstraction instinct as any other code. Parking it until the above is done.
