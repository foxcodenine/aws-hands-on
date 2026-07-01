# 01 — Create Your First Lambda Function

A simple rectangle area calculator (`length × width`) deployed as a Lambda function. The same logic is implemented in three runtimes to compare how each one works.

## What I built

The function takes a JSON payload like `{"length": 5, "width": 3}`, calculates the area, and returns `{"area": 15}`. Nothing clever — the point was to learn the deployment pipeline, not the business logic.

## Runtimes covered

| Folder | Runtime | Notes |
|--------|---------|-------|
| `python/` | Python 3.x | Managed runtime, quickest to deploy |
| `node/` | Node.js | Managed runtime, ES module syntax (`export`) |
| `golang/` | Go (`provided.al2023`) | Custom runtime — required more setup |

## What I learnt

**IAM roles are not optional.** Before a Lambda can run, it needs an execution role with at least `AWSLambdaBasicExecutionRole` attached. AWS handles this silently when you use the console, but doing it via the CLI made it explicit: create the role, write the trust policy, attach the policy, grab the ARN, pass it to `create-function`.

**Go is a different thing.** Python and Node upload a source file and AWS handles the rest. Go requires compiling a static binary (`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`), naming it `bootstrap`, zipping it, and choosing the `provided.al2023` runtime. The handler name in the console must also be set to `bootstrap`.

**The context object differs by runtime.** In Python it's `context.log_group_name`, in Node it's `context.logGroupName`, and in Go the SDK exposes it differently via `lambdacontext.FromContext(ctx)` — and the log group name isn't available at all, only the function ARN.

**CLI invocation needs a flag.** Passing a JSON payload via `aws lambda invoke` requires `--cli-binary-format raw-in-base64-out`, otherwise the CLI expects a base64-encoded payload by default.

## Steps (Go / CLI path)

The `golang/steps/` folder has the full CLI walkthrough:

1. `01-deploy-lambda-from-aws-dashboard.md` — console deployment guide
2. `02-create_a_new_user.md` — IAM user + access key setup
3. `03-create_execution_role_for_lambda.sh` — create the Lambda execution role
4. `04-create-the-actual-lambda-function.sh` — build, zip, and deploy
5. `05-invoke-lambda-via-cli.sh` — invoke and see the output
