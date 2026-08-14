package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadAPIKeys(t *testing.T) {
	t.Setenv("EXCHANGERATE_API_KEY", "test_exchangerate_api_key")
	t.Setenv("OPENEXCHANGERATES_APP_ID", "test_openexchangerates_app_id")

	// Deep copy: loadAPIKeys mutates elements in place, and restoring the slice
	// header would leave the mutated values in the shared backing array.
	originalAPIs := append([]API(nil), APIs...)
	defer func() { APIs = originalAPIs }()
	loadAPIKeys()

	for _, api := range APIs {
		switch api.Name {
		case "er-a":
			if api.APIKey != "test_exchangerate_api_key" {
				t.Errorf("Expected API key for er-a to be 'test_exchangerate_api_key', got '%s'", api.APIKey)
			}
		case "oer":
			if api.APIKey != "test_openexchangerates_app_id" {
				t.Errorf("Expected API key for oer to be 'test_openexchangerates_app_id', got '%s'", api.APIKey)
			}
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
	// Create a test-specific API slice
	testAPI := API{
		Name:         "testapi",
		RequestLimit: 2,
		LastReset:    time.Now().Add(-31 * 24 * time.Hour), // force reset
	}
	// Save original and create test slice
	originalAPIs := APIs
	testAPIs := append([]API{}, testAPI)
	APIs = testAPIs
	// Ensure cleanup after test
	defer func() {
		APIs = originalAPIs
	}()
	api := &APIs[0]

	// Should reset and allow request
	if err := trackRequest(api); err != nil {
		t.Errorf("Expected no error on first request after reset, got %v", err)
	}
	api.RequestCount++ // Simulate increment after success

	// Should allow one more
	if err := trackRequest(api); err != nil {
		t.Errorf("Expected no error on second request, got %v", err)
	}
	api.RequestCount++ // Simulate increment after success

	// Should block further requests
	if err := trackRequest(api); err == nil {
		t.Errorf("Expected error after exceeding limit, got nil")
	}
	// One more call to ensure error is returned after limit is reached
	if err := trackRequest(api); err == nil {
		t.Errorf("Expected error after exceeding limit (second check), got nil")
	}
}

func TestSaveAndLoadAPIState(t *testing.T) {
	origFile, origAPIs := apiStateFile, APIs
	defer func() { apiStateFile, APIs = origFile, origAPIs }()
	apiStateFile = filepath.Join(t.TempDir(), "api_state.json")
	APIs = []API{{Name: "persistapi", RequestLimit: 10, RequestCount: 5, LastReset: time.Now()}}

	saveAPIState()

	APIs[0].RequestCount = 0
	loadAPIState()

	if APIs[0].RequestCount != 5 {
		t.Errorf("Expected RequestCount to be 5 after reload, got %d", APIs[0].RequestCount)
	}
}

// --- Input validation tests ---
func TestIsValidAPI(t *testing.T) {
	if !isValidAPI("er-a") {
		t.Errorf("'er-a' should be a valid API")
	}
	if !isValidAPI("oer") {
		t.Errorf("'oer' should be a valid API")
	}
	if isValidAPI("invalid") {
		t.Errorf("'invalid' should not be a valid API")
	}
}

func TestIsValidCurrencyCode(t *testing.T) {
	valid := []string{"USD", "EUR", "JPY"}
	invalid := []string{"usd", "US", "USDE", "12A", "Eur", "", "123", "US$"}
	for _, code := range valid {
		if !isValidCurrencyCode(code) {
			t.Errorf("'%s' should be valid currency code", code)
		}
	}
	for _, code := range invalid {
		if isValidCurrencyCode(code) {
			t.Errorf("'%s' should be invalid currency code", code)
		}
	}
}

// --- HTTP handler validation tests ---
func TestExchangeRateHandler_InputValidation(t *testing.T) {
	// Route er-a at a mock provider so the happy-path case stays offline and
	// deterministic (403 maps to the "exceeded its request limit" error).
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer mock.Close()
	// Deep copy so the BaseURL mutations below don't leak into other tests via
	// the shared backing array (slice-header restore alone doesn't undo them).
	origAPIs := append([]API(nil), APIs...)
	defer func() { APIs = origAPIs }()
	for i := range APIs {
		if APIs[i].Name == "er-a" {
			APIs[i].BaseURL = mock.URL + "/"
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(exchangeRateHandler))
	defer ts.Close()

	cases := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{"missing api", "", 400, "Missing 'api' query parameter"},
		{"invalid api", "api=invalid", 400, "Invalid 'api' parameter"},
		{"invalid base", "api=er-a&base=usd", 400, "Invalid 'base' parameter"},
		{"invalid base 2", "api=er-a&base=USDE", 400, "Invalid 'base' parameter"},
		{"valid api and base", "api=er-a&base=USD", 500, "API er-a has exceeded its request limit"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + "/exchange-rates?" + tc.query)
			if err != nil {
				t.Fatalf("http.Get failed: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("Expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
			if tc.wantBody != "" && !strings.Contains(string(body), tc.wantBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tc.wantBody, string(body))
			}
		})
	}
}
