package handlers

import (
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"03-dynamodb_crud_with_go/internal/models"
	"03-dynamodb_crud_with_go/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/chi/v5"
)

type fakeDynamo struct {
	putErr       error
	getOutput    *dynamodb.GetItemOutput
	getErr       error
	updateOutput *dynamodb.UpdateItemOutput
	updateErr    error
	deleteOutput *dynamodb.DeleteItemOutput
	deleteErr    error
	queryOutput  *dynamodb.QueryOutput
	queryErr     error
	scanOutput   *dynamodb.ScanOutput
	scanErr      error
}

func (f *fakeDynamo) PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if f.putErr != nil {
		return nil, f.putErr
	}
	return &dynamodb.PutItemOutput{}, nil
}

func (f *fakeDynamo) GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return f.getOutput, f.getErr
}

func (f *fakeDynamo) UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return f.updateOutput, f.updateErr
}

func (f *fakeDynamo) DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return f.deleteOutput, f.deleteErr
}

func (f *fakeDynamo) Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if f.queryOutput == nil && f.queryErr == nil {
		return &dynamodb.QueryOutput{}, nil
	}
	return f.queryOutput, f.queryErr
}

func (f *fakeDynamo) Scan(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if f.scanOutput == nil && f.scanErr == nil {
		return &dynamodb.ScanOutput{}, nil
	}
	return f.scanOutput, f.scanErr
}

func newTestHandler(client repository.DynamoDBAPI) *UserHandler {
	return &UserHandler{
		App: &apptypes.App{Repo: repository.NewRepository(client)},
	}
}

func TestStoreInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("{"))
	recorder := httptest.NewRecorder()

	(&UserHandler{}).Store(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestStoreValidation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"ada@example.com","age":30}`))
	recorder := httptest.NewRecorder()

	(&UserHandler{}).Store(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestStoreCreatesUser(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{})
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Ada","email":"ada@example.com","age":30}`))
	recorder := httptest.NewRecorder()

	handler.Store(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	var user models.User
	if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.Name != "Ada" || user.Email != "ada@example.com" {
		t.Errorf("user = %+v, want Ada / ada@example.com", user)
	}
}

func TestStoreRepositoryError(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{putErr: errors.New("DynamoDB unavailable")})
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Ada","email":"ada@example.com","age":30}`))
	recorder := httptest.NewRecorder()

	handler.Store(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestStoreRejectsExistingEmail(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "existing-user",
		Email:  "ada@example.com",
	})
	handler := newTestHandler(&fakeDynamo{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}},
	})
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Ada","email":"ada@example.com","age":30}`))
	recorder := httptest.NewRecorder()

	handler.Store(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func requestWithUserIDAndBody(method, userID, body string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("userID", userID)
	req := httptest.NewRequest(method, "/users/"+userID, strings.NewReader(body))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
}

func requestWithUserID(userID string) *http.Request {
	return requestWithUserIDAndBody(http.MethodGet, userID, "")
}

func TestShowReturnsUser(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "user-123",
		Name:   "Ada",
		Email:  "ada@example.com",
	})
	handler := newTestHandler(&fakeDynamo{
		getOutput: &dynamodb.GetItemOutput{Item: item},
	})
	recorder := httptest.NewRecorder()

	handler.Show(recorder, requestWithUserID("user-123"))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var user models.User
	if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.Name != "Ada" {
		t.Errorf("Name = %q, want %q", user.Name, "Ada")
	}
}

func TestShowReturnsNotFound(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{
		getOutput: &dynamodb.GetItemOutput{},
	})
	recorder := httptest.NewRecorder()

	handler.Show(recorder, requestWithUserID("missing"))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestShowReturnsRepositoryError(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{getErr: errors.New("DynamoDB unavailable")})
	recorder := httptest.NewRecorder()

	handler.Show(recorder, requestWithUserID("user-123"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestUpdateReturnsUser(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "user-123",
		Name:   "Ada Lovelace",
		Email:  "ada@example.com",
	})
	handler := newTestHandler(&fakeDynamo{
		updateOutput: &dynamodb.UpdateItemOutput{Attributes: item},
	})
	req := requestWithUserIDAndBody(http.MethodPut, "user-123", `{"name":"Ada Lovelace","email":"ada@example.com"}`)
	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var user models.User
	if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.Name != "Ada Lovelace" {
		t.Errorf("Name = %q, want %q", user.Name, "Ada Lovelace")
	}
}

