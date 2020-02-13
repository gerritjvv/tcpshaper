package bandwidth

import (
	"math"
	"sync"
)

// Inf infinite rate
var Inf int64 = math.MaxInt64

// RateConfig holds the limiter configuration limit and burst values.
type RateConfig struct {
	rwLock sync.RWMutex
	// limit is the overall bytes per second rate
	limit int64
	// buts is the number of bytes that can be consumed in a single Read call
	burst int
}

// SetLimit sets the overall bytes per second rate
func (conf *RateConfig) SetLimit(limit int64) {
	conf.rwLock.Lock()
	defer conf.rwLock.Unlock()

	conf.limit = limit
}

// SetBurst sets the number of bytes that can be consumed in a single Read call
func (conf *RateConfig) SetBurst(burst int) {
	conf.rwLock.Lock()
	defer conf.rwLock.Unlock()

	conf.burst = validateBurst(burst, conf.limit)
}

// Limit returns the limit in bytes per second.
func (conf *RateConfig) Limit() int64 {
	conf.rwLock.RLock()
	defer conf.rwLock.RUnlock()

	return conf.limit
}

// Burst returns the burst in bytes per second.
func (conf *RateConfig) Burst() int {
	conf.rwLock.RLock()
	defer conf.rwLock.RUnlock()

	return conf.burst
}

func validateBurst(burst int, limit int64) int {
	if burst <= 0 {
		burst = int(limit)
	}

	return burst
}

func validateLimit(limit int64) int64 {
	if limit < 1 {
		return Inf
	}

	return limit
}

// NewRateConfig contains the over limit in bytes per second and the burst; maximum bytes that can be read in a single call.
// The RateConfig instance that can be read and updated from multiple go routines.
func NewRateConfig(limit int64, burst int) *RateConfig {

	vLimit := validateLimit(limit)

	config := RateConfig{
		limit: vLimit,
		burst: validateBurst(burst, vLimit),
	}

	return &config
}
