package logtide

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerStateClosed(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Initially closed
	if cb.State() != CircuitClosed {
		t.Errorf("initial state = %v, want %v", cb.State(), CircuitClosed)
	}

	// Should allow requests
	if err := cb.Allow(); err != nil {
		t.Errorf("Allow() error = %v, want nil", err)
	}

	// Record success
	cb.RecordSuccess()
	if cb.Failures() != 0 {
		t.Errorf("failures = %d, want 0", cb.Failures())
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Record failures below threshold
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitClosed {
		t.Errorf("state after 2 failures = %v, want %v", cb.State(), CircuitClosed)
	}
	if cb.Failures() != 2 {
		t.Errorf("failures = %d, want 2", cb.Failures())
	}

	// Third failure should open the circuit
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("state after 3 failures = %v, want %v", cb.State(), CircuitOpen)
	}
	if cb.Failures() != 3 {
		t.Errorf("failures = %d, want 3", cb.Failures())
	}

	// Should not allow requests
	err := cb.Allow()
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Allow() error = %v, want %v", err, ErrCircuitOpen)
	}
}

func TestCircuitBreakerTransitionsToHalfOpen(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("state = %v, want %v", cb.State(), CircuitOpen)
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next allow should transition to half-open
	err := cb.Allow()
	if err != nil {
		t.Errorf("Allow() error = %v, want nil", err)
	}

	if cb.State() != CircuitHalfOpen {
		t.Errorf("state after timeout = %v, want %v", cb.State(), CircuitHalfOpen)
	}
}

func TestCircuitBreakerHalfOpenSuccess(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Allow()

	if cb.State() != CircuitHalfOpen {
		t.Errorf("state = %v, want %v", cb.State(), CircuitHalfOpen)
	}

	// Success in half-open should close the circuit
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Errorf("state after success = %v, want %v", cb.State(), CircuitClosed)
	}
	if cb.Failures() != 0 {
		t.Errorf("failures = %d, want 0", cb.Failures())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.Allow()

	if cb.State() != CircuitHalfOpen {
		t.Errorf("state = %v, want %v", cb.State(), CircuitHalfOpen)
	}

	// Failure in half-open should re-open the circuit
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("state after failure = %v, want %v", cb.State(), CircuitOpen)
	}

	// Should not allow requests
	err := cb.Allow()
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Allow() error = %v, want %v", err, ErrCircuitOpen)
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("state = %v, want %v", cb.State(), CircuitOpen)
	}

	// Reset
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Errorf("state after reset = %v, want %v", cb.State(), CircuitClosed)
	}
	if cb.Failures() != 0 {
		t.Errorf("failures after reset = %d, want 0", cb.Failures())
	}

	// Should allow requests
	if err := cb.Allow(); err != nil {
		t.Errorf("Allow() error = %v, want nil", err)
	}
}

func TestCircuitBreakerSuccessResetsFailures(t *testing.T) {
	config := &CircuitBreakerConfig{
		FailureThreshold: 3,
		Timeout:          100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Record some failures
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.Failures() != 2 {
		t.Errorf("failures = %d, want 2", cb.Failures())
	}

	// Success should reset failure count
	cb.RecordSuccess()

	if cb.Failures() != 0 {
		t.Errorf("failures after success = %d, want 0", cb.Failures())
	}
	if cb.State() != CircuitClosed {
		t.Errorf("state = %v, want %v", cb.State(), CircuitClosed)
	}
}

func TestCircuitStateString(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("CircuitState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
