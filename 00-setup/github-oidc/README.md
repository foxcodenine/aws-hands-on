# GitHub OIDC — letting GitHub Actions into AWS without keys

**OIDC** is short for **OpenID Connect**. It is a standard way for one service to
prove who it is to another, without the two of them sharing a password.

## The problem it solves

My workflow runs `terraform plan` and `terraform apply`, so it needs AWS
credentials.

The obvious way is to create an IAM user, generate an access key, and paste the
key and secret into GitHub repository secrets.

That works, but the credentials:

- never expire, so they are valid until I remember to delete them
- work from anywhere — if they leak, anyone can use them from their laptop
- sit in a settings page indefinitely, and I have no easy way to tell whether
  anyone has copied them

It is a house key hidden under the mat. OIDC replaces it with a visitor pass
that expires.

## How it works instead

Nothing is stored. GitHub proves who it is each time, and AWS hands back
credentials that expire in about an hour.

```text
A workflow job starts
   ↓
GitHub creates a short-lived signed token describing the job:
   "this is repo foxcodenine/aws-hands-on, branch main"
   ↓
The job sends that token to AWS
   ↓
AWS checks two things:
   1. was it really signed by GitHub?
   2. do the details match what I said I would accept?
   ↓
AWS returns temporary credentials
```

The token is signed by GitHub, so it cannot be forged, and it describes the exact
job it was made for. AWS never sees a password, and there is nothing sitting in
my repository to steal.

## What I built

Three things, in `terraform/`:

**1. The identity provider** — `main.tf`

```hcl
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
}
```

This is the introduction: it tells AWS that tokens signed by GitHub are worth
looking at. On its own it grants no permissions at all. One per AWS account, and
both roles below share it.

**2. and 3. Two IAM roles**

| role | file | permissions | who may use it |
| ---- | ---- | ----------- | -------------- |
| `aws-hands-on-github-actions` | `main.tf` | read-only | any branch |
| `aws-hands-on-github-actions-apply` | `iam-apply.tf` | write | `main` only |

The role decides everything that matters: which repository may connect, and what
it is allowed to do once it has.

Splitting them keeps the loose trust and the dangerous permissions apart. The
read-only role can be used from any branch, because reading is harmless. The one
that can change infrastructure is reachable only from `main`. The full reasoning
is commented in `iam-apply.tf`.

## The two checks that do the real work

Both roles' trust policies check two claims inside the token.

**`aud`** — who the token was made for:

```text
sts.amazonaws.com
```

GitHub can mint tokens for many services. This confirms the token was meant for
AWS, and not something else that might replay it.

**`sub`** — which job it came from:

```text
repo:foxcodenine/aws-hands-on:*                  # plan role, any branch
repo:foxcodenine/aws-hands-on:ref:refs/heads/main # apply role, main only
```

Note the different matchers. The plan role uses `StringLike`, because the `*` is
a wildcard that has to match anything. The apply role uses `StringEquals`, since
there is no wildcard left — one exact string, and nothing else gets in.

Without this second check, any repository on GitHub could assume my roles.

## How a workflow uses it

Two parts in `.github/workflows/`:

```yaml
permissions:
  id-token: write        # lets the job ask GitHub for a token

steps:
  - uses: aws-actions/configure-aws-credentials@v6
    with:
      role-to-assume: arn:aws:iam::725211237961:role/aws-hands-on-github-actions
      aws-region: eu-west-1
```

`id-token: write` sounds like it grants something dangerous. It does not — it
only allows the job to request its own identity token. Without it, GitHub refuses
to issue one and the step fails.

## Why it was worth doing

- **No AWS keys anywhere in the repo or its settings.** Nothing to leak.
- **Nothing to rotate.** There is no long-lived secret to remember to replace.
- **Credentials expire on their own,** usually within the hour.
- **Scoped to one repository,** and for `apply`, to one branch. A stolen token
  from somewhere else is useless.

## Files

```text
terraform/
  main.tf        identity provider + the read-only plan role
  iam-apply.tf   the write role, main branch only
  data.tf        current AWS account id
  variables.tf   github_owner, github_repo
  outputs.tf     the role ARNs, for pasting into a workflow
```
