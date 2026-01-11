package logtide

import (
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// CircuitClosed means requests are allowed through.
	CircuitClosed CircuitState = iota

	// CircuitOpen means requests are blocked.
	CircuitOpen

	// CircuitHalfOpen means the circuit is testing if the service has recovered.
	CircuitHalfOpen
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
type CircuitBreaker struct {
	mu sync.RWMutex

	// Configuration
	failureThreshold int           // Number of consecutive failures before opening
	timeout          time.Duration // Time to wait before transitioning to half-open

	// State
	state            CircuitState
	failures         int       // Consecutive failure count
	lastFailureTime  time.Time // Time of last failure
	lastStateChange  time.Time // Time of last state change
}

// CircuitBreakerConfig holds the configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	Timeout          time.Duration
}

// DefaultCircuitBreakerConfig returns the default circuit breaker configuration.
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the specified configuration.
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		failureThreshold: config.FailureThreshold,
		timeout:          config.Timeout,
		state:            CircuitClosed,
		lastStateChange:  time.Now(),
	}
}

// Allow checks if a request is allowed based on the circuit breaker state.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from open to half-open
	if cb.state == CircuitOpen {
		if time.Since(cb.lastStateChange) >= cb.timeout {
			cb.state = CircuitHalfOpen
			cb.lastStateChange = time.Now()
		} else {
			return ErrCircuitOpen
		}
	}

	return nil
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset failure count
	cb.failures = 0

	// If we were in half-open state, transition to closed
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.lastStateChange = time.Now()
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	// If we're in half-open state, a single failure trips the circuit
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		cb.lastStateChange = time.Now()
		return
	}

	// Check if we've exceeded the failure threshold
	if cb.failures >= cb.failureThreshold {
		cb.state = CircuitOpen
		cb.lastStateChange = time.Now()
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current consecutive failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Reset resets the circuit breaker to the closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.failures = 0
	cb.lastStateChange = time.Now()
}
