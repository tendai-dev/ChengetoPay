package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type FeeCalculation struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	FeeAmount     float64   `json:"fee_amount"`
	FeeType       string    `json:"fee_type"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

func main() {
	log.Println("Starting Fees Service on port 8092...")

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","service":"fees","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Status endpoint
	http.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": "fees",
			"status":  "active",
			"version": "1.0.0",
			"features": []string{
				"percentage_fees",
				"fixed_fees",
				"tiered_fees",
				"currency_conversion",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Calculate fee endpoint
	http.HandleFunc("/api/v1/fees/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
			Type     string  `json:"type"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Simple fee calculation (2.9% + 0.30)
		feeAmount := request.Amount*0.029 + 0.30

		response := FeeCalculation{
			ID:            fmt.Sprintf("fee_%d", time.Now().Unix()),
			TransactionID: fmt.Sprintf("txn_%d", time.Now().Unix()),
			Amount:        request.Amount,
			FeeAmount:     feeAmount,
			FeeType:       request.Type,
			Currency:      request.Currency,
			CreatedAt:     time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// List fees endpoint
	http.HandleFunc("/api/v1/fees", func(w http.ResponseWriter, r *http.Request) {
		fees := []FeeCalculation{
			{
				ID:            "fee_001",
				TransactionID: "txn_001",
				Amount:        100.00,
				FeeAmount:     3.20,
				FeeType:       "standard",
				Currency:      "USD",
				CreatedAt:     time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"fees":  fees,
			"total": len(fees),
		})
	})

	log.Println("Fees service listening on port 8092")
	if err := http.ListenAndServe(":8092", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
