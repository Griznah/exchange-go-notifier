package main

import (
	"log"
	"os"
	"testing"
	"time"

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
		if api.Name == "exchangerate-api" && api.APIKey != exchangerateAPIKey {
			t.Errorf("Expected API key for exchangerate-api to be %s, got %s", exchangerateAPIKey, api.APIKey)
		}
		if api.Name == "openexchangerates" && api.APIKey != openexchangeratesAppID {
			t.Errorf("Expected APP ID for openexchangerates to be %s, got %s", openexchangeratesAppID, api.APIKey)
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

func TestTrackRequest(t *testing.T) {
	// Set up test APIs
	APIs[0].RequestCount = 0
	APIs[0].LastReset = time.Date(2025, 4, 1, 0, 0, 0, 0, time.Local) // Last reset in April
	APIs[1].RequestCount = 0
	APIs[1].LastReset = time.Date(2025, 4, 1, 0, 0, 0, 0, time.Local)

	// Test reset logic
	err := trackRequest(&APIs[0])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if APIs[0].RequestCount != 1 {
		t.Errorf("Expected RequestCount 1 for API 0, got %d", APIs[0].RequestCount)
	}
	if !APIs[0].LastReset.Equal(time.Date(2025, 5, 1, 0, 0, 0, 0, time.Local)) {
		t.Errorf("Expected LastReset to be updated to May 1, 2025, got %v", APIs[0].LastReset)
	}

	// Test exceeding request limit
	APIs[1].RequestCount = APIs[1].RequestLimit
	err = trackRequest(&APIs[1])
	if err == nil {
		t.Fatalf("Expected error for exceeding request limit, got nil")
	}
}

func TestFetchExchangeRates(t *testing.T) {
	// Mock API state
	APIs[0].RequestCount = 0
	APIs[1].RequestCount = 0

	// Mock environment variables
	os.Setenv("EXCHANGERATE_API_KEY", "testkey1")
	os.Setenv("OPENEXCHANGERATES_APP_ID", "testkey2")
	loadAPIKeys()

	// Test exchangerate-api
	_, err := fetchExchangeRates(&APIs[0], "USD")
	if err != nil {
		t.Errorf("Unexpected error for exchangerate-api: %v", err)
	}
	if APIs[0].RequestCount != 1 {
		t.Errorf("Expected RequestCount 1 for API 0, got %d", APIs[0].RequestCount)
	}

	// Test openexchangerates
	_, err = fetchExchangeRates(&APIs[1], "USD")
	if err != nil {
		t.Errorf("Unexpected error for openexchangerates: %v", err)
	}
	if APIs[1].RequestCount != 1 {
		t.Errorf("Expected RequestCount 1 for API 1, got %d", APIs[1].RequestCount)
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[len(str)-len(substr):] == substr
}