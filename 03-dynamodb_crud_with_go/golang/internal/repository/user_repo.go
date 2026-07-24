package repository

import (
	"03-dynamodb_crud_with_go/internal/models"
	"context"
	"fmt"

	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const tableName = "learning-Users"

// DynamoDBAPI is the subset of the DynamoDB client this repo actually uses.
// Depending on this interface (instead of the concrete *dynamodb.Client) lets
// tests pass in a fake. The real *dynamodb.Client satisfies it automatically.
type DynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// UserRepository handles DynamoDB operations for users.
type UserRepository struct {
	client DynamoDBAPI
}

// NewUserRepository creates a new repository instance.
func NewUserRepository(client DynamoDBAPI) *UserRepository {
	return &UserRepository{client: client}
}

// -----------------------------------------------------------------------------

// Create inserts a new user into DynamoDB.
func (r *UserRepository) Create(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Build the full user record: we generate the ID and timestamps here so
	// the caller only supplies the human-provided fields.
	now := time.Now().UTC()
	user := models.User{
		UserID:    uuid.New().String(),
		Name:      input.Name,
		Email:     input.Email,
		Status:    "active",
		Age:       input.Age,
		Tags:      input.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// DynamoDB doesn't take Go structs directly - turn the struct into its
	// attribute-map form (keyed by the `dynamodbav` struct tags).
	item, err := attributevalue.MarshalMap(user)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal user:%w", err)
	}

	// Write the item. PutItem overwrites by default, so the condition below
	// makes this a real "insert only" - it fails if that user_id already exists.
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,

		// Prevent overwriting if the ID somehow already exists
		ConditionExpression: aws.String("attribute_not_exists(user_id)"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to put item: %w", err)
	}

	// Return the record we built so the caller gets the generated ID/timestamps.
	return &user, nil
}

// -----------------------------------------------------------------------------

// GetByID retrieves a user by their partition key.
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {

	// GetItem fetches a single item by its exact key (here the partition key).
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),

		// The key must include the typed value - S = string attribute.
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},

		// Strongly consistent read: return the latest write, not a possibly
		// stale replica. Costs a bit more; fine for a single lookup.
		ConsistentRead: aws.Bool(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// A missing item isn't an error in DynamoDB - it just returns an empty
	// result. We surface that as (nil, nil) for the caller to interpret.
	if result.Item == nil {
		return nil, nil
	}

	// Reverse of MarshalMap: attribute map -> Go struct.
	var user models.User

	err = attributevalue.UnmarshalMap(result.Item, &user)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// QueryByEmail finds users with a matching email address.
func (r *UserRepository) QueryByEmail(ctx context.Context, email string) ([]models.User, error) {
	keyCond := expression.KeyEqual(
		expression.Key("email"), expression.Value(email),
	)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build email expression: %w", err)
	}

	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("email-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query by email: %w", err)
	}

	var users []models.User
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal email results: %w", err)
	}

	return users, nil
}

// -----------------------------------------------------------------------------

// Update modifies specific fields on an existing user.
func (r *UserRepository) Update(ctx context.Context, userID string, name string, email string) (*models.User, error) {

	// Build the SET clause with the expression builder instead of hand-writing
	// the update string - it handles escaping and reserved-word aliasing for us.
	update := expression.Set(
		expression.Name("name"), expression.Value(name),
	).Set(
		expression.Name("email"), expression.Value(email),
	).Set(
		expression.Name("updated_at"), expression.Value(time.Now().UTC()),
	)

	// Only update if the user actually exists - otherwise UpdateItem would
	// silently create ("upsert") a new, half-empty item.
	condition := expression.AttributeExists(expression.Name("user_id"))

	// Compile the pieces into the concrete strings/maps the API expects.
	expr, err := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	result, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{

		TableName: aws.String(tableName),

		// Which item to update.
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},

		// The compiled expression pieces are passed as separate fields.
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),

		// Ask DynamoDB to return the item as it looks *after* the update, so
		// we can hand the fresh record back without a second read.
		ReturnValues: types.ReturnValueAllNew,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	// result.Attributes holds the post-update item (because of ReturnValueAllNew).
	var user models.User

	err = attributevalue.UnmarshalMap(result.Attributes, &user)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated user: %w", err)
	}

	return &user, nil
}

// -----------------------------------------------------------------------------

// Delete removes a user and returns the deleted item.
func (r *UserRepository) Delete(ctx context.Context, userID string) (*models.User, error) {

	result, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{

		TableName: aws.String(tableName),

		// Which item to delete.
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},

		// Return the item as it was *before* deletion, so we can report back
		// what we removed (and detect "nothing was there").
		ReturnValues: types.ReturnValueAllOld,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to delete item: %w", err)
	}

	// No old attributes means there was no such item to delete.
	if result.Attributes == nil {
		return nil, nil
	}

	var user models.User

	err = attributevalue.UnmarshalMap(result.Attributes, &user)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &user, nil
}

// QueryByStatus finds users by status using a GSI.
func (r *UserRepository) QueryByStatus(ctx context.Context, status string, limit int32) ([]models.User, error) {

	// Query needs a key condition on the index's partition key. This targets
	// the "status-index" GSI, whose partition key is `status`.
	keyCond := expression.KeyEqual(
		expression.Key("status"), expression.Value(status),
	)

	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %w", err)
	}

	// Query reads a set of items sharing a partition key - unlike Scan, it
	// doesn't read the whole table. IndexName points it at the GSI.
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("status-index"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     &limit,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	// Bulk version of UnmarshalMap: a list of items -> a slice of structs.
	var users []models.User

	err = attributevalue.UnmarshalListOfMaps(result.Items, &users)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return users, nil
}

// ListAll returns all users in the table.
func (r *UserRepository) ListAll(ctx context.Context) ([]models.User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	var items []map[string]types.AttributeValue
	for {
		result, err := r.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan users: %w", err)
		}

		items = append(items, result.Items...)
		if len(result.LastEvaluatedKey) == 0 {
			break
		}
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}

	var users []models.User
	if err := attributevalue.UnmarshalListOfMaps(items, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}
	return users, nil
}
