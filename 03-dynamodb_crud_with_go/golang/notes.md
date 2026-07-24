# Initialize the module

go mod init 03-dynamodb_crud_with_go

# Install the AWS SDK v2 modules
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression
go get github.com/google/uuid

# Install go-chi
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
go get github.com/go-chi/cors



