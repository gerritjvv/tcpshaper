package bandwidth

import "sync/atomic"

type RateConfig struct {
	limit atomic.Value
	burst atomic.Value
}

func (conf *RateConfig) SetLimit(limit float64) {
	conf.limit.Store(limit)
}

func (conf *RateConfig) SetBurst(burst int) {
	conf.limit.Store(burst)
}

func (conf *RateConfig) Limit() float64 {
	// We don't expect a type cast here because we know the value is int
	l, _ := conf.limit.Load().(float64)
	return l
}

func (conf *RateConfig) Burst() int {
	// We don't expect a type cast here because we know the value is int
	b, _ := conf.limit.Load().(int)
	return b
}