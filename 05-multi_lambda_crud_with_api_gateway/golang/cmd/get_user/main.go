package main

import (
	"05-multi_lambda_crud_with_api_gateway/internal/api"
	"05-multi_lambda_crud_with_api_gateway/internal/db"
	"05-multi_lambda_crud_with_api_gateway/internal/logging"
	"05-multi_lambda_crud_with_api_gateway/internal/repository"
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	logging.Setup()

	ddb := db.NewClient(os.Stdout, context.Background())

	userHandler := api.NewHandler(repository.NewRepository(ddb))
	lambda.Start(userHandler.Get)
}
