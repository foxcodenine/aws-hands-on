package repository

import (
	"03-dynamodb_crud_with_go/internal/models"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// fakeDynamo implements DynamoDBAPI. Each method just calls the func field the
// test sets, so every test controls exactly the calls it cares about.
type fakeDynamo struct {
	putItem func(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	getItem func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
}

func (f *fakeDynamo) PutItem(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return f.putItem(in)
}

func (f *fakeDynamo) GetItem(_ context.Context, in *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return f.getItem(in)
}

// Unused by these tests, but needed to satisfy the interface.
func (f *fakeDynamo) UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return nil, nil
}
func (f *fakeDynamo) DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return nil, nil
}
func (f *fakeDynamo) Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return nil, nil
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

// Compile-time proof the fake stays in sync with the interface.
var _ DynamoDBAPI = (*fakeDynamo)(nil)
