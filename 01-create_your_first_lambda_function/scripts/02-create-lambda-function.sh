#!/usr/bin/env bash
set -euo pipefail

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip myLambdaFunctionGo.zip bootstrap

ROLE_ARN=$(aws iam get-role --role-name lambda-basic-execution --query 'Role.Arn' --output text --profile developer)

aws lambda create-function \
  --function-name myLambdaFunctionGo \
  --runtime provided.al2023 \
  --handler bootstrap \
  --architectures x86_64 \
  --role "$ROLE_ARN" \
  --zip-file fileb://myLambdaFunctionGo.zip \
  --profile developer
