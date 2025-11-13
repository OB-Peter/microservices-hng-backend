package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	halfOpenMax  int

	mu            sync.RWMutex
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	successCount  int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		halfOpenMax:  3, // Allow 3 test requests in half-open state
		state:        StateClosed,
	}
}

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	err := fn()

	cb.afterRequest(err)
	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			fmt.Printf("🔄 Circuit breaker transitioning to HALF_OPEN\n")
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Limit concurrent requests in half-open state
		if cb.successCount >= cb.halfOpenMax {
			return ErrTooManyRequests
		}
		return nil

	default:
		return nil
	}
}

// afterRequest updates the circuit breaker state based on the result
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0

	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenMax {
			cb.state = StateClosed
			cb.failures = 0
			cb.successCount = 0
			fmt.Printf("✅ Circuit breaker CLOSED (service recovered)\n")
		}
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			fmt.Printf("⚠️  Circuit breaker OPENED (failures: %d/%d)\n", cb.failures, cb.maxFailures)
		}

	case StateHalfOpen:
		cb.state = StateOpen
		cb.successCount = 0
		fmt.Printf("⚠️  Circuit breaker OPENED again (half-open test failed)\n")
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":          cb.state.String(),
		"failures":       cb.failures,
		"max_failures":   cb.maxFailures,
		"success_count":  cb.successCount,
		"last_fail_time": cb.lastFailTime,
	}
}

// CircuitBreakerManager manages multiple circuit breakers for different services
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetBreaker returns or creates a circuit breaker for a service
func (cbm *CircuitBreakerManager) GetBreaker(serviceName string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	cbm.mu.RLock()
	breaker, exists := cbm.breakers[serviceName]
	cbm.mu.RUnlock()

	if exists {
		return breaker
	}

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := cbm.breakers[serviceName]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(maxFailures, resetTimeout)
	cbm.breakers[serviceName] = breaker
	fmt.Printf("🔧 Circuit breaker created for service: %s\n", serviceName)
	return breaker
}

// GetAllMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllMetrics() map[string]interface{} {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	metrics := make(map[string]interface{})
	for name, breaker := range cbm.breakers {
		metrics[name] = breaker.GetMetrics()
	}
	return metrics
}

// Example: API Gateway integration
var circuitBreakerManager = NewCircuitBreakerManager()

// CallServiceWithCircuitBreaker makes a call to a microservice with circuit breaker protection
func CallServiceWithCircuitBreaker(serviceName string, fn func() error) error {
	breaker := circuitBreakerManager.GetBreaker(
		serviceName,
		5,              // Open after 5 failures
		30*time.Second, // Try again after 30 seconds
	)

	err := breaker.Execute(fn)
	
	if err == ErrCircuitOpen {
		return fmt.Errorf("service %s is currently unavailable (circuit breaker open)", serviceName)
	}
	
	if err == ErrTooManyRequests {
		return fmt.Errorf("service %s is recovering, please try again later", serviceName)
	}

	return err
}

// Example: Use in API Gateway handlers
func ExampleGatewayHandler() {
	// Example: Calling RabbitMQ to publish notification
	err := CallServiceWithCircuitBreaker("rabbitmq", func() error {
		// Your RabbitMQ publish logic here
		// return publishToRabbitMQ(payload)
		return nil
	})

	if err != nil {
		// Handle circuit breaker error
		fmt.Printf("Failed to publish: %v\n", err)
		// Return fallback response or queue for later
	}
}

// Example: Health check endpoint with circuit breaker status
func HealthCheckWithCircuitBreakers() map[string]interface{} {
	return map[string]interface{}{
		"status":          "healthy",
		"circuit_breakers": circuitBreakerManager.GetAllMetrics(),
	}
}

// Example: Middleware to track service health
func TrackServiceHealth(serviceName string, err error) {
	breaker := circuitBreakerManager.GetBreaker(
		serviceName,
		5,
		30*time.Second,
	)

	if err != nil {
		breaker.afterRequest(err)
	} else {
		breaker.afterRequest(nil)
	}
}

// Example: Fallback handler when circuit is open
func HandleCircuitOpen(serviceName string) error {
	// Options when circuit is open:
	// 1. Return cached response
	// 2. Return degraded functionality
	// 3. Queue request for later processing
	// 4. Return user-friendly error

	fmt.Printf("⚠️  Circuit open for %s - using fallback strategy\n", serviceName)
	
	// Example: Queue for later processing
	// queueForLaterProcessing(request)
	
	return fmt.Errorf("service temporarily unavailable, request queued for processing")
}
