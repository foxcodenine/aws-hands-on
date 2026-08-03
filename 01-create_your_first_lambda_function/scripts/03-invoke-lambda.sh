#!/usr/bin/env bash
set -euo pipefail

aws lambda invoke \
  --function-name myLambdaFunctionGo \
  --payload '{"length": 5, "width": 3}' \
  --cli-binary-format raw-in-base64-out \
  --profile developer \
  output.json

cat output.json

rm output.json
