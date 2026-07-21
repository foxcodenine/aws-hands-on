package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	// Step 1: read bucket name and object key out of the S3 event record
	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name

	// Step 2: decode the key (S3 URL-encodes spaces/special chars in event keys)

	key, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		return "", fmt.Errorf("failed to decode key %s: %w", record.S3.Object.Key, err)
	}

	// Step 3: load AWS config and create the S3 client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	// Step 4: fetch the object's metadata (headers only, no body) and read its content type
	output, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		// Step 5: log and return the error so the Lambda invocation shows as failed
		log.Printf("Error getting object %s from bucket %s. Make sure they exist and your bucket is in the same region as this function.", key, bucket)
		return "", err
	}

	contentType := *output.ContentType

	log.Printf("CONTENT TYPE: %s", contentType)
	log.Printf("BUCKET: %s, FILE NAME: %s", bucket, key)
	return contentType, nil
}

func main() {
	lambda.Start(handler)
}

// ---------------------------------------------------------------------

// go mod init s3TriggerTutorial
// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

//   go get github.com/aws/aws-lambda-go/lambda
//   go get github.com/aws/aws-sdk-go-v2/config
//   go get github.com/aws/aws-sdk-go-v2/service/s3
//   go get github.com/aws/aws-lambda-go/events
