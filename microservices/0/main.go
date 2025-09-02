package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.String("port", "8114", "Port to listen on")
	flag.Parse()

	serviceName := "0"
	log.Printf("Starting %s Service on port %s...", serviceName, *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"%s","timestamp":"%s"}`, serviceName, time.Now().Format(time.RFC3339))
	})
	
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service": serviceName,
			"status": "active",
			"version": "1.0.0",
			"port": *port,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	server := &http.Server{
		Addr:           ":" + *port,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("%s service listening on port %s", serviceName, *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Shutting down %s service...", serviceName)
	log.Printf("%s service exited", serviceName)
}
