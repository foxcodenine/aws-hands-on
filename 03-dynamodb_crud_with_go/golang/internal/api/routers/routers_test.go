package routers

import (
	apptypes "03-dynamodb_crud_with_go/internal/app"
	"03-dynamodb_crud_with_go/internal/repository"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type fakeDynamo struct{}

func (fakeDynamo) PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, nil
}

func (fakeDynamo) GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{}, nil
}

func (fakeDynamo) UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return &dynamodb.UpdateItemOutput{}, nil
}

func (fakeDynamo) DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
}

func (fakeDynamo) Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, nil
}

func (fakeDynamo) Scan(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{}, nil
}

func TestUserRoutes(t *testing.T) {
	app := apptypes.App{Repo: repository.NewRepository(fakeDynamo{})}
	router := Routes(app)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{name: "store", method: http.MethodPost, path: "/users/", body: "{", status: http.StatusBadRequest},
		{name: "index", method: http.MethodGet, path: "/users/", status: http.StatusOK},
		{name: "active", method: http.MethodGet, path: "/users/active", status: http.StatusOK},
		{name: "by email validation", method: http.MethodGet, path: "/users/by-email", status: http.StatusBadRequest},
		{name: "show", method: http.MethodGet, path: "/users/user-123", status: http.StatusNotFound},
		{name: "update", method: http.MethodPut, path: "/users/user-123", body: "{", status: http.StatusBadRequest},
		{name: "destroy", method: http.MethodDelete, path: "/users/user-123", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}
		})
	}
}
