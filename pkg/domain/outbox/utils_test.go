package outbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNextRetryDuration(t *testing.T) {
	var tables = []struct {
		name       string
		backoff    time.Duration
		retryCount int
		expect     time.Duration
	}{
		{
			name:       "Case1",
			backoff:    10 * time.Second,
			retryCount: 1,
			expect:     20 * time.Second,
		},
		{
			name:       "Case2",
			backoff:    10 * time.Second,
			retryCount: 2,
			expect:     40 * time.Second,
		},
		{
			name:       "Case3",
			backoff:    10 * time.Second,
			retryCount: 3,
			expect:     80 * time.Second,
		},
		{
			name:       "Case4",
			backoff:    3 * time.Second,
			retryCount: 1,
			expect:     6 * time.Second,
		},
		{
			name:       "Case5",
			backoff:    3 * time.Second,
			retryCount: 2,
			expect:     12 * time.Second,
		},
		{
			name:       "Case6",
			backoff:    3 * time.Second,
			retryCount: 3,
			expect:     24 * time.Second,
		},
	}
	for idx := range tables {
		tt := tables[idx]
		t.Run(tt.name, func(t *testing.T) {
			result := getNextRetryDuration(tt.backoff, tt.retryCount)
			assert.Equal(t, result, tt.expect)
		})
	}
}