func TestUpdateValidation(t *testing.T) {
	req := requestWithUserIDAndBody(http.MethodPut, "user-123", `{"name":"","email":"ada@example.com"}`)
	recorder := httptest.NewRecorder()

	(&UserHandler{}).Update(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUpdateRepositoryError(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{updateErr: errors.New("DynamoDB unavailable")})
	req := requestWithUserIDAndBody(http.MethodPut, "user-123", `{"name":"Ada","email":"ada@example.com"}`)
	recorder := httptest.NewRecorder()

	handler.Update(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestDestroyDeletesUser(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{UserID: "user-123", Name: "Ada"})
	handler := newTestHandler(&fakeDynamo{
		deleteOutput: &dynamodb.DeleteItemOutput{Attributes: item},
	})
	recorder := httptest.NewRecorder()

	handler.Destroy(recorder, requestWithUserID("user-123"))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestDestroyReturnsNotFound(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{
		deleteOutput: &dynamodb.DeleteItemOutput{},
	})
	recorder := httptest.NewRecorder()

	handler.Destroy(recorder, requestWithUserID("missing"))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDestroyRepositoryError(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{deleteErr: errors.New("DynamoDB unavailable")})
	recorder := httptest.NewRecorder()

	handler.Destroy(recorder, requestWithUserID("user-123"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestIndexReturnsAllUsers(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "user-123",
		Name:   "Ada",
		Status: "active",
	})
	handler := newTestHandler(&fakeDynamo{
		scanOutput: &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{item}},
	})
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()

	handler.Index(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var users []models.User
	if err := json.NewDecoder(recorder.Body).Decode(&users); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Ada" {
		t.Errorf("users = %+v, want one user named Ada", users)
	}
}

func TestActiveReturnsActiveUsers(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{UserID: "user-123", Name: "Ada", Status: "active"})
	handler := newTestHandler(&fakeDynamo{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}},
	})
	req := httptest.NewRequest(http.MethodGet, "/users/active", nil)
	recorder := httptest.NewRecorder()

	handler.Active(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestIndexRepositoryError(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{scanErr: errors.New("DynamoDB unavailable")})
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()

	handler.Index(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestByEmailReturnsUser(t *testing.T) {
	item, _ := attributevalue.MarshalMap(models.User{
		UserID: "user-123",
		Name:   "Ada",
		Email:  "ada@example.com",
	})
	handler := newTestHandler(&fakeDynamo{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item}},
	})
	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=ada@example.com", nil)
	recorder := httptest.NewRecorder()

	handler.ByEmail(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var user models.User
	if err := json.NewDecoder(recorder.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.Name != "Ada" {
		t.Errorf("Name = %q, want %q", user.Name, "Ada")
	}
}

func TestByEmailReturnsNotFound(t *testing.T) {
	handler := newTestHandler(&fakeDynamo{queryOutput: &dynamodb.QueryOutput{}})
	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=missing@example.com", nil)
	recorder := httptest.NewRecorder()

	handler.ByEmail(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestShowRequiresUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	recorder := httptest.NewRecorder()

	(&UserHandler{}).Show(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestByEmailReturnsConflictForDuplicates(t *testing.T) {
	first, _ := attributevalue.MarshalMap(models.User{UserID: "user-1", Email: "same@example.com"})
	second, _ := attributevalue.MarshalMap(models.User{UserID: "user-2", Email: "same@example.com"})
	handler := newTestHandler(&fakeDynamo{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{first, second}},
	})
	req := httptest.NewRequest(http.MethodGet, "/users/by-email?email=same@example.com", nil)
	recorder := httptest.NewRecorder()

	handler.ByEmail(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func TestActiveRejectsInvalidLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/active?limit=0", nil)
	recorder := httptest.NewRecorder()

	(&UserHandler{}).Active(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
