// Package api holds the Lambda handlers behind the API Gateway routes.
//
// Unlike the net/http handlers in tutorial 03, a Lambda handler takes a request
// value and *returns* a response value instead of writing into a ResponseWriter:
//
//	json.NewDecoder(r.Body).Decode(&x)  ->  json.Unmarshal([]byte(req.Body), &x)
//	chi.URLParam(r, "userID")           ->  req.PathParameters["userID"]
//	w.WriteHeader(200); Encode(x)       ->  return respond(200, x)
//
// The DynamoDB work still lives in the repository, exactly as in tutorial 03.
package api

import (
	"05-multi_lambda_crud_with_api_gateway/internal/models"
	"05-multi_lambda_crud_with_api_gateway/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Page size bounds for List. A cap matters here because Lambda can only return
// a 6 MB response, and every returned item is billed read capacity.
const (
	defaultPageSize = 25
	maxPageSize     = 100
)

// UserHandler has one method per operation. Each Lambda in cmd/ builds it once
// at cold start, then points lambda.Start at the single method its route needs.
type UserHandler struct {
	Repo *repository.Repository
}

// ---------------------------------------------------------------------------------------------------------------------

func NewHandler(repo *repository.Repository) UserHandler {
	return UserHandler{Repo: repo}
}

// ---------------------------------------------------------------------------------------------------------------------

// Create serves POST /users.
func (h *UserHandler) Create(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Step 1: the fields we accept - the repository generates user_id,
	// timestamps and status itself.
	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int64  `json:"age"`
	}

	var body Request

	// Step 2: decode. API Gateway hands the body over as a plain string.
	err := json.Unmarshal([]byte(req.Body), &body)
	if err != nil {
		return respond(400, map[string]string{"error": "invalid JSON"})
	}

	// Step 3: validate before touching the database.
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Email) == "" || body.Age < 0 {
		return respond(400, map[string]string{"error": "name, email, and a non-negative age are required"})
	}

	// Step 4: reject duplicate emails. Read-then-write, so simultaneous requests
	// can still slip through - a GSI can't enforce uniqueness on its own.
	users, err := h.Repo.User.QueryByEmail(ctx, strings.TrimSpace(body.Email))
	if err != nil {
		return respond(500, map[string]string{"error": "failed to check email"})
	}
	if len(users) > 0 {
		return respond(409, map[string]string{"error": "email is already in use"})
	}

	// Step 5: the repository does the actual PutItem.
	user, err := h.Repo.User.Create(ctx, models.CreateUserInput{
		Name:  body.Name,
		Email: strings.TrimSpace(body.Email),
		Age:   int(body.Age),
	})
	if err != nil {
		return respond(500, map[string]string{"error": "failed to create user"})
	}

	// Step 6: return the stored record, so the caller learns the generated ID.
	return respond(201, user)
}

// ---------------------------------------------------------------------------------------------------------------------

// listResponse is what GET /users sends back. omitempty drops next_cursor
// entirely on the last page, so callers can just loop until it's gone.
type listResponse struct {
	Users      []models.User `json:"users"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// List serves GET /users?limit=25&cursor=<last user_id>.
func (h *UserHandler) List(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Step 1: how many to return. QueryStringParameters is the Lambda
	// equivalent of r.URL.Query().
	limit, err := pageLimit(req.QueryStringParameters["limit"])
	if err != nil {
		return respond(400, map[string]string{"error": err.Error()})
	}

	// Step 2: fetch one page. An empty cursor starts from the beginning.
	users, next, err := h.Repo.User.List(ctx, limit, req.QueryStringParameters["cursor"])
	if err != nil {
		return respond(500, map[string]string{"error": "failed to list users"})
	}

	return respond(200, listResponse{Users: users, NextCursor: next})
}

// pageLimit reads ?limit=, falling back to the default when it is missing.
func pageLimit(raw string) (int32, error) {
	if raw == "" {
		return defaultPageSize, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 || n > maxPageSize {
		return 0, fmt.Errorf("limit must be between 1 and %d", maxPageSize)
	}

	return int32(n), nil
}

// ---------------------------------------------------------------------------------------------------------------------

// Get serves GET /users/{userID}.
func (h *UserHandler) Get(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Step 1: the ID comes from the URL - a GET has no body. This key must match
	// the {userID} placeholder in the Terraform route.
	userID := req.PathParameters["userID"]
	if userID == "" {
		return respond(400, map[string]string{"error": "user ID is required"})
	}

	// Step 2: look the user up.
	user, err := h.Repo.User.GetByID(ctx, userID)
	if err != nil {
		return respond(500, map[string]string{"error": "failed to get user"})
	}

	// Step 3: a miss isn't an error in DynamoDB - GetItem just returns nothing,
	// which the repository surfaces as (nil, nil). Turn that into a 404.
	if user == nil {
		return respond(404, map[string]string{"error": "user not found"})
	}

	return respond(200, user)
}

// ---------------------------------------------------------------------------------------------------------------------

// Update serves PUT /users/{userID} - ID from the URL, new values from the body.
func (h *UserHandler) Update(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Step 1: who to update.
	userID := req.PathParameters["userID"]
	if userID == "" {
		return respond(400, map[string]string{"error": "user ID is required"})
	}

	// Step 2: what to change. The repository refreshes updated_at on its own.
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	err := json.Unmarshal([]byte(req.Body), &body)
	if err != nil {
		return respond(400, map[string]string{"error": "invalid JSON"})
	}

	// Step 3: validate.
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Email) == "" {
		return respond(400, map[string]string{"error": "name and email are required"})
	}

	// Step 4: apply it. The repository guards with attribute_exists(user_id), so
	// an unknown ID won't silently create a half-empty record.
	user, err := h.Repo.User.Update(ctx, userID, body.Name, body.Email)
	if err != nil {
		return respond(500, map[string]string{"error": "failed to update user"})
	}
	if user == nil {
		return respond(404, map[string]string{"error": "user not found"})
	}

	// Step 5: return the record as it looks after the update.
	return respond(200, user)
}

// ---------------------------------------------------------------------------------------------------------------------

// Delete serves DELETE /users/{userID}.
func (h *UserHandler) Delete(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Step 1: the ID comes from the URL - a DELETE has no body either.
	userID := req.PathParameters["userID"]
	if userID == "" {
		return respond(400, map[string]string{"error": "user ID is required"})
	}

	// Step 2: delete. DeleteItem succeeds even for a key that was never there,
	// so the repository returns the old record - nil means nothing was deleted.
	user, err := h.Repo.User.Delete(ctx, userID)
	if err != nil {
		return respond(500, map[string]string{"error": "failed to delete user"})
	}
	if user == nil {
		return respond(404, map[string]string{"error": "user not found"})
	}

	// Step 3: built directly, not via respond() - a 204 must carry no body.
	return events.APIGatewayProxyResponse{StatusCode: 204}, nil
}

// ---------------------------------------------------------------------------------------------------------------------

// Echo returns whatever JSON it was sent, without touching DynamoDB. Carried
// over from tutorial 04 to test the wiring; no route points at it.
func (h *UserHandler) Echo(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// interface{} accepts any JSON shape, so nothing sent gets dropped.
	var body interface{}

	err := json.Unmarshal([]byte(req.Body), &body)

	if err != nil {
		return respond(400, map[string]string{"error": "invalid request body"})
	}

	return respond(200, body)
}

// ---------------------------------------------------------------------------------------------------------------------

// respond replaces the w.WriteHeader + json.NewEncoder(w).Encode pair from a
// net/http handler: the body is marshalled and handed back, not written out.
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
