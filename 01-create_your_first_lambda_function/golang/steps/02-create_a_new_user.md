# Create a New IAM User

## Part 1: Set Up Group and User (AWS Console Terminal)

### Step 1: Create the `developers` Group and Attach Policies

```bash

~ $ aws iam create-group --group-name developers
{
    "Group": {
        "Path": "/",
        "GroupName": "developers",
        "GroupId": "AGPA2RWPVJZER4XZWICNI",
        "Arn": "arn:aws:iam::725211237961:group/developers",
        "CreateDate": "2026-07-01T14:45:29+00:00"
    }
}
~ $ aws iam attach-group-policy --group-name developers --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess
~ $ aws iam attach-group-policy --group-name developers --policy-arn arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator
~ $ aws iam attach-group-policy --group-name developers --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess
~ $ aws iam attach-group-policy --group-name developers --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
~ $ aws iam attach-group-policy --group-name developers --policy-arn arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
~ $ 
~ $ 
~ $ aws iam create-user --user-name developer
{
    "User": {
        "Path": "/",
        "UserName": "developer",
        "UserId": "AI***********42",
        "Arn": "arn:aws:iam::725211237961:user/developer",
        "CreateDate": "2026-07-01T14:50:32+00:00"
    }
}
~ $ 
~ $ 
~ $ aws iam add-user-to-group --user-name developer --group-name developers
~ $ 
~ $ 
~ $ aws iam create-access-key --user-name developer
{
    "AccessKey": {
        "UserName": "developer",
        "AccessKeyId": "AK******AN",
        "Status": "Active",
        "SecretAccessKey": "M5**************12",
        "CreateDate": "2026-07-01T14:52:40+00:00"
    }
}
~ $ 
```

---

## Part 2: Create a Custom IAM Policy for Lambda Role Management

Rather than granting broad IAM access, this custom policy gives the `developers` group only the specific permissions needed to create and manage Lambda execution roles.

### Step 2: Write the Policy Document to a File

```bash
$ cat > lambda-role-management-policy.json << 'EOF'
> {
>   "Version": "2012-10-17",
>   "Statement": [
>     {
>       "Effect": "Allow",
>       "Action": [
>         "iam:CreateRole",
>         "iam:DeleteRole",
>         "iam:AttachRolePolicy",
>         "iam:DetachRolePolicy",
>         "iam:PassRole",
>         "iam:GetRole",
>         "iam:ListRolePolicies",
>         "iam:ListAttachedRolePolicies"
>       ],
>       "Resource": "*"
>     }
>   ]
> }
> EOF
~ $ 
```

### Step 3: Create the Named Custom Policy in AWS

```bash
~ $ aws iam create-policy \
>   --policy-name LambdaRoleManagement \
>   --policy-document file://lambda-role-management-policy.json
{
    "Policy": {
        "PolicyName": "LambdaRoleManagement",
        "PolicyId": "ANPA2RWPVJZE2MTCNJ7KX",
        "Arn": "arn:aws:iam::725211237961:policy/LambdaRoleManagement",
        "Path": "/",
        "DefaultVersionId": "v1",
        "AttachmentCount": 0,
        "PermissionsBoundaryUsageCount": 0,
        "IsAttachable": true,
        "CreateDate": "2026-07-01T15:49:54+00:00",
        "UpdateDate": "2026-07-01T15:49:54+00:00"
    }
}
```

### Step 4: Attach the Custom Policy and Verify All Attached Policies

```bash
~ $ aws iam attach-group-policy \
>   --group-name developers \
>   --policy-arn arn:aws:iam::725211237961:policy/LambdaRoleManagement
~ $ 
~ $ aws iam list-attached-group-policies --group-name developers
{
    "AttachedPolicies": [
        {
            "PolicyName": "AmazonAPIGatewayAdministrator",
            "PolicyArn": "arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator"
        },
        {
            "PolicyName": "CloudWatchLogsFullAccess",
            "PolicyArn": "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
        },
        {
            "PolicyName": "AmazonDynamoDBFullAccess",
            "PolicyArn": "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
        },
        {
            "PolicyName": "AmazonS3FullAccess",
            "PolicyArn": "arn:aws:iam::aws:policy/AmazonS3FullAccess"
        },
        {
            "PolicyName": "AWSLambda_FullAccess",
            "PolicyArn": "arn:aws:iam::aws:policy/AWSLambda_FullAccess"
        },
        {
            "PolicyName": "LambdaRoleManagement",
            "PolicyArn": "arn:aws:iam::725211237961:policy/LambdaRoleManagement"
        }
    ]
}
```

---

## Part 3: Verify the Setup Locally

Use the `developer` profile configured in `~/.aws/credentials` to confirm the user is set up correctly and can access Lambda.

```bash
foxcodenine@foxcodenine-NUC12WSHi7:~$ aws sts get-caller-identity --profile developer
{
    "UserId": "AIDA2RWPVJZEQ6MJCYK42",
    "Account": "725211237961",
    "Arn": "arn:aws:iam::725211237961:user/developer"
}
foxcodenine@foxcodenine-NUC12WSHi7:~$ aws lambda list-functions --profile developer
{
    "Functions": []
}
```
