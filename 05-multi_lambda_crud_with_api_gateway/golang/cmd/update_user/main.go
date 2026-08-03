package main

import (
	"05-multi_lambda_crud_with_api_gateway/internal/api"
	"05-multi_lambda_crud_with_api_gateway/internal/db"
	"05-multi_lambda_crud_with_api_gateway/internal/logging"
	"05-multi_lambda_crud_with_api_gateway/internal/repository"
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

// main is the only place that exits - everything below it returns errors.
func main() {
	err := run()

	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logging.Setup()

	ddb, err := db.NewClient(context.Background())

	if err != nil {
		return err
	}

	userHandler := api.NewHandler(repository.NewRepository(ddb))

	// Blocks for the life of the execution environment, so nothing after this
	// line runs until Lambda shuts the container down.
	lambda.Start(userHandler.Update)

	return nil
}
