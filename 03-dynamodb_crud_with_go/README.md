# 03 — DynamoDB CRUD with Go

In this tutorial I am learning how to talk to DynamoDB from Go — designing a table with Terraform, then writing a Go data access layer to create, read, update, delete and query items. No Lambda here yet; it is just a local CLI hitting the real table (I may wrap it in a Lambda in a later tutorial).

Tutorial: [Use DynamoDB with Go](https://oneuptime.com/blog/post/2026-02-12-use-dynamodb-with-go/view)

## What I am building

```text
Go CLI → AWS SDK v2 → DynamoDB table (learning-Users)
```

A `learning-Users` table with a `user_id` partition key and a `status-index` GSI for querying by status, provisioned with Terraform. The Go code uses `attributevalue` for struct marshalling and expression builders for queries, following a repository-pattern approach.

## Project structure

- `golang/` — Go source (CLI in `cmd/app`, repository + client in `internal/`)
- `terraform/` — DynamoDB table infrastructure
- `How to Use DynamoDB with Go.pdf` — offline copy of a related reference doc

## What I want to learn

- How to design a DynamoDB table (partition key vs. GSI) with Terraform
- How Go's strong typing maps onto DynamoDB's schemaless items via struct tags
- CRUD operations and conditional writes with the AWS SDK v2 expression builder
- Querying a GSI vs. a full table Scan
