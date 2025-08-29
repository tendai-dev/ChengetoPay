package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// LoadTestConfig holds load testing configuration
type LoadTestConfig struct {
	BaseURL         string
	TotalRequests   int
	ConcurrentUsers int
	RampUpDuration  time.Duration
	TestDuration    time.Duration
	RequestsPerSec  int
}

// TestScenario represents a load test scenario
type TestScenario struct {
	Name        string
	Weight      int
	RequestFunc func(*http.Client, string) (*http.Response, error)
}

// LoadTestResults holds test results
type LoadTestResults struct {
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedRequests    int64
	TotalDuration     time.Duration
	AverageLatency    time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	RequestsPerSecond float64
	ErrorRate         float64
	StatusCodes       map[int]int64
	Latencies         []time.Duration
}

// LoadTester performs load testing
type LoadTester struct {
	config    LoadTestConfig
	scenarios []TestScenario
	client    *http.Client
	results   *LoadTestResults
	mu        sync.RWMutex
}

// NewLoadTester creates a new load tester
func NewLoadTester(config LoadTestConfig) *LoadTester {
	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		results: &LoadTestResults{
			StatusCodes: make(map[int]int64),
			Latencies:   make([]time.Duration, 0),
		},
	}
}

// AddScenario adds a test scenario
func (lt *LoadTester) AddScenario(scenario TestScenario) {
	lt.scenarios = append(lt.scenarios, scenario)
}

// RunLoadTest executes the load test
func (lt *LoadTester) RunLoadTest(ctx context.Context) *LoadTestResults {
	log.Printf("Starting load test with %d concurrent users for %v", 
		lt.config.ConcurrentUsers, lt.config.TestDuration)

	startTime := time.Now()
	
	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(lt.config.RequestsPerSec), lt.config.RequestsPerSec)
	
	// Create worker pool
	var wg sync.WaitGroup
	requestChan := make(chan TestScenario, lt.config.ConcurrentUsers*10)
	
	// Start workers
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go lt.worker(ctx, &wg, requestChan, limiter)
	}
	
	// Generate load
	go lt.generateLoad(ctx, requestChan)
	
	// Wait for test duration or context cancellation
	select {
	case <-ctx.Done():
		log.Println("Load test cancelled")
	case <-time.After(lt.config.TestDuration):
		log.Println("Load test duration completed")
	}
	
	close(requestChan)
	wg.Wait()
	
	lt.results.TotalDuration = time.Since(startTime)
	lt.calculateResults()
	
	return lt.results
}

// worker processes requests
func (lt *LoadTester) worker(ctx context.Context, wg *sync.WaitGroup, requestChan <-chan TestScenario, limiter *rate.Limiter) {
	defer wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			return
		case scenario, ok := <-requestChan:
			if !ok {
				return
			}
			
			// Wait for rate limiter
			if err := limiter.Wait(ctx); err != nil {
				return
			}
			
			lt.executeRequest(scenario)
		}
	}
}

// generateLoad generates load according to scenarios
func (lt *LoadTester) generateLoad(ctx context.Context, requestChan chan<- TestScenario) {
	ticker := time.NewTicker(time.Second / time.Duration(lt.config.RequestsPerSec))
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scenario := lt.selectScenario()
			select {
			case requestChan <- scenario:
			case <-ctx.Done():
				return
			default:
				// Channel full, skip this request
			}
		}
	}
}

// selectScenario selects a scenario based on weights
func (lt *LoadTester) selectScenario() TestScenario {
	totalWeight := 0
	for _, scenario := range lt.scenarios {
		totalWeight += scenario.Weight
	}
	
	r := rand.Intn(totalWeight)
	currentWeight := 0
	
	for _, scenario := range lt.scenarios {
		currentWeight += scenario.Weight
		if r < currentWeight {
			return scenario
		}
	}
	
	return lt.scenarios[0] // fallback
}

// executeRequest executes a single request
func (lt *LoadTester) executeRequest(scenario TestScenario) {
	start := time.Now()
	
	resp, err := scenario.RequestFunc(lt.client, lt.config.BaseURL)
	latency := time.Since(start)
	
	atomic.AddInt64(&lt.results.TotalRequests, 1)
	
	lt.mu.Lock()
	lt.results.Latencies = append(lt.results.Latencies, latency)
	lt.mu.Unlock()
	
	if err != nil {
		atomic.AddInt64(&lt.results.FailedRequests, 1)
		log.Printf("Request failed: %v", err)
		return
	}
	
	if resp != nil {
		resp.Body.Close()
		atomic.AddInt64(&lt.results.StatusCodes[resp.StatusCode], 1)
		
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			atomic.AddInt64(&lt.results.SuccessfulReqs, 1)
		} else {
			atomic.AddInt64(&lt.results.FailedRequests, 1)
		}
	}
}

// calculateResults calculates final test results
func (lt *LoadTester) calculateResults() {
	if len(lt.results.Latencies) == 0 {
		return
	}
	
	// Calculate latency statistics
	var totalLatency time.Duration
	lt.results.MinLatency = lt.results.Latencies[0]
	lt.results.MaxLatency = lt.results.Latencies[0]
	
	for _, latency := range lt.results.Latencies {
		totalLatency += latency
		if latency < lt.results.MinLatency {
			lt.results.MinLatency = latency
		}
		if latency > lt.results.MaxLatency {
			lt.results.MaxLatency = latency
		}
	}
	
	lt.results.AverageLatency = totalLatency / time.Duration(len(lt.results.Latencies))
	lt.results.RequestsPerSecond = float64(lt.results.TotalRequests) / lt.results.TotalDuration.Seconds()
	
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedRequests) / float64(lt.results.TotalRequests) * 100
	}
}

