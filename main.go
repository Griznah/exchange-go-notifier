package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type API struct {
	Name          string
	BaseURL       string
	APIKey        string
	RequestCount  int
	RequestLimit  int
	ResetInterval time.Duration
	LastReset     time.Time
}

type ExchangeRateResponse struct {
	Rates map[string]float64 `json:"rates"`
}

var APIs = []API{
	{
		Name:          "er-a",
		BaseURL:       "https://v6.exchangerate-api.com/v6/",
		RequestLimit:  1500,                // max 1500 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
	{
		Name:          "oer",
		BaseURL:       "https://openexchangerates.org/api/",
		RequestLimit:  1000,                // max 1000 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
}

var apiStateMutex sync.Mutex

const apiStateFile = "api_state.json"

// Load API state from file
func loadAPIState() {
	apiStateMutex.Lock()
	defer apiStateMutex.Unlock()
	f, err := os.Open(apiStateFile)
	if err != nil {
		fmt.Printf("[DEBUG] Could not read %s: %v\n", apiStateFile, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Printf("[DEBUG] Could not read %s: %v\n", apiStateFile, err)
		return
	}
	var loadedAPIs []API
	if err := json.Unmarshal(data, &loadedAPIs); err != nil {
		fmt.Printf("[DEBUG] Could not unmarshal %s: %v\n", apiStateFile, err)
		return
	}
	for i := range APIs {
		for _, loaded := range loadedAPIs {
			if APIs[i].Name == loaded.Name {
				APIs[i].RequestCount = loaded.RequestCount
				APIs[i].LastReset = loaded.LastReset
			}
		}
	}
	fmt.Println("[DEBUG] Loaded API state from file.")
}

// Save API state to file
func saveAPIState() {
	apiStateMutex.Lock()
	defer apiStateMutex.Unlock()
	data, err := json.MarshalIndent(APIs, "", "  ")
	if err != nil {
		fmt.Printf("[DEBUG] Could not marshal API state: %v\n", err)
		return
	}
	if err := os.WriteFile(apiStateFile, data, 0644); err != nil {
		fmt.Printf("[DEBUG] Could not write %s: %v\n", apiStateFile, err)
		return
	}
	fmt.Println("[DEBUG] Saved API state to file.")
}

// Modify the loadAPIKeys function to store API keys separately
func loadAPIKeys() {
	for i := range APIs {
		if APIs[i].Name == "er-a" {
			APIs[i].APIKey = os.Getenv("EXCHANGERATE_API_KEY")
		} else if APIs[i].Name == "oer" {
			APIs[i].APIKey = os.Getenv("OPENEXCHANGERATES_APP_ID")
		}
	}
}

// Update the trackRequest function to handle request limit correctly
// Only checks/reset, does not increment or save
func trackRequest(api *API) error {
	currentTime := time.Now()
	firstOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())

	apiStateMutex.Lock()
	defer apiStateMutex.Unlock()
	if api.LastReset.Before(firstOfMonth) {
		fmt.Printf("[DEBUG] Resetting request count for %s\n", api.Name)
		api.RequestCount = 0
		api.LastReset = firstOfMonth
	}

	if api.RequestCount >= api.RequestLimit {
		fmt.Printf("[DEBUG] API %s has exceeded its request limit (%d/%d)\n", api.Name, api.RequestCount, api.RequestLimit)
		return fmt.Errorf("API %s has exceeded its request limit", api.Name)
	}
	return nil
}

// Update the fetchExchangeRates function
func fetchExchangeRates(api *API, baseCurrency string) (*ExchangeRateResponse, error) {
	if err := trackRequest(api); err != nil {
		fmt.Printf("[DEBUG] trackRequest error: %v\n", err)
		return nil, err
	}

	// Construct the URL with proper query parameters
	var url string
	if api.Name == "er-a" {
		url = fmt.Sprintf("%s%s/latest/%s", api.BaseURL, api.APIKey, baseCurrency)
	} else if api.Name == "oer" {
		url = fmt.Sprintf("%slatest.json?app_id=%s&base=%s", api.BaseURL, api.APIKey, baseCurrency)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("[DEBUG] HTTP GET error: %v\n", err)
		return nil, fmt.Errorf("failed to fetch exchange rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("API %s has exceeded its request limit", api.Name)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[DEBUG] API %s returned status %d\n", api.Name, resp.StatusCode)
		return nil, fmt.Errorf("API %s returned status %d", api.Name, resp.StatusCode)
	}

	var rates ExchangeRateResponse
	if api.Name == "er-a" {
		var response struct {
			ConversionRates map[string]float64 `json:"conversion_rates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			fmt.Printf("[DEBUG] JSON decode error: %v\n", err)
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
		rates.Rates = response.ConversionRates
	} else if api.Name == "oer" {
		if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
			fmt.Printf("[DEBUG] JSON decode error: %v\n", err)
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
	}

	// Only increment and save after a successful response
	apiStateMutex.Lock()
	api.RequestCount++
	fmt.Printf("[DEBUG] Incremented %s request count: %d\n", api.Name, api.RequestCount)
	apiStateMutex.Unlock()
	saveAPIState()

	fmt.Printf("[DEBUG] Successfully fetched rates: %+v\n", rates)
	return &rates, nil
}

// Helper to validate API provider
func isValidAPI(api string) bool {
	for _, a := range APIs {
		if a.Name == api {
			return true
		}
	}
	return false
}

// Helper to validate currency code (3 uppercase letters)
func isValidCurrencyCode(code string) bool {
	if len(code) != 3 {
		return false
	}
	for _, c := range code {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

// Add logging for incoming requests
func exchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %s %s\n", r.Method, r.URL.String())

	apiName := r.URL.Query().Get("api")
	baseCurrency := r.URL.Query().Get("base")
	if apiName == "" {
		http.Error(w, "Missing 'api' query parameter", http.StatusBadRequest)
		fmt.Printf("[DEBUG] Missing 'api' query parameter\n")
		return
	}
	if !isValidAPI(apiName) {
		http.Error(w, "Invalid 'api' parameter", http.StatusBadRequest)
		fmt.Printf("[DEBUG] Invalid 'api' parameter: %s\n", apiName)
		return
	}
	if baseCurrency == "" {
		baseCurrency = "USD"
	}
	if !isValidCurrencyCode(baseCurrency) {
		http.Error(w, "Invalid 'base' parameter", http.StatusBadRequest)
		fmt.Printf("[DEBUG] Invalid 'base' parameter: %s\n", baseCurrency)
		return
	}

	var selectedAPI *API
	for i := range APIs {
		if APIs[i].Name == apiName {
			selectedAPI = &APIs[i]
			break
		}
	}

	if selectedAPI == nil {
		http.Error(w, "API not found", http.StatusNotFound)
		fmt.Printf("[DEBUG] API not found: %s\n", apiName)
		return
	}

	rates, err := fetchExchangeRates(selectedAPI, baseCurrency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("[DEBUG] fetchExchangeRates error: %v\n", err)
		return
	}

	fmt.Printf("Outgoing response: %v\n", rates)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		fmt.Printf("[DEBUG] Failed to encode response: %v\n", err)
	}
}

func ensureEnvVars() {
	// Try to load .env file if required env vars are missing
	if os.Getenv("EXCHANGERATE_API_KEY") == "" || os.Getenv("OPENEXCHANGERATES_APP_ID") == "" {
		_ = godotenv.Load(".env")
	}
}

func main() {
	ensureEnvVars()
	fmt.Println("Available APIs:")
	for _, api := range APIs {
		fmt.Printf("- %s (Base URL: %s, Limit: %d requests per %v)\n", api.Name, api.BaseURL, api.RequestLimit, api.ResetInterval)
	}

	loadAPIKeys()
	loadAPIState()

	http.HandleFunc("/exchange-rates", exchangeRateHandler)
	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
