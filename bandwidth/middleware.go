package bandwidth

import (
	"context"
	"net"
)

// ListenerConfig groups together the configuration for a Listener and the limiters that should be used.
type ListenerConfig struct {

	// ReadServerRate the global server read limit and burst config
	ReadServerRate *RateConfig

	// WriteServerRate the global server write limit and burst config
	WriteServerRate *RateConfig

	// ReadConnRate the per connection read limit and burst config
	ReadConnRate *RateConfig

	// WriteConnRate the per connection write limit and burst config
	WriteConnRate *RateConfig
}

// NewListenerConfig is a helper function to create a ListenerConfig from a single RateConfig.
// the ReadServerRate, WriterServerRate, ReadConnRate, and WriteConnRate are all set to the RateConfig.
func NewListenerConfig(rateConfig *RateConfig) *ListenerConfig {
	return &ListenerConfig{
		ReadServerRate:  rateConfig,
		WriteServerRate: rateConfig,
		ReadConnRate:    rateConfig,
		WriteConnRate:   rateConfig,
	}
}

type rateListWrapper struct {
	net.Listener

	serverReadLimiter  Limiter
	serverWriteLimiter Limiter

	listenerConfig *ListenerConfig

	ctx context.Context
}

// Accept returns a new connection or error.
// The new connection is rate limited, configured by the connection rate limits and the parent serverLimiter
func (w *rateListWrapper) Accept() (net.Conn, error) {

	conn, err := w.Listener.Accept()
	if err != nil {
		return nil, err
	}

	readLimiter := w.serverReadLimiter.Child(w.listenerConfig.ReadConnRate)
	writeLimiter := w.serverWriteLimiter.Child(w.listenerConfig.WriteConnRate)

	// The child will check its connection rate limits and also the overall serverLimiter
	return NewRateLimitedConn(w.ctx, readLimiter, writeLimiter, conn), err
}

// NewListener returns a net.Listener that will apply rate limits to each connection and also globally for all connections
// via the listenerConfig.ReadServerRate and listenerConfig.WriteServerRate configs.
func NewListener(ctx context.Context, listenerConfig *ListenerConfig, listener net.Listener) net.Listener {

	return &rateListWrapper{
		Listener: listener,

		serverReadLimiter:  NewBandwidthLimiter(listenerConfig.ReadServerRate),
		serverWriteLimiter: NewBandwidthLimiter(listenerConfig.WriteServerRate),

		listenerConfig: listenerConfig,

		ctx: ctx,
	}
}
