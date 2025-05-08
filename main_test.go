package main

import (
	"os"
	"testing"
)

func TestLoadAPIKeys(t *testing.T) {
	os.Setenv("EXCHANGERATE_API_KEY", "test_key_exchangerate")
	os.Setenv("OPENEXCHANGERATES_APP_ID", "test_key_openexchangerates")

	loadAPIKeys()

	for _, api := range APIs {
		if api.Name == "exchangerate-api" && !contains(api.BaseURL, "test_key_exchangerate") {
			t.Errorf("Expected API key for exchangerate-api not found in BaseURL")
		}
		if api.Name == "openexchangerates" && !contains(api.BaseURL, "test_key_openexchangerates") {
			t.Errorf("Expected API key for openexchangerates not found in BaseURL")
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