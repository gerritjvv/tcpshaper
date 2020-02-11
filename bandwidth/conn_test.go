package bandwidth

import (
	"context"
	"io"
	"math/rand"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type connReadData struct {
	ts time.Duration
	n  int // bytes read
}

type mockConn struct {
	data   []byte
	cycles int

	counter int32
}

func (c *mockConn) GetAndAdd() int {
	v := c.counter
	atomic.AddInt32(&c.counter, 1)
	return int(v)
}

func newMockConn(byteLen int, cycles int) *mockConn {

	data := make([]byte, byteLen)
	for i := 0; i < byteLen; i++ {
		data[i] = byte(rand.Int())
	}

	return &mockConn{cycles: cycles, data: data, counter: 0}
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	i := c.GetAndAdd()
	if i < c.cycles {
		return copy(b, c.data), nil
	}

	return 0, io.EOF
}

func (c mockConn) Write(b []byte) (n int, err error) {
	panic("implement me")
}

func (c mockConn) Close() error {
	return nil
}

func (c mockConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (c mockConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c mockConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (c mockConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (c mockConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

// TestReadMoreThanBurst test that we get an error when we read more than the burst value
func TestReadMoreThanBurst(t *testing.T)  {
	conf := NewRateConfig(10, 20)

	limiter := NewBandwidthLimiter(conf)
	ctx := context.Background()

	conn := newMockConn(100, 3)
	rConn := NewRateLimitedConn(ctx, limiter, conn)

	_, err := rConn.Read(make([]byte, conf.Burst() + 1))

	if err == nil {
		t.Fatal()
	}
}

// TestNewRateLimitedConn checks that the overall limit and burst settings work as expected
// for a wrapped connection.
func TestNewRateLimitedConn(t *testing.T) {

	// The limit is 10 bytes per second overall
	// we can read 20 bytes in a single call
	// so if we do:
	//   read 20,
	//   then we have to wait 2 seconds

	conf := NewRateConfig(10, 20)

	limiter := NewBandwidthLimiter(conf)
	ctx := context.Background()

	conn := newMockConn(100, 3)
	rConn := NewRateLimitedConn(ctx, limiter, conn)

	var readData []connReadData

	startTime := time.Now()

	for {
		// read 20 bytes, now we need to wait 2 seconds
		n, err := rConn.Read(make([]byte, 20))
		if err != nil {
			break
		}
		timeAfterRead := time.Now()

		readData = append(readData, connReadData{
			ts: timeAfterRead.Sub(startTime).Round(time.Second),
			n:  n,
		})

		startTime = timeAfterRead
	}

	// check that all but the first time elapsed is 2 seconds
	for _, d := range readData[1:] {
		if d.ts != (2 * time.Second) {
			t.Fatalf("the time difference is not 2 seconds, we got %f", d.ts.Round(time.Second).Seconds())
		}

		if d.n != 20 {
			t.Fatalf("the number of bytes expected is 20 but we got %d", d.n)
		}
	}
}