// PrintResults prints test results
func (lt *LoadTester) PrintResults() {
	results := lt.results
	
	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Total Requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", results.SuccessfulReqs)
	fmt.Printf("Failed Requests: %d\n", results.FailedRequests)
	fmt.Printf("Test Duration: %v\n", results.TotalDuration)
	fmt.Printf("Requests/Second: %.2f\n", results.RequestsPerSecond)
	fmt.Printf("Error Rate: %.2f%%\n", results.ErrorRate)
	fmt.Printf("Average Latency: %v\n", results.AverageLatency)
	fmt.Printf("Min Latency: %v\n", results.MinLatency)
	fmt.Printf("Max Latency: %v\n", results.MaxLatency)
	
	fmt.Printf("\nStatus Code Distribution:\n")
	for code, count := range results.StatusCodes {
		fmt.Printf("  %d: %d\n", code, count)
	}
	
	// Calculate percentiles
	if len(results.Latencies) > 0 {
		fmt.Printf("\nLatency Percentiles:\n")
		fmt.Printf("  50th: %v\n", lt.getPercentile(50))
		fmt.Printf("  90th: %v\n", lt.getPercentile(90))
		fmt.Printf("  95th: %v\n", lt.getPercentile(95))
		fmt.Printf("  99th: %v\n", lt.getPercentile(99))
	}
}

// getPercentile calculates latency percentile
func (lt *LoadTester) getPercentile(percentile int) time.Duration {
	if len(lt.results.Latencies) == 0 {
		return 0
	}
	
	// Sort latencies (simple bubble sort for small datasets)
	latencies := make([]time.Duration, len(lt.results.Latencies))
	copy(latencies, lt.results.Latencies)
	
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
	
	index := (percentile * len(latencies)) / 100
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	
	return latencies[index]
}

// Test scenarios
func createEscrowRequest(client *http.Client, baseURL string) (*http.Response, error) {
	data := map[string]interface{}{
		"buyer_id":  fmt.Sprintf("buyer-%d", rand.Intn(1000)),
		"seller_id": fmt.Sprintf("seller-%d", rand.Intn(1000)),
		"amount":    map[string]interface{}{"value": rand.Float64() * 1000, "currency": "USD"},
		"terms":     "Load test escrow",
	}
	
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/escrow/v1/escrows", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer load-test-token")
	
	return client.Do(req)
}

func createPaymentRequest(client *http.Client, baseURL string) (*http.Response, error) {
	data := map[string]interface{}{
		"amount":         map[string]interface{}{"value": rand.Float64() * 500, "currency": "USD"},
		"payment_method": "credit_card",
		"provider":       "stripe",
		"description":    "Load test payment",
	}
	
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/payment/v1/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer load-test-token")
	
	return client.Do(req)
}

func healthCheckRequest(client *http.Client, baseURL string) (*http.Response, error) {
	return client.Get(baseURL + "/health")
}

func getRootRequest(client *http.Client, baseURL string) (*http.Response, error) {
	return client.Get(baseURL + "/")
}

func main() {
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8090",
		TotalRequests:   10000,
		ConcurrentUsers: 50,
		TestDuration:    2 * time.Minute,
		RequestsPerSec:  100,
	}
	
	tester := NewLoadTester(config)
	
	// Add test scenarios
	tester.AddScenario(TestScenario{
		Name:        "Create Escrow",
		Weight:      30,
		RequestFunc: createEscrowRequest,
	})
	
	tester.AddScenario(TestScenario{
		Name:        "Create Payment",
		Weight:      30,
		RequestFunc: createPaymentRequest,
	})
	
	tester.AddScenario(TestScenario{
		Name:        "Health Check",
		Weight:      20,
		RequestFunc: healthCheckRequest,
	})
	
	tester.AddScenario(TestScenario{
		Name:        "Get Root",
		Weight:      20,
		RequestFunc: getRootRequest,
	})
	
	// Run load test
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	results := tester.RunLoadTest(ctx)
	tester.PrintResults()
	
	// Check performance targets
	fmt.Printf("\n=== Performance Analysis ===\n")
	
	targetRPS := 100.0
	targetLatency := 20 * time.Millisecond
	targetErrorRate := 1.0
	
	if results.RequestsPerSecond >= targetRPS {
		fmt.Printf("✅ RPS Target: %.2f >= %.2f\n", results.RequestsPerSecond, targetRPS)
	} else {
		fmt.Printf("❌ RPS Target: %.2f < %.2f\n", results.RequestsPerSecond, targetRPS)
	}
	
	if results.AverageLatency <= targetLatency {
		fmt.Printf("✅ Latency Target: %v <= %v\n", results.AverageLatency, targetLatency)
	} else {
		fmt.Printf("❌ Latency Target: %v > %v\n", results.AverageLatency, targetLatency)
	}
	
	if results.ErrorRate <= targetErrorRate {
		fmt.Printf("✅ Error Rate Target: %.2f%% <= %.2f%%\n", results.ErrorRate, targetErrorRate)
	} else {
		fmt.Printf("❌ Error Rate Target: %.2f%% > %.2f%%\n", results.ErrorRate, targetErrorRate)
	}
}
