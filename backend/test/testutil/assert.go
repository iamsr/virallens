package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// AssertValidUUID checks if a UUID is valid and not nil
func AssertValidUUID(t *testing.T, id uuid.UUID, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, uuid.Nil, id, msgAndArgs...)
}

// AssertTimeRecent checks if a time is within the last minute
func AssertTimeRecent(t *testing.T, timestamp time.Time, msgAndArgs ...interface{}) {
	t.Helper()
	now := time.Now()
	diff := now.Sub(timestamp)
	assert.True(t, diff < time.Minute && diff >= 0, msgAndArgs...)
}

// AssertNoError is a convenience wrapper for assert.NoError with better formatting
func AssertNoError(t *testing.T, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s failed: %v", operation, err)
	}
}

// AssertError checks that an error occurred and optionally matches a message
func AssertError(t *testing.T, err error, expectedMsg string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Error(t, err, msgAndArgs...)
	if expectedMsg != "" {
		assert.Contains(t, err.Error(), expectedMsg, msgAndArgs...)
	}
}
