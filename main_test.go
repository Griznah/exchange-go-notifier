package main

import (
	"os"
	"testing"
	"time"
)

func TestLoadAPIKeys(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("EXCHANGERATE_API_KEY", "test_exchangerate_api_key")
	os.Setenv("OPENEXCHANGERATES_APP_ID", "test_openexchangerates_app_id")
	defer os.Unsetenv("EXCHANGERATE_API_KEY")
	defer os.Unsetenv("OPENEXCHANGERATES_APP_ID")

	// Call the function to load API keys
	loadAPIKeys()

	// Verify that the API keys are correctly loaded
	for _, api := range APIs {
		if api.Name == "exchangerate-api" && api.APIKey != "test_exchangerate_api_key" {
			t.Errorf("Expected API key for exchangerate-api to be 'test_exchangerate_api_key', got '%s'", api.APIKey)
		}
		if api.Name == "openexchangerates" && api.APIKey != "test_openexchangerates_app_id" {
			t.Errorf("Expected API key for openexchangerates to be 'test_openexchangerates_app_id', got '%s'", api.APIKey)
		}
	}
}

func TestAPIsInitialization(t *testing.T) {
	// Verify that the APIs are initialized with correct values
	if len(APIs) != 2 {
		t.Fatalf("Expected 2 APIs, got %d", len(APIs))
	}

	tests := []struct {
		name          string
		expectedBase  string
		expectedLimit int
	}{
		{"exchangerate-api", "https://v6.exchangerate-api.com/v6/", 1500},
		{"openexchangerates", "https://openexchangerates.org/api/", 1000},
	}

	for _, test := range tests {
		found := false
		for _, api := range APIs {
			if api.Name == test.name {
				found = true
				if api.BaseURL != test.expectedBase {
					t.Errorf("Expected BaseURL for %s to be '%s', got '%s'", test.name, test.expectedBase, api.BaseURL)
				}
				if api.RequestLimit != test.expectedLimit {
					t.Errorf("Expected RequestLimit for %s to be %d, got %d", test.name, test.expectedLimit, api.RequestLimit)
				}
				if api.ResetInterval != 30*24*time.Hour {
					t.Errorf("Expected ResetInterval for %s to be 30 days, got %v", test.name, api.ResetInterval)
				}
			}
		}
		if !found {
			t.Errorf("API %s not found in APIs list", test.name)
		}
	}
}
