package bandwidth

import (
	"context"
	"net"
	"testing"
	"time"
)

type connWriteData struct {
	ts time.Duration
	n  int // bytes read
}

type mockConn struct {
}

func (c *mockConn) Read(b []byte) (int, error) {
	return len(b), nil
}

func (c mockConn) Write(b []byte) (int, error) {
	return len(b), nil
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

func (c mockConn) SetDeadline(_ time.Time) error {
	panic("implement me")
}

func (c mockConn) SetReadDeadline(_ time.Time) error {
	panic("implement me")
}

func (c mockConn) SetWriteDeadline(_ time.Time) error {
	panic("implement me")
}

// TestReadMoreThanBurst test that we get an error when we write more than the burst value
func TestReadMoreThanBurst(t *testing.T) {
	conf := NewRateConfig(10, 20)

	limiter := NewBandwidthLimiter(conf)
	ctx := context.Background()

	conn := &mockConn{}
	rConn := NewRateLimitedConn(ctx, limiter, limiter, conn)

	_, err := rConn.Write(make([]byte, conf.Burst()+1))

	if err == nil {
		t.Fatal()
	}

	_, err = rConn.Read(make([]byte, conf.Burst()+1))
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

	conn := &mockConn{}
	rConn := NewRateLimitedConn(ctx, limiter, limiter, conn)

	var writeData []connWriteData

	startTime := time.Now()

	var n int
	var err error

	bts := make([]byte, 20)

	for i := 0; i < 3; i++ {
		// read 20 bytes, now we need to wait 2 seconds
		if i%2 == 0 {
			n, err = rConn.Read(bts)
		} else {
			n, err = rConn.Write(bts)
		}

		if err != nil {
			t.Fatalf("no error expected here %s", err)
			return
		}
		timeAfterRead := time.Now()

		writeData = append(writeData, connWriteData{
			ts: timeAfterRead.Sub(startTime).Round(time.Second),
			n:  n,
		})

		startTime = timeAfterRead
	}

	// check that all but the first time elapsed is 2 seconds
	for _, d := range writeData[1:] {
		if d.ts != (2 * time.Second) {
			t.Fatalf("the time difference is not 2 seconds, we got %f", d.ts.Round(time.Second).Seconds())
		}

		if d.n != 20 {
			t.Fatalf("the number of bytes expected is 20 but we got %d", d.n)
		}
	}
}
