# 03 — Connect a Lambda Function to DynamoDB

In this tutorial I am learning how to connect a Go Lambda function to DynamoDB — designing a table with Terraform, then writing a Go data access layer to read and write items from it.

Tutorial: [Use DynamoDB with Go](https://oneuptime.com/blog/post/2026-02-12-use-dynamodb-with-go/view)

## What I am building

```text
Go Lambda function → AWS SDK v2 → DynamoDB table (Users)
```

A `Users` table with a `user_id` partition key and a `status-index` GSI for querying by status, provisioned with Terraform. The Go code will use `attributevalue` for struct marshalling and expression builders for queries, following the repository-pattern approach from the tutorial.

## Project structure

- `golang/` — Go Lambda source code
- `terraform/` — DynamoDB table and IAM infrastructure
- `How to Use DynamoDB with Go.pdf` — offline copy of a related reference doc

## What I want to learn

- How to design a DynamoDB table (partition key vs. GSI) with Terraform
- How Go's strong typing maps onto DynamoDB's schemaless items via struct tags
- CRUD operations and conditional writes with the AWS SDK v2 expression builder
- Querying a GSI vs. a full table Scan
