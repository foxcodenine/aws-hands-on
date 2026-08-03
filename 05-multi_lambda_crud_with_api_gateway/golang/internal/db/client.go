package db

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewClient creates a configured DynamoDB client.
// It uses the default credential chain (env vars, shared config, IAM role).
func NewClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(getRegion()),
	)

	// Returned, not logged: main already logs it through slog, and writing here
	// too would put a second, non-JSON copy on the same stdout stream.
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	// Support endpoint override for local development
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	if endpoint != "" {
		return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		}), nil
	}

	return dynamodb.NewFromConfig(cfg), nil
}

// ---------------------------------------------------------------------

func getRegion() string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}

	return "eu-west-1"
}
