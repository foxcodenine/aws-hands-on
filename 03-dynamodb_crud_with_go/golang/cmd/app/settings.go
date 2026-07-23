package main

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var envPath = "./"

func loadEnv(envPath string) error {
	os.Setenv("TZ", "UTC")

	fullPath := filepath.Join(envPath, ".env")

	return godotenv.Load(fullPath)
}
