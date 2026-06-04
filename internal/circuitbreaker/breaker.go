package circuitbreaker

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	}
	return "UNKNOWN"
}

type CircuitBreaker struct {
	name             string
	state            State
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	mu               sync.RWMutex
	failureThreshold int
	timeout          time.Duration
	maxHalfOpenReq   int
}

func NewCircuitBreaker(name string) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: 3,
		timeout:          60 * time.Second,
		maxHalfOpenReq:   3,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		switch cb.state {
		case StateClosed:
			cb.failureCount = 0
		case StateHalfOpen:
			cb.successCount++
			if cb.successCount >= cb.maxHalfOpenReq {
				cb.state = StateClosed
				cb.failureCount = 0
			}
		}
	} else {
		switch cb.state {
		case StateClosed:
			cb.failureCount++
			if cb.failureCount >= cb.failureThreshold {
				cb.state = StateOpen
				cb.lastFailureTime = time.Now()
			}
		case StateHalfOpen:
			cb.state = StateOpen
			cb.lastFailureTime = time.Now()
		}
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStateCode returns numeric code for Prometheus
func (cb *CircuitBreaker) GetStateCode() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	switch cb.state {
	case StateClosed:
		return 0
	case StateOpen:
		return 1
	case StateHalfOpen:
		return 2
	default:
		return 0
	}
}
