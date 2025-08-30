package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ServiceInstance represents a registered service instance
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Health   string            `json:"health"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

// ServiceRegistry manages service discovery and health checking
type ServiceRegistry struct {
	mu        sync.RWMutex
	services  map[string][]*ServiceInstance
	callbacks map[string][]func([]*ServiceInstance)
	client    *http.Client
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:  make(map[string][]*ServiceInstance),
		callbacks: make(map[string][]func([]*ServiceInstance)),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
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
	log.Printf("Registered service instance: %s/%s at %s:%d", instance.Name, instance.ID, instance.Address, instance.Port)
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
			log.Printf("Deregistered service instance: %s/%s", serviceName, instanceID)
			return nil
		}
	}

	return fmt.Errorf("service instance not found: %s/%s", serviceName, instanceID)
}

// GetHealthyInstances returns all healthy instances of a service
func (sr *ServiceRegistry) GetHealthyInstances(serviceName string) ([]*ServiceInstance, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances := sr.services[serviceName]
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found for service: %s", serviceName)
	}

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

// GetServiceEndpoint returns a service endpoint using round-robin load balancing
func (sr *ServiceRegistry) GetServiceEndpoint(serviceName string) (string, error) {
	instances, err := sr.GetHealthyInstances(serviceName)
	if err != nil {
		return "", err
	}

	// Simple round-robin selection
	instance := instances[time.Now().UnixNano()%int64(len(instances))]
	return fmt.Sprintf("http://%s:%d", instance.Address, instance.Port), nil
}

// StartHealthChecking begins periodic health checks
func (sr *ServiceRegistry) StartHealthChecking(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting health checks with interval: %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping health checks")
			return
		case <-ticker.C:
			sr.performHealthChecks()
		}
	}
}

// performHealthChecks checks health of all registered services
func (sr *ServiceRegistry) performHealthChecks() {
	sr.mu.RLock()
	services := make(map[string][]*ServiceInstance)
	for name, instances := range sr.services {
		services[name] = make([]*ServiceInstance, len(instances))
		copy(services[name], instances)
	}
	sr.mu.RUnlock()

	var wg sync.WaitGroup
	for serviceName, instances := range services {
		for _, instance := range instances {
			wg.Add(1)
			go func(sName string, inst *ServiceInstance) {
				defer wg.Done()
				sr.checkInstanceHealth(sName, inst)
			}(serviceName, instance)
		}
	}
	wg.Wait()
}

// checkInstanceHealth performs a health check on a single instance
func (sr *ServiceRegistry) checkInstanceHealth(serviceName string, instance *ServiceInstance) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	healthURL := fmt.Sprintf("http://%s:%d/health", instance.Address, instance.Port)
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		instance.Health = "critical"
		return
	}

	resp, err := sr.client.Do(req)
	if err != nil {
		instance.Health = "critical"
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		instance.Health = "passing"
	} else {
		instance.Health = "warning"
	}

	// Update instance in registry
	sr.RegisterService(instance)
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

// GetAllServices returns all registered services
func (sr *ServiceRegistry) GetAllServices() map[string][]*ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make(map[string][]*ServiceInstance)
	for name, instances := range sr.services {
		result[name] = make([]*ServiceInstance, len(instances))
		copy(result[name], instances)
	}
	return result
}

// InitializeDefaultServices registers default service instances
func (sr *ServiceRegistry) InitializeDefaultServices() {
	defaultServices := map[string]int{
		"escrow-service":        8081,
		"payment-service":       8083,
		"ledger-service":        8084,
		"risk-service":          8085,
		"treasury-service":      8086,
		"evidence-service":      8087,
		"compliance-service":    8088,
		"workflow-service":      8089,
		"journal-service":       8091,
		"fees-service":          8092,
		"refunds-service":       8093,
		"transfers-service":     8094,
		"fx-service":            8095,
		"payouts-service":       8096,
		"reserves-service":      8097,
		"reconciliation-service": 8098,
		"kyb-service":           8099,
		"sca-service":           8100,
		"disputes-service":      8101,
		"dx-service":            8102,
		"auth-service":          8103,
		"idempotency-service":   8104,
		"eventbus-service":      8105,
		"saga-service":          8106,
		"vault-service":         8119,
		"webhooks-service":      8108,
		"observability-service": 8109,
		"config-service":        8110,
		"workers-service":       8111,
		"portal-service":        8112,
		"data-platform-service": 8113,
		"compliance-ops-service": 8114,
		"database-service":      8115,
		"monitoring-service":    8116,
	}

	for serviceName, port := range defaultServices {
		instance := &ServiceInstance{
			ID:      fmt.Sprintf("%s-1", serviceName),
			Name:    serviceName,
			Address: "localhost",
			Port:    port,
			Health:  "unknown",
			Tags:    []string{"default"},
			Metadata: map[string]string{
				"version": "1.0.0",
				"env":     "development",
			},
		}
		sr.RegisterService(instance)
	}

	log.Printf("Initialized %d default service instances", len(defaultServices))
}
