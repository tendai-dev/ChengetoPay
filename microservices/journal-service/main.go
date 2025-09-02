package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("Starting Journal Service on port 8091...")
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"healthy","service":"journal","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})
	
	log.Fatal(http.ListenAndServe(":8091", nil))
}
