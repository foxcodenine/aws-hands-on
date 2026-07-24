package repository

import (
	"03-dynamodb_crud_with_go/internal/models"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// fakeDynamo implements DynamoDBAPI. Each method just calls the func field the
// test sets, so every test controls exactly the calls it cares about.
type fakeDynamo struct {
	putItem func(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	getItem func(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	query   func(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	scan    func(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
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

func TestListAllCombinesScanPages(t *testing.T) {
	first, _ := attributevalue.MarshalMap(models.User{UserID: "user-1", Name: "Ada"})
	second, _ := attributevalue.MarshalMap(models.User{UserID: "user-2", Name: "Grace"})
	startKey := map[string]types.AttributeValue{
		"user_id": &types.AttributeValueMemberS{Value: "user-1"},
	}
	var calls []*dynamodb.ScanInput
	fake := &fakeDynamo{
		scan: func(in *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
			calls = append(calls, in)
			if len(calls) == 1 {
				return &dynamodb.ScanOutput{
					Items:            []map[string]types.AttributeValue{first},
					LastEvaluatedKey: startKey,
				}, nil
			}
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{second}}, nil
		},
	}
	repo := NewUserRepository(fake)

	users, err := repo.ListAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("scan calls = %d, want 2", len(calls))
	}
	if len(calls[1].ExclusiveStartKey) != 1 {
		t.Fatalf("second scan did not use the first page key")
	}
	if len(users) != 2 {
		t.Fatalf("users = %+v, want 2 users", users)
	}
}

// Compile-time proof the fake stays in sync with the interface.
var _ DynamoDBAPI = (*fakeDynamo)(nil)
