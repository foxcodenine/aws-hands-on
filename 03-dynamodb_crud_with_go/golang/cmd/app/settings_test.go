package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	t.Run("loads values from an existing .env file", func(t *testing.T) {
		// t.TempDir() is auto-removed when the test finishes.
		dir := t.TempDir()
		writeEnvFile(t, dir, "GREETING=hello\n")

		err := loadEnv(dir)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := os.Getenv("GREETING"); got != "hello" {
			t.Errorf("GREETING = %q, want %q", got, "hello")
		}
		if got := os.Getenv("TZ"); got != "UTC" {
			t.Errorf("TZ = %q, want %q", got, "UTC")
		}
	})

	t.Run("returns error when file is missing", func(t *testing.T) {
		err := loadEnv(filepath.Join(t.TempDir(), "does-not-exist"))

		if err == nil {
			t.Fatal("expected an error for a missing .env, got nil")
		}
	})
}

func writeEnvFile(t *testing.T, dir, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600)

	if err != nil {
		t.Fatalf("failed to write test .env: %v", err)
	}
}
