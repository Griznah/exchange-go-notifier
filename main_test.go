package main

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestLoadAPIKeys(t *testing.T) {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	exchangerateAPIKey := os.Getenv("EXCHANGERATE_API_KEY")
	openexchangeratesAppID := os.Getenv("OPENEXCHANGERATES_APP_ID")

	loadAPIKeys()

	for _, api := range APIs {
		if api.Name == "exchangerate-api" && !contains(api.BaseURL, exchangerateAPIKey) {
			t.Errorf("Expected API key for exchangerate-api not found in BaseURL")
		}
		if api.Name == "openexchangerates" && !contains(api.BaseURL, openexchangeratesAppID) {
			t.Errorf("Expected APP ID for openexchangerates not found in BaseURL")
		}
	}
}

func TestSaveAndLoadAPIState(t *testing.T) {
	// Modify APIs for testing
	APIs[0].RequestCount = 10
	APIs[1].RequestCount = 20

	err := saveAPIState()
	if err != nil {
		t.Fatalf("Failed to save API state: %v", err)
	}

	// Reset APIs to simulate loading
	APIs[0].RequestCount = 0
	APIs[1].RequestCount = 0

	err = loadAPIState()
	if err != nil {
		t.Fatalf("Failed to load API state: %v", err)
	}

	if APIs[0].RequestCount != 10 {
		t.Errorf("Expected RequestCount 10 for API 0, got %d", APIs[0].RequestCount)
	}
	if APIs[1].RequestCount != 20 {
		t.Errorf("Expected RequestCount 20 for API 1, got %d", APIs[1].RequestCount)
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[len(str)-len(substr):] == substr
}