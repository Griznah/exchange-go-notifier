package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		BaseURL:       "https://v6.exchangerate-api.com/v6/APIKEY/latest/CURRENCY",
		RequestLimit:  1000,
		ResetInterval: 24 * time.Hour,
		LastReset:     time.Now(),
	},
	{
		Name:          "API2",
		BaseURL:       "https://api2.example.com/rates",
		RequestLimit:  500,
		ResetInterval: 24 * time.Hour,
		LastReset:     time.Now(),
	},
}

var mu sync.Mutex

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

	url := fmt.Sprintf("%s?base=%s", api.BaseURL, baseCurrency)
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

	return &rates, nil
}

func exchangeRateHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rates)
}

func main() {
	fmt.Println("Available APIs:")
	for _, api := range APIs {
		fmt.Printf("- %s (Base URL: %s, Limit: %d requests per %v)\n", api.Name, api.BaseURL, api.RequestLimit, api.ResetInterval)
	}

	http.HandleFunc("/exchange-rates", exchangeRateHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}