package db

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewClient creates a configured DynamoDB client.
// It uses the default credential chain (env vars, shared config, IAM role).
func NewClient(w io.Writer, ctx context.Context) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(getRegion()),
	)

	if err != nil {
		fmt.Fprintf(w, "unable to load AWS config:%v", err)
	}

	// Support endpoint override for local development
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	if endpoint != "" {
		return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		})
	}

	return dynamodb.NewFromConfig(cfg)
}

// ---------------------------------------------------------------------

func getRegion() string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}

	return "eu-west-1"
}
