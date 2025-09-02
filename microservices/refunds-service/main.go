package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Refund struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Reason        string    `json:"reason"`
	Status        string    `json:"status"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	ProcessedAt   time.Time `json:"processed_at"`
}

func main() {
	log.Println("Starting Refunds Service on port 8093...")

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","service":"refunds","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Status endpoint
	http.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": "refunds",
			"status":  "active",
			"version": "1.0.0",
			"capabilities": []string{
				"full_refund",
				"partial_refund",
				"batch_refund",
				"instant_refund",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Create refund endpoint
	http.HandleFunc("/api/v1/refunds", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var request struct {
				TransactionID string  `json:"transaction_id"`
				Amount        float64 `json:"amount"`
				Reason        string  `json:"reason"`
				Currency      string  `json:"currency"`
			}

			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			refund := Refund{
				ID:            fmt.Sprintf("ref_%d", time.Now().Unix()),
				TransactionID: request.TransactionID,
				Amount:        request.Amount,
				Reason:        request.Reason,
				Status:        "pending",
				Currency:      request.Currency,
				CreatedAt:     time.Now(),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(refund)

		case "GET":
			// List refunds
			refunds := []Refund{
				{
					ID:            "ref_001",
					TransactionID: "txn_001",
					Amount:        50.00,
					Reason:        "Customer request",
					Status:        "completed",
					Currency:      "USD",
					CreatedAt:     time.Now().Add(-24 * time.Hour),
					ProcessedAt:   time.Now().Add(-23 * time.Hour),
				},
				{
					ID:            "ref_002",
					TransactionID: "txn_002",
					Amount:        100.00,
					Reason:        "Product defect",
					Status:        "pending",
					Currency:      "USD",
					CreatedAt:     time.Now(),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"refunds": refunds,
				"total":   len(refunds),
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Process refund endpoint
	http.HandleFunc("/api/v1/refunds/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			RefundID string `json:"refund_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		response := map[string]interface{}{
			"refund_id": request.RefundID,
			"status":    "processed",
			"message":   "Refund processed successfully",
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("Refunds service listening on port 8093")
	if err := http.ListenAndServe(":8093", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
