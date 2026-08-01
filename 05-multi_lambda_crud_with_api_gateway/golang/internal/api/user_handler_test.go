package api

import (
	"05-multi_lambda_crud_with_api_gateway/internal/models"
	"05-multi_lambda_crud_with_api_gateway/internal/repository"
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// fakeDynamo stands in for DynamoDB. Each method calls the func field the test
// sets, so a test only has to fill in the calls it actually cares about.
type fakeDynamo struct {
	putItem    func(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	getItem    func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	updateItem func(*dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	deleteItem func(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	query      func(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	scan       func(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

func (f *fakeDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return f.putItem(in)
}
func (f *fakeDynamo) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return f.getItem(in)
}
func (f *fakeDynamo) Query(_ context.Context, in *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return f.query(in)
}
func (f *fakeDynamo) Scan(_ context.Context, in *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return f.scan(in)
}

func (f *fakeDynamo) UpdateItem(_ context.Context, in *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return f.updateItem(in)
}
func (f *fakeDynamo) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return f.deleteItem(in)
}

// newHandler wires a handler to a fake DynamoDB, so no test touches AWS.
func newHandler(fake repository.DynamoDBAPI) UserHandler {
	return NewHandler(repository.NewRepository(fake))
}

// A missing item is not an error in DynamoDB - GetItem succeeds and simply
// returns nothing. Without the nil check in the handler this would be a 200
// with an empty body instead of a 404.
func TestGetReturns404WhenUserMissing(t *testing.T) {
	h := newHandler(&fakeDynamo{
		getItem: func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{}, nil // no Item = not found
		},
	})

	resp, err := h.Get(context.Background(), events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"userID": "nope"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
}

// PathParameters["userID"] has to match the {userID} placeholder in
// api_gateway.tf. If those ever drift apart, every request looks like this.
func TestGetReturns400WithoutUserID(t *testing.T) {
	h := newHandler(&fakeDynamo{})

	resp, _ := h.Get(context.Background(), events.APIGatewayProxyRequest{})

	if resp.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
	}
}

// The email check has to stop the write, not just report a conflict afterwards.
func TestCreateRejectsDuplicateEmail(t *testing.T) {
	wrote := false

	h := newHandler(&fakeDynamo{
		// QueryByEmail finds an existing user with this address.
		query: func(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{
				Items: []map[string]types.AttributeValue{
					{"user_id": &types.AttributeValueMemberS{Value: "existing"}},
				},
			}, nil
		},
		putItem: func(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			wrote = true
			return &dynamodb.PutItemOutput{}, nil
		},
	})

	resp, _ := h.Create(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"name":"Ada","email":"ada@example.com","age":36}`,
	})

	if resp.StatusCode != 409 {
		t.Errorf("StatusCode = %d, want 409", resp.StatusCode)
	}
	if wrote {
		t.Error("saved the user anyway - the duplicate check did not stop the write")
	}
}

// Validation runs before DynamoDB is touched, so a nil fake is safe here: if
// any of these reached the database the test would panic.
func TestCreateRejectsBadInput(t *testing.T) {
	bodies := map[string]string{
		"no name":     `{"name":"","email":"ada@example.com","age":36}`,
		"no email":    `{"name":"Ada","email":"","age":36}`,
		"age below 0": `{"name":"Ada","email":"ada@example.com","age":-1}`,
		"not json":    `not json at all`,
	}

	h := newHandler(&fakeDynamo{})

	for name, body := range bodies {
		resp, _ := h.Create(context.Background(), events.APIGatewayProxyRequest{Body: body})

		if resp.StatusCode != 400 {
			t.Errorf("%s: StatusCode = %d, want 400", name, resp.StatusCode)
		}
	}
}

func TestUpdateRejectsBadInput(t *testing.T) {
	h := newHandler(&fakeDynamo{})

	t.Run("no user ID in the path", func(t *testing.T) {
		resp, _ := h.Update(context.Background(), events.APIGatewayProxyRequest{
			Body: `{"name":"Ada","email":"ada@example.com"}`,
		})
		if resp.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		resp, _ := h.Update(context.Background(), events.APIGatewayProxyRequest{
			PathParameters: map[string]string{"userID": "abc-123"},
			Body:           `{"name":"","email":"ada@example.com"}`,
		})
		if resp.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
		}
	})
}

func TestUpdateReturnsTheUpdatedUser(t *testing.T) {
	// ReturnValues: ALL_NEW means DynamoDB hands back the record as it looks
	// after the write, so the caller does not need a second read.
	updated, _ := attributevalue.MarshalMap(models.User{
		UserID: "abc-123",
		Name:   "Ada King",
		Email:  "ada.king@example.com",
	})

	h := newHandler(&fakeDynamo{
		updateItem: func(*dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			return &dynamodb.UpdateItemOutput{Attributes: updated}, nil
		},
	})

	resp, _ := h.Update(context.Background(), events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"userID": "abc-123"},
		Body:           `{"name":"Ada King","email":"ada.king@example.com"}`,
	})

	if resp.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	var user models.User
	if err := json.Unmarshal([]byte(resp.Body), &user); err != nil {
		t.Fatal(err)
	}
	if user.Name != "Ada King" {
		t.Errorf("Name = %q, want the updated name", user.Name)
	}
}

func TestDeleteReturns204WithNoBody(t *testing.T) {
	// 204 means "done, nothing to send back" - a body here would be invalid,
	// which is why Delete builds its response directly instead of via respond().
	existing, _ := attributevalue.MarshalMap(models.User{UserID: "abc-123"})

	h := newHandler(&fakeDynamo{
		deleteItem: func(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
			// Old attributes come back, so something really was deleted.
			return &dynamodb.DeleteItemOutput{Attributes: existing}, nil
		},
	})

	resp, _ := h.Delete(context.Background(), events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"userID": "abc-123"},
	})

	if resp.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", resp.StatusCode)
	}
	if resp.Body != "" {
		t.Errorf("Body = %q, want empty on a 204", resp.Body)
	}
}

func TestDeleteReturns404WhenNothingWasDeleted(t *testing.T) {
	// DeleteItem succeeds for a key that never existed, so empty old attributes
	// are the only signal that there was nothing to delete.
	h := newHandler(&fakeDynamo{
		deleteItem: func(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
			return &dynamodb.DeleteItemOutput{Attributes: nil}, nil
		},
	})

	resp, _ := h.Delete(context.Background(), events.APIGatewayProxyRequest{
		PathParameters: map[string]string{"userID": "never-existed"},
	})

	if resp.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", resp.StatusCode)
	}
}

// The page size is capped because Lambda can only return a 6 MB response.
func TestListRejectsBadLimit(t *testing.T) {
	h := newHandler(&fakeDynamo{})

	for _, limit := range []string{"0", "-5", "101", "abc"} {
		resp, _ := h.List(context.Background(), events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{"limit": limit},
		})

		if resp.StatusCode != 400 {
			t.Errorf("limit=%q: StatusCode = %d, want 400", limit, resp.StatusCode)
		}
	}
}

// The happy path, mostly to pin down the JSON a caller actually receives.
func TestListReturnsUsersAndNextCursor(t *testing.T) {
	h := newHandler(&fakeDynamo{
		scan: func(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"user_id": &types.AttributeValueMemberS{Value: "u1"}},
				},
				// A LastEvaluatedKey means there is another page after this one.
				LastEvaluatedKey: map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "u1"},
				},
			}, nil
		},
	})

	resp, _ := h.List(context.Background(), events.APIGatewayProxyRequest{})

	if resp.StatusCode != 200 {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Users      []models.User `json:"users"`
		NextCursor string        `json:"next_cursor"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatal(err)
	}

	if len(body.Users) != 1 || body.Users[0].UserID != "u1" {
		t.Errorf("users = %+v, want one user u1", body.Users)
	}
	if body.NextCursor != "u1" {
		t.Errorf("next_cursor = %q, want u1", body.NextCursor)
	}
}
