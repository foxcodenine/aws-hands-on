package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ddb       *dynamodb.Client
	tableName = "lambda-apigateway"
)

// event mirrors the Python/Node shape: an operation name plus a payload
// whose fields map straight onto the matching DynamoDB input.
//
// In Python, boto3's Table resource lets you pass plain dicts like
// {"id": "1234ABCD"} straight into put_item/get_item - it converts them to
// DynamoDB's internal typed format for you. The Go SDK's low-level client
// doesn't do that conversion automatically, so Item/Key/ExpressionAttributeValues
// are read here as plain JSON first, then converted with the toAttrs()
// helper below right before each DynamoDB call.
type event struct {
	Operation string `json:"operation"`
	Payload   struct {
		Item                      map[string]interface{} `json:"Item,omitempty"`
		Key                       map[string]interface{} `json:"Key,omitempty"`
		UpdateExpression          string                  `json:"UpdateExpression,omitempty"`
		ExpressionAttributeNames  map[string]string       `json:"ExpressionAttributeNames,omitempty"`
		ExpressionAttributeValues map[string]interface{}  `json:"ExpressionAttributeValues,omitempty"`
	} `json:"payload"`
}

// toAttrs converts plain values (e.g. {"id": "1234ABCD"}) into DynamoDB's
// typed attribute format (e.g. {"id": {"S": "1234ABCD"}}).
func toAttrs(m map[string]interface{}) map[string]types.AttributeValue {
	attrs, _ := attributevalue.MarshalMap(m)
	return attrs
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var e event

	err := json.Unmarshal([]byte(req.Body), &e)

	if err != nil {
		return respond(400, map[string]string{"error": "invalid request body"})
	}

	var out any

	switch e.Operation {
	case "create":
		out, err = ddb.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName), Item: toAttrs(e.Payload.Item),
		})

	case "read":
		out, err = ddb.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tableName), Key: toAttrs(e.Payload.Key),
		})

	case "update":
		out, err = ddb.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       toAttrs(e.Payload.Key),
			UpdateExpression:          aws.String(e.Payload.UpdateExpression),
			ExpressionAttributeNames:  e.Payload.ExpressionAttributeNames,
			ExpressionAttributeValues: toAttrs(e.Payload.ExpressionAttributeValues),
		})

	case "delete":
		out, err = ddb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName), Key: toAttrs(e.Payload.Key),
		})

	case "echo":
		// Payload is typed for the CRUD fields above, so arbitrary keys like
		// {"somekey1": "somevalue1"} get silently dropped when unmarshaled
		// into it. Re-parse the body generically here instead, so echo
		// actually returns whatever was sent - same as the Python version.
		var raw struct {
			Payload interface{} `json:"payload"`
		}
		json.Unmarshal([]byte(req.Body), &raw)
		out = raw.Payload

	default:
		return respond(400, map[string]string{"error": fmt.Sprintf("unrecognized operation %q", e.Operation)})

	}

	if err != nil {
		return respond(500, map[string]string{"error": err.Error()})
	}
	return respond(200, out)

}

// ---------------------------------------------------------------------

func respond(status int, body any) (events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(body)

	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

// ---------------------------------------------------------------------

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		panic(err)
	}

	ddb = dynamodb.NewFromConfig(cfg)

	lambda.Start(handler)
}
