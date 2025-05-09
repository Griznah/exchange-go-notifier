package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	//"os/signal"
	//"syscall"
	"time"
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
		Name:          "exchangerate-api",
		BaseURL:       "https://v6.exchangerate-api.com/v6/",
		RequestLimit:  1500,                // max 1500 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
	{
		Name:          "openexchangerates",
		BaseURL:       "https://openexchangerates.org/api/",
		RequestLimit:  1000,                // max 1000 requests per month
		ResetInterval: 30 * 24 * time.Hour, // monthly reset
		LastReset:     time.Now(),
	},
}

var mu sync.Mutex

// Modify the loadAPIKeys function to store API keys separately
func loadAPIKeys() {
	for i := range APIs {
		if APIs[i].Name == "exchangerate-api" {
			APIs[i].APIKey = os.Getenv("EXCHANGERATE_API_KEY")
		} else if APIs[i].Name == "openexchangerates" {
			APIs[i].APIKey = os.Getenv("OPENEXCHANGERATES_APP_ID")
		}
	}
}

// Save API state to a file
func saveAPIState() error {
	tempFile, err := os.CreateTemp("", "api_state_*.json")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	mu.Lock()
	defer mu.Unlock()

	if err := json.NewEncoder(tempFile).Encode(APIs); err != nil {
		return fmt.Errorf("failed to encode API state: %v", err)
	}

	// Rename the temporary file to the actual file
	if err := os.Rename(tempFile.Name(), "api_state.json"); err != nil {
		return fmt.Errorf("failed to rename temporary file: %v", err)
	}

	return nil
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

// Update the trackRequest function to handle request limit correctly
func trackRequest(api *API) error {
	currentTime := time.Now()
	firstOfMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, currentTime.Location())

	if api.LastReset.Before(firstOfMonth) {
		api.RequestCount = 0
		api.LastReset = firstOfMonth
	}

	if api.RequestCount >= api.RequestLimit {
		return fmt.Errorf("API %s has exceeded its request limit", api.Name)
	}

	api.RequestCount++
	if err := saveAPIState(); err != nil {
		return fmt.Errorf("failed to save API state: %v", err)
	}

	return nil
}

// Update the fetchExchangeRates function to save state only on successful calls
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
		url = fmt.Sprintf("%s%s/latest/%s", api.BaseURL, api.APIKey, baseCurrency)
	} else if api.Name == "openexchangerates" {
		url = fmt.Sprintf("%slatest.json?app_id=%s", api.BaseURL, api.APIKey)
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

	// Save state only on successful response
	if err := saveAPIState(); err != nil {
		return nil, fmt.Errorf("failed to save API state: %v", err)
	}

	var rates ExchangeRateResponse
	if api.Name == "exchangerate-api" {
		var response struct {
			ConversionRates map[string]float64 `json:"conversion_rates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
		rates.Rates = response.ConversionRates
	} else if api.Name == "openexchangerates" {
		if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
			return nil, fmt.Errorf("failed to decode response: %v", err)
		}
	}

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
	if err := json.NewEncoder(w).Encode(rates); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
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

	// Add a function to periodically save API state
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := saveAPIState(); err != nil {
				fmt.Printf("Failed to save API state: %v\n", err)
			}
		}
	}()

	// Ensure API state is saved on application exit
	/* 	c := make(chan os.Signal, 1)
	   	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	   	go func() {
	   		<-c
	   		fmt.Println("Saving API state before exit...")
	   		if err := saveAPIState(); err != nil {
	   			fmt.Printf("Failed to save API state on exit: %v\n", err)
	   		}
	   		os.Exit(0)
	   	}() */

	http.HandleFunc("/exchange-rates", exchangeRateHandler)
	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}

	if err := saveAPIState(); err != nil {
		fmt.Println("Failed to save API state:", err)
	}
}
