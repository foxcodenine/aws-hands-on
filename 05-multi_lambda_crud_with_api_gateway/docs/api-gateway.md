# Understanding `terraform/api_gateway.tf`

Notes to myself, working through the file slowly. In tutorial 04 the whole API
was one path and one method, so this was easy to skim past. With six routes it
stopped being obvious, so this is me writing out what each piece actually does.

## The one idea to hold on to

API Gateway does not have a "route" resource. There is no single block where you
say *"`GET /users/{userID}` is handled by the `get_user` Lambda"*.

Instead one route is assembled from separate pieces:

| Piece | What it does |
|---|---|
| `aws_api_gateway_resource` | Creates **one segment of the URL**, like `users`. Segments are built up one at a time, each pointing at its parent, so `/users/{userID}` takes two of them. This only creates the address — nothing answers there yet. |
| `aws_api_gateway_method` | Declares **which HTTP verb that address accepts**, and whether the caller needs to authenticate. A path with no method still exists, but rejects every request that reaches it. |
| `aws_api_gateway_integration` | Points that address-and-verb at **whatever actually does the work** — for me, one specific Lambda. This is the real wiring: the bit that says `GET /users/{userID}` means *run `get_user`*. |
| `aws_lambda_permission` | Sits on the **Lambda side, not the API side**, and lets the API Gateway service invoke that function. The wiring above can be perfect and the call still gets refused without this. |

Miss any one of them and the route fails in a different way. That is really the
whole file — everything else is avoiding writing those four things out six times.

It helps to compare with tutorial 03, where all four were a single line of Chi:

```go
r.Get("/users/{userID}", userHandler.Show)
```

That one line said the path exists, that it answers `GET`, and which function
handles it — and no permission was needed at all, because the router and the
handler were the same running program. Here the handler lives in a *separate
service*, so API Gateway has to be told the address, the verb, how to reach the
code, and be granted permission to call it. Four resources instead of one line is
the price of that split.

## Part 1: the API container

```hcl
resource "aws_api_gateway_rest_api" "api" {
  name = "${var.prefix}${var.lesson}-api"
}
```

This creates the API and its hostname, and nothing else:

```
https://5rsbluc2b5.execute-api.eu-west-1.amazonaws.com
```

The address is real but every request returns `403` — there are no paths on it
yet. Everything else in this file attaches to this one API, which is why
`rest_api_id = aws_api_gateway_rest_api.api.id` keeps reappearing.

The one thing to remember: the API comes with a resource for `/` that I never
declare. Top-level paths use it as their parent:

```hcl
parent_id = aws_api_gateway_rest_api.api.root_resource_id   # "my parent is /"
```

## Part 2: paths are a tree, not strings

I expected to write `"/users/{userID}"` as one string. Instead every segment is
its own resource, and each one names its parent:

```
/  (comes free with the API)
├── users            <- aws_api_gateway_resource.users
│   └── {userID}     <- aws_api_gateway_resource.user_id
└── echo             <- aws_api_gateway_resource.echo
```

So `/users/{userID}` is two blocks, not one:

```hcl
resource "aws_api_gateway_resource" "users" {
  parent_id = aws_api_gateway_rest_api.api.root_resource_id   # parent is /
  path_part = "users"
}

resource "aws_api_gateway_resource" "user_id" {
  parent_id = aws_api_gateway_resource.users.id               # parent is /users
  path_part = "{userID}"
}
```

The braces make it a **path parameter**: `/users/abc-123` matches, and
`abc-123` is captured under the name inside them.

That name has to match the Go side exactly:

```
path_part = "{userID}"     ->     req.PathParameters["userID"]
```

Get it wrong and nothing errors — the handler just reads an empty string and
answers `400` on every request.

## Part 3: the route table

The path from Part 2 does nothing on its own — it exists, but rejects every
request. Turning it into a working route takes three more blocks, and I need six
routes. Rather than write eighteen blocks, the file lists the routes as **data**
first:

