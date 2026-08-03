package models

import "time"

// User represents a user record in DynamoDB.
//
// The dynamodbav tags control how struct fields map to DynamoDB attributes; the
// json tags do the same for the API responses. Both are needed and neither
// reads the other - without the json tags, encoding/json falls back to the Go
// field names and the API answers with "UserID" while it accepts "name".
type User struct {
	UserID    string    `dynamodbav:"user_id" json:"user_id"`
	Name      string    `dynamodbav:"name" json:"name"`
	Email     string    `dynamodbav:"email" json:"email"`
	Status    string    `dynamodbav:"status" json:"status"`
	Age       int       `dynamodbav:"age,omitempty" json:"age,omitempty"`
	Tags      []string  `dynamodbav:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at" json:"updated_at"`
}

// CreateUserInput holds the fields needed to create a new user.
type CreateUserInput struct {
	Name  string
	Email string
	Age   int
	Tags  []string
}
