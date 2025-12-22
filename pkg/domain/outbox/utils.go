package outbox

import (
	"math"
	"time"
)

// getNextRetryDuration return the next retry duration in seconds based on the attempt,
// the formula: `backoff * 2 ^ attempt`
func getNextRetryDuration(backoff time.Duration, attempt int) time.Duration {
	return time.Duration(backoff.Seconds()*math.Pow(2, float64(attempt))) * time.Second
}
