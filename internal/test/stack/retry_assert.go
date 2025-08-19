package stack

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// AssertEqualWithRetry performs an assertion with retry logic
// It will retry the assertion up to maxRetries times with delay between attempts
func AssertEqualWithRetry(t *testing.T, getValue func() interface{}, expected interface{}, maxRetries int, delay time.Duration) {
	var lastActual interface{}

	for i := 0; i < maxRetries; i++ {
		lastActual = getValue()

		// Try the assertion - if it passes, we're done
		if lastActual == expected {
			assert.Equal(t, lastActual, expected)
			return
		}

		// If this is not the last retry, wait before trying again
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}

	// All retries failed, perform final assertion to show the failure
	assert.Equal(t, lastActual, expected, fmt.Sprintf("Assertion failed after %d retries", maxRetries))
}

// AssertEqualWithRetryDefault uses default retry settings (5 retries, 1 second delay)

func AssertEqualWithRetryDefault(t *testing.T, getValue func() interface{}, expected interface{}) {

	AssertEqualWithRetry(t, getValue, expected, 5, 1*time.Second)

}

// Retry runs predicate until it returns true or the retries are exhausted.
func Retry(t *testing.T, predicate func() bool, maxRetries int, delay time.Duration) {
	t.Helper()
	for i := 0; i < maxRetries; i++ {
		if predicate() {
			return
		}
		if i < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	t.Fatalf("condition not met after %d retries (delay %s)", maxRetries, delay)
}
