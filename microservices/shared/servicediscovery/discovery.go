package servicediscovery

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ServiceInstance represents a service instance
type ServiceInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Health  string
	Tags    []string
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	mu        sync.RWMutex
	services  map[string][]*ServiceInstance
	callbacks map[string][]func([]*ServiceInstance)
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:  make(map[string][]*ServiceInstance),
		callbacks: make(map[string][]func([]*ServiceInstance)),
	}
}

// RegisterService registers a service instance
func (sr *ServiceRegistry) RegisterService(instance *ServiceInstance) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.services[instance.Name] == nil {
		sr.services[instance.Name] = make([]*ServiceInstance, 0)
	}

	// Check if instance already exists
	for i, existing := range sr.services[instance.Name] {
		if existing.ID == instance.ID {
			sr.services[instance.Name][i] = instance
			sr.notifyCallbacks(instance.Name)
			return nil
		}
	}

	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	sr.notifyCallbacks(instance.Name)
	return nil
}

// DeregisterService removes a service instance
func (sr *ServiceRegistry) DeregisterService(serviceName, instanceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances := sr.services[serviceName]
	for i, instance := range instances {
		if instance.ID == instanceID {
			sr.services[serviceName] = append(instances[:i], instances[i+1:]...)
			sr.notifyCallbacks(serviceName)
			return nil
		}
	}

	return fmt.Errorf("service instance not found: %s/%s", serviceName, instanceID)
}

// GetService returns all healthy instances of a service
func (sr *ServiceRegistry) GetService(serviceName string) ([]*ServiceInstance, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances := sr.services[serviceName]
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

	// Filter healthy instances
	var healthy []*ServiceInstance
	for _, instance := range instances {
		if instance.Health == "passing" || instance.Health == "" {
			healthy = append(healthy, instance)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service: %s", serviceName)
	}

	return healthy, nil
}

// GetServiceInstance returns a single healthy instance using round-robin
func (sr *ServiceRegistry) GetServiceInstance(serviceName string) (*ServiceInstance, error) {
	instances, err := sr.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	// Simple random selection for load balancing
	return instances[rand.Intn(len(instances))], nil
}

// Subscribe to service changes
func (sr *ServiceRegistry) Subscribe(serviceName string, callback func([]*ServiceInstance)) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.callbacks[serviceName] == nil {
		sr.callbacks[serviceName] = make([]func([]*ServiceInstance), 0)
	}
	sr.callbacks[serviceName] = append(sr.callbacks[serviceName], callback)
}

// notifyCallbacks notifies subscribers of service changes
func (sr *ServiceRegistry) notifyCallbacks(serviceName string) {
	callbacks := sr.callbacks[serviceName]
	instances := sr.services[serviceName]

	for _, callback := range callbacks {
		go callback(instances)
	}
}

// LoadBalancer provides load balancing strategies
type LoadBalancer struct {
	registry *ServiceRegistry
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(registry *ServiceRegistry) *LoadBalancer {
	return &LoadBalancer{registry: registry}
}

// GetEndpoint returns a service endpoint with load balancing
func (lb *LoadBalancer) GetEndpoint(serviceName string) (string, error) {
	instance, err := lb.registry.GetServiceInstance(serviceName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d", instance.Address, instance.Port), nil
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.BaseDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			// Exponential backoff with jitter
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			// Add jitter (±25%)
			jitter := time.Duration(rand.Float64() * float64(delay) * 0.5)
			delay = delay + jitter - time.Duration(float64(delay)*0.25)
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}
	}

	return fmt.Errorf("max retries exceeded, last error: %w", lastErr)
}

// HealthChecker performs health checks on service instances
type HealthChecker struct {
	registry *ServiceRegistry
	interval time.Duration
	timeout  time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(registry *ServiceRegistry, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		interval: interval,
		timeout:  timeout,
	}
}

// Start begins health checking
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkAllServices()
		}
	}
}

// checkAllServices performs health checks on all registered services
func (hc *HealthChecker) checkAllServices() {
	hc.registry.mu.RLock()
	services := make(map[string][]*ServiceInstance)
	for name, instances := range hc.registry.services {
		services[name] = make([]*ServiceInstance, len(instances))
		copy(services[name], instances)
	}
	hc.registry.mu.RUnlock()

	for serviceName, instances := range services {
		for _, instance := range instances {
			go hc.checkInstance(serviceName, instance)
		}
	}
}

// checkInstance performs a health check on a single instance
func (hc *HealthChecker) checkInstance(serviceName string, instance *ServiceInstance) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	// Simple TCP connection check
	// In a real implementation, this would make HTTP health check requests
	_ = fmt.Sprintf("%s:%d", instance.Address, instance.Port)
	
	// Simulate health check
	select {
	case <-ctx.Done():
		instance.Health = "critical"
	case <-time.After(10 * time.Millisecond):
		instance.Health = "passing"
	}

	// Update instance in registry
	hc.registry.RegisterService(instance)
}