```hcl
locals {
  routes = {
    create_user = { path = aws_api_gateway_resource.users.id,   method = "POST" }
    list_users  = { path = aws_api_gateway_resource.users.id,   method = "GET" }
    get_user    = { path = aws_api_gateway_resource.user_id.id, method = "GET" }
    update_user = { path = aws_api_gateway_resource.user_id.id, method = "PUT" }
    delete_user = { path = aws_api_gateway_resource.user_id.id, method = "DELETE" }
    echo        = { path = aws_api_gateway_resource.echo.id,    method = "POST" }
  }
}
```

This creates nothing. It is just a table:

- the **key** is a Lambda name — and it must match a folder in `golang/cmd/`
- the **value** says which path that Lambda sits on, and which verb it answers

Reading the `get_user` line: *the `get_user` Lambda answers `GET` on the
`/users/{userID}` path.*

## Part 4: the three blocks

Each of the three blocks below has `for_each = local.routes`, which means
**Terraform runs it once per row of that table**. Inside the block:

- `each.key` is the row's name — `"get_user"`
- `each.value` is the row's data — `{ path = ..., method = "GET" }`

So one block written once produces six real resources. To make that concrete, I
show each block as written, then what it becomes for the `get_user` row.

### Block 1 — the method: which verb the path answers

```hcl
resource "aws_api_gateway_method" "routes" {
  for_each = local.routes

  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = each.value.path
  http_method   = each.value.method
  authorization = "NONE"
}
```

For the `get_user` row, `each.value` fills in to give:

```hcl
  resource_id   = <the /users/{userID} path>
  http_method   = "GET"
  authorization = "NONE"      # no login required
```

`/users/{userID}` now accepts `GET`. It still has no idea what should handle it.

### Block 2 — the integration: which Lambda runs

```hcl
resource "aws_api_gateway_integration" "routes" {
  for_each = local.routes

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = each.value.path
  http_method = each.value.method

  type = "AWS_PROXY"
  uri  = aws_lambda_function.functions[each.key].invoke_arn

  integration_http_method = "POST"

  depends_on = [aws_api_gateway_method.routes]
}
```

For the `get_user` row:

```hcl
  resource_id = <the /users/{userID} path>     # which route
  http_method = "GET"                          #   this is for
  uri         = <arn of the get_user Lambda>   # what runs
```

`each.key` is doing the important work in `uri`: it picks the matching Lambda out
of the six built in `lambda.tf`. The route table keys and the `cmd/` folder names
line up on purpose — that is what lets one line find the right function.

Three details in this block are worth knowing:

**`depends_on` is not optional here.** Terraform decides build order by looking
at what references what — and nothing in this block mentions the method, since
`http_method` reads its value from the route table instead. Left alone, Terraform
would think the two are unrelated and create them in parallel, which AWS rejects:
an integration cannot attach to a method that does not exist yet. Because it is a
race, it might pass one run and fail the next.

An alternative is to drop `depends_on` and write
`http_method = aws_api_gateway_method.routes[each.key].http_method`, which
produces the same string but states the dependency by mentioning the method. Both
work; `depends_on` just says out loud what the other version implies.

**`AWS_PROXY`** means give the function the whole request and let it decide the
whole response — which is why the handler takes an `APIGatewayProxyRequest` and
returns an `APIGatewayProxyResponse`, status and headers and body all from Go.

**`integration_http_method` is always `POST`**, and this is the trap, because now
two verbs sit side by side meaning different things:

```
http_method             = "GET"    <- the verb my API answers
integration_http_method = "POST"   <- the verb used to call Lambda
```

Invoking a Lambda is always a `POST` to the AWS API, whatever the caller sent. In
04 I set this from the method's own verb and it worked — but only because that
route happened to be POST. Here it would break `GET`, `PUT` and `DELETE`.

### Block 3 — the permission: letting API Gateway call the function

```hcl
resource "aws_lambda_permission" "routes" {
  for_each = local.routes

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.functions[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}
```

For the `get_user` row this grants: *the API Gateway service may invoke the
`get_user` function*.

This block lives on the **Lambda** side, and is the mirror image of the role in
`iam.tf`:

- the **role** says what the function may *do* — read DynamoDB, write logs
- the **permission** says who may *call* the function — API Gateway

Miss the role and the function runs but gets `AccessDenied` from DynamoDB. Miss
the permission and it never runs at all: a `500` with completely empty logs,
which looks like broken code but is not.

### What actually gets created

Three blocks, six rows, eighteen resources. Terraform keeps them keyed by row
name, which is how they appear in `terraform plan`:

```
aws_api_gateway_method.routes["get_user"]
aws_api_gateway_integration.routes["get_user"]
aws_lambda_permission.routes["get_user"]
... and the same three for the other five rows
```

Adding a seventh route means adding a `cmd/` folder, a line in `local.functions`,
and a line in `local.routes`. No new blocks.

## Part 5: deployment and stage

Everything so far is only a **draft**. The routes exist, but the URL still
returns `403`, because API Gateway keeps two things apart:

- the route configuration — the draft
- the published copy — what the URL actually serves

In tutorial 04 I crossed that gap by clicking **"Deploy API"** in the console.
These two blocks are that button.

**The deployment** publishes a copy of the routes as they are right now:

```hcl
resource "aws_api_gateway_deployment" "api" { ... }
```

**The stage** gives that copy a name, and the name becomes part of the URL:

```hcl
resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.api.id
  stage_name    = "test"
}
```

That `test` is where the `/test` in the URL comes from:

```
https://5rsbluc2b5.execute-api.eu-west-1.amazonaws.com/test/users
                                                      ^^^^
```

Later you could add a `prod` stage pointing at a different copy, so `/test` and
`/prod` serve different versions of the same API.

### The one gotcha: `triggers`

Terraform only rebuilds a resource when one of its arguments changes. Look at the
deployment block — its only real argument is `rest_api_id`, and that never
changes. So if I add a route:

1. Terraform creates the new method and integration ✓
2. it looks at the deployment block, sees it is identical to last time, skips it
3. nothing ever pressed "Deploy", so the URL still serves the old routes

`apply` reports success and nothing actually changed. `triggers` is the
workaround — an argument that exists *only* to change:

```hcl
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_method.routes,
      aws_api_gateway_integration.routes,
    ]))
  }
```

`sha1(jsonencode(...))` just squashes all the methods and integrations into one
short string. Edit any route and that string changes, which makes Terraform
create a new deployment.

`create_before_destroy` then builds the new copy before removing the old one, so
the stage is never pointing at nothing.

## Following one request all the way through

`GET /users/abc-123` on the deployed API:

1. API Gateway matches the host to the **REST API**, and `/test` to the **stage**.
2. `/users/abc-123` walks the **resource** tree: `users`, then `{userID}`,
   capturing `userID = "abc-123"`.
3. `GET` on that resource finds the **method**.
4. Its **integration** says `AWS_PROXY` to the `get_user` Lambda.
5. API Gateway checks the Lambda's **permission** — is `apigateway.amazonaws.com`
   allowed to invoke it?
6. It `POST`s the event to the Lambda service (always POST, see above).
7. `cmd/get_user/main.go` starts the handler; `Get` reads
   `req.PathParameters["userID"]` → `"abc-123"`.
8. The repository calls DynamoDB using the **execution role** from `iam.tf`.
9. The handler returns an `APIGatewayProxyResponse`; API Gateway turns it into a
   real HTTP response.

Every numbered step is a separate thing that can be misconfigured, which is why
the file is longer than it feels like it should be.

## Things that would bite me

- Renaming `{userID}` without updating `req.PathParameters["userID"]` → every
  request 400s, no error anywhere.
- Setting `integration_http_method` from the caller's verb → GET/PUT/DELETE break.
- Forgetting `aws_lambda_permission` → 500 with nothing in the logs.
- Forgetting `triggers` on the deployment → changes apply but nothing changes.
- Adding a route to `local.routes` whose key is not a folder in `golang/cmd/` →
  `aws_lambda_function.functions[each.key]` fails on a missing key.
