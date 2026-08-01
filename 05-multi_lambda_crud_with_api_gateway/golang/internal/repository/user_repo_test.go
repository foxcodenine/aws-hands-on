package repository

import (
	"05-multi_lambda_crud_with_api_gateway/internal/models"
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// fakeDynamo implements DynamoDBAPI. Each method just calls the func field the
// test sets, so every test controls exactly the calls it cares about.
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

func (f *fakeDynamo) UpdateItem(_ context.Context, in *dynamodb.UpdateItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return f.updateItem(in)
}

func (f *fakeDynamo) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return f.deleteItem(in)
}
func (f *fakeDynamo) Query(_ context.Context, in *dynamodb.QueryInput, _ ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return f.query(in)
}
func (f *fakeDynamo) Scan(_ context.Context, in *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return f.scan(in)
}

func TestCreate(t *testing.T) {
	// Capture what the repo sends to DynamoDB so we can assert on it.
	var captured *dynamodb.PutItemInput

	fake := &fakeDynamo{
		putItem: func(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			captured = in
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	repo := NewUserRepository(fake)

	user, err := repo.Create(context.Background(), models.CreateUserInput{
		Name:  "Ada",
		Email: "ada@example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The repo should fill in the fields the caller doesn't provide.
	if user.UserID == "" {
		t.Error("expected a generated UserID, got empty string")
	}
	if user.Status != "active" {
		t.Errorf("Status = %q, want %q", user.Status, "active")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// And it should send the insert-only condition so we never overwrite.
	if captured == nil {
		t.Fatal("PutItem was never called")
	}
	if got := *captured.ConditionExpression; got != "attribute_not_exists(user_id)" {
		t.Errorf("ConditionExpression = %q, want the insert-only guard", got)
	}
}

func TestGetByID(t *testing.T) {
	t.Run("returns nil when the item does not exist", func(t *testing.T) {
		fake := &fakeDynamo{
			getItem: func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
				// A missing item comes back with a nil Item, not an error.
				return &dynamodb.GetItemOutput{Item: nil}, nil
			},
		}
		repo := NewUserRepository(fake)

		user, err := repo.GetByID(context.Background(), "missing-id")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got %+v", user)
		}
	})

	t.Run("unmarshals a found item into a User", func(t *testing.T) {
		item, _ := attributevalue.MarshalMap(models.User{
			UserID: "abc-123",
			Name:   "Grace",
			Status: "active",
		})

		fake := &fakeDynamo{
			getItem: func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{Item: item}, nil
			},
		}
		repo := NewUserRepository(fake)

		user, err := repo.GetByID(context.Background(), "abc-123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user == nil {
			t.Fatal("expected a user, got nil")
		}
		if user.Name != "Grace" {
			t.Errorf("Name = %q, want %q", user.Name, "Grace")
		}
	})
}

func TestQueryByEmailUsesEmailIndex(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "user-123",
		Email:  "ada@example.com",
	})
	var captured *dynamodb.QueryInput
	fake := &fakeDynamo{
		query: func(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			captured = in
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}}, nil
		},
	}
	repo := NewUserRepository(fake)

	users, err := repo.QueryByEmail(context.Background(), "ada@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("Query was not called")
	}
	if captured.IndexName == nil || *captured.IndexName != "email-index" {
		t.Fatalf("IndexName = %v, want email-index", captured.IndexName)
	}
	if len(users) != 1 || users[0].Email != "ada@example.com" {
		t.Errorf("users = %+v, want one matching user", users)
	}
}

func TestQueryByEmailReturnsDuplicateResults(t *testing.T) {
	first, _ := attributevalue.MarshalMap(models.User{UserID: "user-1", Email: "same@example.com"})
	second, _ := attributevalue.MarshalMap(models.User{UserID: "user-2", Email: "same@example.com"})
	fake := &fakeDynamo{
		query: func(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{first, second}}, nil
		},
	}
	repo := NewUserRepository(fake)

	users, err := repo.QueryByEmail(context.Background(), "same@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d users, want 2 duplicate results", len(users))
	}
}

