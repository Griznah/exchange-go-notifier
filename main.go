package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type API struct {
	Name          string
	BaseURL       string
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
		Name:          "exchangerate-api",
		BaseURL:       "https://v6.exchangerate-api.com/v6/",
		RequestLimit:  1500, // max 1500 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
	{
		Name:          "openexchangerates",
		BaseURL:       "https://openexchangerates.org/api/",
		RequestLimit:  1000, // max 1000 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
}

var mu sync.Mutex

// Load API keys from environment variables
func loadAPIKeys() {
	for i := range APIs {
		if APIs[i].Name == "exchangerate-api" {
			APIs[i].BaseURL += os.Getenv("EXCHANGERATE_API_KEY")
		} else if APIs[i].Name == "openexchangerates" {
			APIs[i].BaseURL += os.Getenv("OPENEXCHANGERATES_APP_ID")
		}
	}
}

// Save API state to a file
func saveAPIState() error {
	file, err := os.Create("api_state.json")
	if err != nil {
		return err
	}
	defer file.Close()

	mu.Lock()
	defer mu.Unlock()

	return json.NewEncoder(file).Encode(APIs)
}

// Load API state from a file
func loadAPIState() error {
	file, err := os.Open("api_state.json")
	if err != nil {
		return err
	}
	defer file.Close()

	mu.Lock()
	defer mu.Unlock()

	return json.NewDecoder(file).Decode(&APIs)
}

func trackRequest(api *API) error {
	if time.Since(api.LastReset) > api.ResetInterval {
		api.RequestCount = 0
		api.LastReset = time.Now()
	}

	if api.RequestCount >= api.RequestLimit {
		return fmt.Errorf("API %s has exceeded its request limit", api.Name)
	}

	api.RequestCount++
	return nil
}

func fetchExchangeRates(api *API, baseCurrency string) (*ExchangeRateResponse, error) {
	mu.Lock()
	err := trackRequest(api)
	if err != nil {
		mu.Unlock()
		return nil, err
	}
	mu.Unlock()

	// Construct the URL with proper query parameters
	var url string
	if api.Name == "exchangerate-api" {
		url = fmt.Sprintf("%s/latest/%s", api.BaseURL, baseCurrency)
	} else if api.Name == "openexchangerates" {
		url = fmt.Sprintf("%s/latest.json?app_id=%s&base=%s", api.BaseURL, os.Getenv("OPENEXCHANGERATES_APP_ID"), baseCurrency)
	}

	fmt.Printf("Outgoing API call: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API %s returned status %d", api.Name, resp.StatusCode)
	}

	var rates ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("API response: %v\n", rates)
	return &rates, nil
}

// Add logging for incoming requests
func exchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %s %s\n", r.Method, r.URL.String())
	
	apiName := r.URL.Query().Get("api")
	baseCurrency := r.URL.Query().Get("base")
	if apiName == "" || baseCurrency == "" {
		http.Error(w, "Missing 'api' or 'base' query parameter", http.StatusBadRequest)
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
		return
	}

	rates, err := fetchExchangeRates(selectedAPI, baseCurrency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Outgoing response: %v\n", rates)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

func main() {
	fmt.Println("Available APIs:")
	for _, api := range APIs {
		fmt.Printf("- %s (Base URL: %s, Limit: %d requests per %v)\n", api.Name, api.BaseURL, api.RequestLimit, api.ResetInterval)
	}

	loadAPIKeys()
	if err := loadAPIState(); err != nil {
		fmt.Println("Failed to load API state:", err)
	}

	http.HandleFunc("/exchange-rates", exchangeRateHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)

	if err := saveAPIState(); err != nil {
		fmt.Println("Failed to save API state:", err)
	}
}