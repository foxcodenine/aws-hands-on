package main

import (
	"03-dynamodb_crud_with_go/internal/db"
	"03-dynamodb_crud_with_go/internal/models"
	"03-dynamodb_crud_with_go/internal/repository"
	"context"
	"fmt"
	"io"
	"os"
)

func main() {

	run(os.Stdout)
}

func run(out io.Writer) {

	err := loadEnv(envPath)

	if err != nil {
		fmt.Fprintf(out, "[envConfig] WARNING: no .env at %s: %v\n", envPath, err)
	}

	// -----------------------------------------------------------------

	ctx := context.Background()

	// Initialize the DynamoDB client
	client := db.NewClient(os.Stdout, ctx)
	userRepo := repository.NewUserRepository(client)

	// ---------------------------------------------

	// Create a user
	user, err := createUser(ctx, userRepo, "Alice Johnson", "alice@example.com")
	if err != nil {
		fmt.Fprintf(out, "create failed: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(out, "Created user: %s\n", user.UserID)

	// ---------------------------------------------

	// Read the user back
	found, err := userRepo.GetByID(ctx, user.UserID)

	if err != nil {
		fmt.Fprintf(out, "get failed: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(out, "Found user: %s (%s)\n", found.Name, found.Email)

	// ---------------------------------------------

	// Query by status
	activeUsers, err := userRepo.QueryByStatus(ctx, "active", 10)
	if err != nil {
		fmt.Fprintf(out, "query failed: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(out, "Found %d active users\n", len(activeUsers))

}

func createUser(
	ctx context.Context,
	userRepo *repository.UserRepository,
	name string,
	email string,
) (*models.User, error) {

	user, err := userRepo.Create(ctx, models.CreateUserInput{
		Name:  name,
		Email: email,
		Tags:  []string{"admin", "beta-tester"},
	})

	return user, err
}
