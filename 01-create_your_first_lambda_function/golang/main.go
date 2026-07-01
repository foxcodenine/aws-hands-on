package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// ---------------------------------------------------------------------

// Event defines the shape of the JSON input Lambda receives.
// The runtime deserializes the incoming JSON payload into this struct.
type Event struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
}

// Response defines the shape of the JSON output Lambda returns.
type Response struct {
	Area float64 `json:"area"`
}

// ---------------------------------------------------------------------

// calculateArea computes the area of a rectangle.
func calculateArea(length, width float64) float64 {
	return length * width
}

// ---------------------------------------------------------------------

// handler is the Lambda entry point. AWS calls this function for each invocation,
// passing the deserialized event and a context that carries Lambda metadata.
func handler(ctx context.Context, event Event) (string, error) {

	// Get the length and width from the event and calculate the area.
	area := calculateArea(event.Length, event.Width)
	fmt.Printf("The area is %g\n", area)

	// lambdacontext.FromContext extracts Lambda-specific metadata from the context.
	// In Python/Node the equivalent is context.log_group_name / context.logGroupName.
	// The Go SDK does not expose the log group name, so we log the function ARN instead.
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		log.Printf("Invoked function ARN: %s", lc.InvokedFunctionArn)
	}

	// Marshal the result into a JSON string and return it.
	data := Response{Area: area}
	out, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func main() {
	// lambda.Start registers the handler and blocks, waiting for invocations from AWS.
	lambda.Start(handler)
}
