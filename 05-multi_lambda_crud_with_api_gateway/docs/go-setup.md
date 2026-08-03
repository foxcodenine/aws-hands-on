# Initialize the module

go mod init 05-multi_lambda_crud_with_api_gateway

# Install the AWS SDK v2 modules
go get github.com/aws/aws-lambda-go/lambda
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression
go get github.com/google/uuid
