package db

import (
	"context"
	"io"
	"testing"
)

func TestGetRegion(t *testing.T) {
	t.Run("uses AWS_REGION when set", func(t *testing.T) {
		t.Setenv("AWS_REGION", "us-east-1")

		got := getRegion()

		if got != "us-east-1" {
			t.Errorf("got %q, want %q", got, "us-east-1")
		}

	})

	t.Run("falls back to eu-west-1 when unset", func(t *testing.T) {
		t.Setenv("AWS_REGION", "")

		got := getRegion()

		if got != "eu-west-1" {
			t.Errorf("got %q, want %q", got, "eu-west-1")
		}
	})
}

func TestNewClient(t *testing.T) {
	// Smoke test: the client should build without panicking and never be nil,
	// both on the default-AWS path and the local-endpoint override path.
	t.Run("default endpoint", func(t *testing.T) {
		t.Setenv("DYNAMODB_ENDPOINT", "")

		got := NewClient(io.Discard, context.Background())

		if got == nil {
			t.Fatal("NewClient returned nil")
		}
	})

	t.Run("endpoint override", func(t *testing.T) {
		t.Setenv("DYNAMODB_ENDPOINT", "http://localhost:8000")

		got := NewClient(io.Discard, context.Background())

		if got == nil {
			t.Fatal("NewClient returned nil")
		}
	})
}
