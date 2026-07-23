package models

import "time"

// User represents a user record in DynamoDB.
// The dynamodbav tags control how struct fields map to DynamoDB attributes.
type User struct {
	UserID    string    `dynamodbav:"user_id"`
	Name      string    `dynamodbav:"name"`
	Email     string    `dynamodbav:"email"`
	Status    string    `dynamodbav:"status"`
	Age       int       `dynamodbav:"age,omitempty"`
	Tags      []string  `dynamodbav:"tags,omitempty"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

// CreateUserInput holds the fields needed to create a new user.
type CreateUserInput struct {
	Name  string
	Email string
	Age   int
	Tags  []string
}
