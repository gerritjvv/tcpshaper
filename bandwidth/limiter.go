package bandwidth

import (
	"context"
	"golang.org/x/time/rate"
)

type limiter struct {
	*rate.Limiter
	conf   *RateConfig
	parent Limiter
}

// Limiter abstracts the idea of a rate limiter in this package.
// A Limiter can also create a hierarchy of parent child limiters.
type Limiter interface {
	// Wait blocks till n bytes per second are available.
	// This can be for the server or per connection
	WaitN(tx context.Context, n int) error
	Configure(conf *RateConfig)

	// Child create's a child limiter, that will call check the parent's limit before
	// checking its own limit
	Child(conf *RateConfig) Limiter
}

// NewBandwidthLimiter creates a limiter to use with tcp connection and tcp listener bytes per second rate limiting.
func NewBandwidthLimiter(conf *RateConfig) Limiter {
	return newBandwidthLimiter(nil, conf)
}

func newBandwidthLimiter(parent Limiter, conf *RateConfig) Limiter {
	return &limiter{
		conf:    conf,
		Limiter: rate.NewLimiter(rate.Limit(conf.Limit()), conf.Burst()),
		parent:  parent,
	}
}

func (l *limiter) Child(conf *RateConfig) Limiter {
	return newBandwidthLimiter(l, conf)
}

func (l *limiter) WaitN(ctx context.Context, n int) error {

	// call parent limiter is present
	if l.parent != nil {
		err := l.parent.WaitN(ctx, n)

		if err != nil {
			return err
		}
	}

	// this is the simplest way to ensure we always have the updated config
	// alternatives such as chaining Configure functions or having config listeners
	// do not see worth the complication here, especially when having to deal with cleaning
	// out listeners to avoid memory leaks.
	l.Configure(l.conf)

	return l.Limiter.WaitN(ctx, n)
}

// Configure updates the Limiter's limit and burst values from RateConfig.
func (l *limiter) Configure(conf *RateConfig) {
	l.Limiter.SetLimit(rate.Limit(conf.Limit()))
	l.SetBurst(conf.Burst())
}
