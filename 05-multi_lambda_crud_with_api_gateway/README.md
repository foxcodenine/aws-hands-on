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
- `steps/` — my walkthrough and learning notes

## Deploy

Terraform zips the binaries but does **not** compile Go, so always build first:

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

- **A 403 `Missing Authentication Token` means the route does not exist**, not that auth is needed. I had called `/test/test` — the `/test` in the URL is the *stage* name, not a path.
- **`terraform apply` does not compile Go.** I edited the Go source, applied, and got "no changes" — the old binary stayed live and kept querying a table that no longer existed. `source_code_hash` only notices when the zip changes, and the zip only changes after `build.sh` runs.
- **The handlers swallow the underlying error.** A `500 {"error":"failed to check email"}` left nothing in CloudWatch but `START`/`END`, so the cause had to be found from the outside. Logging `err` before returning the generic message would have named it immediately.
- **`PUT` on a missing user returns 500, not 404.** The repository's `attribute_exists` guard makes DynamoDB reject the write as an *error*, so the handler's `user == nil` branch never runs. Still to fix.
- In a `.http` file, a request body runs until the next `###` — a `@variable` defined underneath one gets swallowed into the body and breaks the JSON.
