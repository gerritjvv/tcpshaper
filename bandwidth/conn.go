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

// Read reads data from the connection and is rate limited at a byte level.
// len(b) must be bigger than the burst size set on the limiter, otherwise an error is returned.
func (c *rateLimitedConnWrapper) Read(b []byte) (int, error) {

	// Note: len(b) cannot be bigger than the Limiters burst size
	// see https://pkg.go.dev/golang.org/x/time/rate?tab=doc#Limiter.WaitN
	err := c.limiter.WaitN(c.ctx, len(b))
	if err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}

func RateLimitedConn(ctx context.Context, limiter Limiter, conn net.Conn) net.Conn {
	return &rateLimitedConnWrapper{
		Conn:    conn,
		ctx:     ctx,
		limiter: limiter,
	}
}
