# Deploy Go Lambda via AWS Dashboard

### Step 1: Build and Zip

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

zip myLambdaFunctionGo.zip bootstrap
```

(or just re-run the zip command with the new name — either works, the internal `bootstrap` filename is what matters)

### Step 2: Create the function in the console

1. Go to **Lambda → Functions → Create function**
2. Choose **Author from scratch**
3. **Function name:** `myLambdaFunctionGo`
4. **Runtime:** scroll down — Go isn't listed as a managed runtime anymore, so choose **Amazon Linux 2023** under the "Provide your own bootstrap on Amazon Linux 2023" option (this is the `provided.al2023` runtime we used for the CLI approach)
5. **Architecture:** `x86_64` (matches your `GOARCH=amd64` build)
6. **Execution role:** leave the default — **"Create a new role with basic Lambda permissions"** (this is the same auto-role-creation you saw with Python/Node)
7. Click **Create function**

### Step 3: Upload your code

1. In the function's page, scroll to the **Code** section
2. Click **Upload from → .zip file**
3. Select `myLambdaFunctionGo.zip`
4. Click **Save**

### Step 4: Set the handler (important — this one trips people up)

Since you're using `provided.al2023`, go to **Configuration → General configuration → Edit**, and make sure **Handler** is set to `bootstrap`. If it's blank or wrong, the runtime won't know what to execute.

### Step 5: Test it

Go to the **Test** tab, create a new test event:

```json
{ "length": 5, "width": 3 }
```

Click **Test** — you should get back `{"area":15}` and see your `fmt.Printf` and `log.Printf` lines in the execution results / CloudWatch logs.

---

### Cleanup

**Console:** go to the function → **Actions → Delete** → confirm.

Or via CLI, if you want to start practicing there already:

```bash
aws lambda delete-function --function-name myLambdaFunctionGo
```
