package bandwidth

import (
	"context"
	"net"
)

type rateLimitedConnWrapper struct {
	net.Conn
	limiter Limiter
	ctx     context.Context
}

// Read reads data from the connection and is rate limited at bytes per second.
// len(b) must be bigger than the burst size set on the limiter, otherwise an error is returned.
func (c *rateLimitedConnWrapper) Read(b []byte) (int, error) {
	err := c.limiter.WaitN(c.ctx, len(b))
	if err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}

// NewRateLimitedConn returns a net.Conn that has its Read method rate limited
// by the limiter.
func NewRateLimitedConn(ctx context.Context, limiter Limiter, conn net.Conn) net.Conn {
	return &rateLimitedConnWrapper{
		Conn:    conn,
		ctx:     ctx,
		limiter: limiter,
	}
}
