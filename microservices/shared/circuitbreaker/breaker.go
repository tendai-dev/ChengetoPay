package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                sync.RWMutex
	state             State
	failureCount      int
	successCount      int
	lastFailureTime   time.Time
	failureThreshold  int
	successThreshold  int
	timeout           time.Duration
	onStateChange     func(from, to State)
}

// Config holds circuit breaker configuration
type Config struct {
	FailureThreshold int           // Number of failures to open circuit
	SuccessThreshold int           // Number of successes to close circuit from half-open
	Timeout          time.Duration // Time to wait before trying half-open
	OnStateChange    func(from, to State)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold <= 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}

	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: config.FailureThreshold,
		successThreshold: config.SuccessThreshold,
		timeout:          config.Timeout,
		onStateChange:    config.OnStateChange,
	}
}

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.recordResult(err == nil)
	return err
}

// allowRequest checks if the request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		return cb.successCount < cb.successThreshold
	default:
		return false
	}
}

// recordResult records the result of a request
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles successful requests
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.setState(StateClosed)
		}
	}
}

// onFailure handles failed requests
func (cb *CircuitBreaker) onFailure() {
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	oldState := cb.state
	cb.state = newState

	switch newState {
	case StateClosed:
		cb.failureCount = 0
		cb.successCount = 0
	case StateOpen:
		cb.successCount = 0
	case StateHalfOpen:
		cb.successCount = 0
	}

	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Metrics returns current metrics
func (cb *CircuitBreaker) Metrics() (State, int, int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failureCount, cb.successCount
}
