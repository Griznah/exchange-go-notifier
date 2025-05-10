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

	// Use new API names
	APIs[0].Name = "er-a"
	APIs[1].Name = "oer"
	loadAPIKeys()

	// Verify that the API keys are correctly loaded
	for _, api := range APIs {
		if api.Name == "er-a" && api.APIKey != "test_exchangerate_api_key" {
			t.Errorf("Expected API key for er-a to be 'test_exchangerate_api_key', got '%s'", api.APIKey)
		}
		if api.Name == "oer" && api.APIKey != "test_openexchangerates_app_id" {
			t.Errorf("Expected API key for oer to be 'test_openexchangerates_app_id', got '%s'", api.APIKey)
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
		{"er-a", "https://v6.exchangerate-api.com/v6/", 1500},
		{"oer", "https://openexchangerates.org/api/", 1000},
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

func TestTrackRequestAndStatePersistence(t *testing.T) {
	api := &API{
		Name:         "testapi",
		RequestLimit: 2,
		LastReset:    time.Now().Add(-31 * 24 * time.Hour), // force reset
	}
	APIs = append(APIs, *api)
	// Should reset and allow request
	if err := trackRequest(api); err != nil {
		t.Errorf("Expected no error on first request after reset, got %v", err)
	}
	// Simulate increment and save
	api.RequestCount++
	saveAPIState()
	// Should allow one more
	if err := trackRequest(api); err != nil {
		t.Errorf("Expected no error on second request, got %v", err)
	}
	api.RequestCount++
	saveAPIState()
	// Should block further requests
	if err := trackRequest(api); err == nil {
		t.Errorf("Expected error after exceeding limit, got nil")
	}
	// Clean up
	APIs = APIs[:len(APIs)-1]
}

func TestSaveAndLoadAPIState(t *testing.T) {
	api := &API{
		Name:         "persistapi",
		RequestLimit: 10,
		RequestCount: 5,
		LastReset:    time.Now(),
	}
	APIs = append(APIs, *api)
	saveAPIState()
	// Zero out and reload
	APIs[len(APIs)-1].RequestCount = 0
	loadAPIState()
	if APIs[len(APIs)-1].RequestCount != 5 {
		t.Errorf("Expected RequestCount to be 5 after reload, got %d", APIs[len(APIs)-1].RequestCount)
	}
	// Clean up
	APIs = APIs[:len(APIs)-1]
}
