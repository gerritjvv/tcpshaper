package bandwidth

import (
	"context"
	"net"
)

type rateLimitedConnWrapper struct {
	net.Conn
	writeLimiter Limiter
	readLimiter Limiter

	ctx     context.Context
}

// Read data from a connection. The reads are rate limited at bytes per second.
// len(b) must be bigger than the burst size set on the limiter, otherwise an error is returned.
func (c *rateLimitedConnWrapper) Read(b []byte) (int, error) {
	err := c.readLimiter.WaitN(c.ctx, len(b))
	if err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}


// Write data to a connection. The writes are rate limited at bytes per second.
// len(b) must be bigger than the burst size set on the limiter, otherwise an error is returned.
func (c *rateLimitedConnWrapper) Write(b []byte) (int, error) {
	err := c.writeLimiter.WaitN(c.ctx, len(b))
	if err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}

// NewRateLimitedConn returns a net.Conn that has its Read method rate limited
// by the limiter.
func NewRateLimitedConn(ctx context.Context, readLimiter Limiter, writeLimiter Limiter, conn net.Conn) net.Conn {
	return &rateLimitedConnWrapper{
		Conn:    conn,
		ctx:     ctx,
		readLimiter: readLimiter,
		writeLimiter: writeLimiter,
	}
}