func TestUpdateOnlyTouchesUsersThatExist(t *testing.T) {
	// UpdateItem creates the item if it is missing ("upsert"), which would leave
	// a half-empty record for an ID that was never real. The repo guards against
	// that with attribute_exists, so check the guard is actually sent.
	var captured *dynamodb.UpdateItemInput

	fake := &fakeDynamo{
		updateItem: func(in *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			captured = in
			return &dynamodb.UpdateItemOutput{}, nil
		},
	}
	repo := NewUserRepository(fake)

	_, err := repo.Update(context.Background(), "abc-123", "Ada", "ada@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("UpdateItem was never called")
	}
	if captured.ConditionExpression == nil ||
		!strings.Contains(*captured.ConditionExpression, "attribute_exists") {
		t.Errorf("ConditionExpression = %v, want an attribute_exists guard", captured.ConditionExpression)
	}
}

func TestDeleteReturnsNilWhenThereWasNothingToDelete(t *testing.T) {
	// DeleteItem succeeds even for a key that never existed, so the repo asks for
	// the old attributes back. Empty attributes is the only way to tell the
	// difference between "deleted it" and "there was nothing there".
	fake := &fakeDynamo{
		deleteItem: func(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
			return &dynamodb.DeleteItemOutput{Attributes: nil}, nil
		},
	}
	repo := NewUserRepository(fake)

	user, err := repo.Delete(context.Background(), "never-existed")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %+v", user)
	}
}

func TestList(t *testing.T) {
	// List fetches exactly one page and hands back a cursor, so the caller
	// decides whether to ask for more.

	t.Run("reads one page and returns the next cursor", func(t *testing.T) {
		item, _ := attributevalue.MarshalMap(models.User{UserID: "user-1", Name: "Ada"})
		var calls []*dynamodb.ScanInput

		fake := &fakeDynamo{
			scan: func(in *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
				calls = append(calls, in)
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{item},
					// More rows exist after this one.
					LastEvaluatedKey: map[string]types.AttributeValue{
						"user_id": &types.AttributeValueMemberS{Value: "user-1"},
					},
				}, nil
			},
		}
		repo := NewUserRepository(fake)

		users, next, err := repo.List(context.Background(), 25, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(calls) != 1 {
			t.Fatalf("scan calls = %d, want 1 - List must not page through the table", len(calls))
		}
		if *calls[0].Limit != 25 {
			t.Errorf("Limit = %d, want 25", *calls[0].Limit)
		}
		if len(users) != 1 {
			t.Errorf("users = %+v, want 1", users)
		}
		if next != "user-1" {
			t.Errorf("next cursor = %q, want user-1", next)
		}
	})

	t.Run("sends the cursor back as the start key", func(t *testing.T) {
		var captured *dynamodb.ScanInput

		fake := &fakeDynamo{
			scan: func(in *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
				captured = in
				return &dynamodb.ScanOutput{}, nil
			},
		}
		repo := NewUserRepository(fake)

		_, _, err := repo.List(context.Background(), 10, "user-1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		startKey, ok := captured.ExclusiveStartKey["user_id"].(*types.AttributeValueMemberS)
		if !ok || startKey.Value != "user-1" {
			t.Errorf("ExclusiveStartKey = %v, want user-1", captured.ExclusiveStartKey)
		}
	})

	t.Run("returns an empty slice rather than nil", func(t *testing.T) {
		// A nil slice marshals to JSON null; the API should send [] instead.
		fake := &fakeDynamo{
			scan: func(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
				return &dynamodb.ScanOutput{Items: nil}, nil
			},
		}
		repo := NewUserRepository(fake)

		users, next, err := repo.List(context.Background(), 25, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if users == nil {
			t.Error("users is nil, want an empty slice")
		}
		if next != "" {
			t.Errorf("next cursor = %q, want empty on the last page", next)
		}
	})
}

// Compile-time proof the fake stays in sync with the interface.
var _ DynamoDBAPI = (*fakeDynamo)(nil)
