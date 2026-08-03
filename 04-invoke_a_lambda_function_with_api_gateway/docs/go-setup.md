# Initialize the module

go mod init aws-hands-on/04-invoke_a_lambda_function_with_api_gateway

# Install the AWS SDK v2 modules
go get github.com/aws/aws-lambda-go/lambda
go get github.com/aws/aws-lambda-go/events
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb