# 03 — DynamoDB CRUD with Go

In this tutorial I am learning how to talk to DynamoDB from Go — designing a table with Terraform, then building a small HTTP API to create, read, update, delete and query users. No Lambda here yet; the local Go server talks directly to DynamoDB.

Tutorial: [Use DynamoDB with Go](https://oneuptime.com/blog/post/2026-02-12-use-dynamodb-with-go/view)

## What I am building

```text
HTTP API → Go repository → AWS SDK v2 → DynamoDB table (learning-Users)
```

A `learning-Users` table with a `user_id` partition key, a `status-index` GSI for active-user queries, and an `email-index` GSI for email lookups. The Go code uses `attributevalue` for struct marshalling and expression builders for queries, following a repository-pattern approach.

## Run the API

First create the DynamoDB table from `terraform/`, then start the Go server:

```bash
cd terraform
terraform init
terraform apply

cd ../golang
go run ./cmd/app
```

The server listens on port `8080` by default. Set `PORT` to use another port.

## API endpoints

```text
POST   /users/                  Create a user
GET    /users/                  Return all users
GET    /users/active            Return active users
GET    /users/by-email?email=…  Find a user by email
GET    /users/{userID}          Return one user
PUT    /users/{userID}          Update a user
DELETE /users/{userID}          Delete a user
```

The create handler checks whether an email already exists and returns `409 Conflict` when it finds one. DynamoDB GSIs do not enforce uniqueness by themselves, so simultaneous requests can still require a transaction-based uniqueness lock in a future improvement.

## Project structure

- `golang/` — Go source (HTTP server in `cmd/app`, handlers, routers, repository and client in `internal/`)
- `terraform/` — DynamoDB table infrastructure
- `How to Use DynamoDB with Go.pdf` — offline copy of a related reference doc

## What I want to learn

- How to design a DynamoDB table (partition key vs. GSI) with Terraform
- How Go's strong typing maps onto DynamoDB's schemaless items via struct tags
- CRUD operations and conditional writes with the AWS SDK v2 expression builder
- Querying a GSI vs. a full table Scan

## Test

```bash
cd golang
go test ./...
go vet ./...
```
