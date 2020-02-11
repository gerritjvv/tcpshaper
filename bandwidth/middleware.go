package bandwidth

import (
	"context"
	"net"
)

type rateListWrapper struct {
	net.Listener

	serverLimiter Limiter
	connRate      *rateConfig

	ctx context.Context
}

// Accept returns a new connection or error.
// The new connection is rate limited, configured by connRateConf and the parent serverLimiter
func (w *rateListWrapper) Accept() (net.Conn, error) {

	conn, err := w.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// The child will check its connRateConf limits and also the overall serverLimiter
	return NewRateLimitedConn(w.ctx, w.serverLimiter.Child(w.connRate), conn), err
}

// NewListener returns a net.Listener that will apply rate limits each connection and also globally for all connections
// via the serverRate config.
func NewListener(ctx context.Context, serverRate *rateConfig, connRate *rateConfig, listener net.Listener) net.Listener {
	serverLimiter := NewBandwidthLimiter(serverRate)

	return &rateListWrapper{
		Listener:      listener,
		serverLimiter: serverLimiter,
		connRate:      connRate,
		ctx:           ctx,
	}
}
