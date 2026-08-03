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
  - [`next-steps.md`](./docs/next-steps.md) — what is built so far, and what is left

## Deploy

Pushing to `main` deploys. The workflow builds the binaries and runs
`terraform apply` in the same job, so the zip can never contain stale code.

To apply by hand — Terraform zips the binaries but does **not** compile Go, so
always build first:

```bash
cd golang && ./build.sh
cd ../terraform && terraform apply
```

`terraform output invoke_url` gives the base URL for `golang/http/users.http`.

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

## What I want to learn

- How to structure a Go repo with multiple Lambda entrypoints sharing common internal packages
- How Terraform scales once there are several Lambda functions instead of one — repeated resource blocks vs. `for_each`
- How API Gateway maps several routes to several different Lambda integrations
- Trade-offs between one monolithic Lambda (04) and many single-purpose Lambdas (05) — cold starts, deployment size, IAM blast radius, and how much shared code actually makes sense to factor out

## What tripped me up

### `403 Missing Authentication Token` means the route does not exist

Nothing to do with authentication, even though every method here uses
`authorization = "NONE"`. API Gateway returns this for *any* path it cannot
match. I had requested `/test/test`, because I forgot the `/test` already in the
base URL is the **stage name**, not part of the path.

Rule of thumb: 403 means the route is wrong, 500 means the route was found and
something behind it failed.

### `terraform apply` does not compile Go

I changed the Go source, ran `apply`, and got "no changes". The old binary stayed
live, still querying a table that no longer existed.

Terraform only zips what is already on disk. `source_code_hash` notices when the
*zip* changes, and the zip only changes once `build.sh` has run. So the order is
always `./build.sh` first, `terraform apply` second.

### The handlers throw away the real error

A `500 {"error":"failed to check email"}` left nothing in CloudWatch but
`START`/`END`/`REPORT`, so the cause had to be worked out from the outside.

The handler returns a generic message — which is right, callers should not see
internals — but it never logs `err` first, so the detail is lost entirely. One
`log.Printf` before each error response would have named the problem straight
away.

### `PUT` on a missing user returns 500, not 404

The repository guards `UpdateItem` with `attribute_exists(user_id)`. When the
user does not exist, DynamoDB rejects the write with
`ConditionalCheckFailedException` — an **error**, not an empty result. So the
handler takes its `err != nil` branch and returns 500, and the `user == nil`
branch that would return 404 is unreachable.

`DELETE` does not have this problem: it genuinely returns `(nil, nil)` when there
was nothing to delete. Still to fix.

### In a `.http` file, the body runs until the next `###`

I put a variable directly under a request and the request stopped working. The
reason is that everything after the blank line, all the way to the next `###`,
counts as the body:

```
POST {{baseUrl}}/users
Content-Type: application/json
                              <- blank line: headers end, body starts
{
  "name": "Ada Lovelace"
}
@userID = abc-123             <- still the body, not a variable
                              
### Next request              <- only here does the body end
```

So the Lambda received the JSON *plus* `@userID = abc-123` stuck on the end,
which is not valid JSON — hence `400 invalid JSON`. Variables have to go at the
top of the file, above every request.
