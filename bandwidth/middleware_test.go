package bandwidth

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
)

func writeByteString(w io.Writer, v string) (int, error) {
	// ignore correct value for n here, this is a test helper function
	n, err := w.Write([]byte{byte(len(v))})
	if err != nil {
		return n, err
	}

	return w.Write([]byte(v))
}

func readByteString(r io.Reader) (string, error) {
	// ignore correct use of n and checking for sending correct bytes or overflow
	lenBytes := make([]byte, 1)

	// wait for first byte
	for {
		n, err := r.Read(lenBytes)
		if err != nil {
			return "", err
		}
		if n > 0 {
			break
		}
	}

	strBytes := make([]byte, int(lenBytes[0]))
	_, err := r.Read(strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes), nil
}

// TestNewListener runs a full server and client cycle of reading and writing.
// The aim of the test is not to test rate limiting but rather a sanity check for the Listener wrapper.
// For rate limit tests please see the conn_test.go and limiter_test.go files.
func TestNewListener(t *testing.T) {

	tcpListner, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	rateConf := NewRateConfig(10, 100)

	listener := NewListener(ctx, NewListenerConfig(rateConf), tcpListner)

	testString := "test string"

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("error while accepting connection %s", err)
		}

		// ignore errors here
		_, _ = writeByteString(conn, testString)

		err = conn.Close()
		if err != nil {
			t.Fatalf("server conn.Close, no error expected here %s", err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial, no error expected here %s", err)
	}

	v, err := readByteString(conn)
	if err != nil {
		t.Fatalf("client readByteString, no error expected here %s", err)
	}

	err = conn.Close()
	if err != nil {
		t.Fatalf("no error expected here %s", err)
	}

	err = listener.Close()
	if err != nil {
		t.Fatalf("no error expected here %s", err)
	}

	// wait for server routine to complete
	wg.Wait()

	if v != testString {
		t.Fatalf("%s expected from server but got %s", testString, v)
	}

	// check that we error listener accept after close
	_, err = listener.Accept()
	if err == nil {
		t.Fatalf("error expected here but got %s", err)
	}

}

// ExampleNewListener shows how reate limit an existing net.Listener
func ExampleNewListener() {

	// Get a Listener, e.g tcp on any port
	tcpListner, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// configure rate limit to:
	// total read+write traffic == 1mbs
	// Each connection read+write == 2kbs
	// The maximum that can be read or written at any one time is 2kb
	serverRate := NewRateConfig(1024*1024, 2048)
	connRate := NewRateConfig(2048, 2048)

	listener := NewListener(ctx, &ListenerConfig{
		ReadServerRate:  serverRate,
		WriteServerRate: serverRate,
		ReadConnRate:    connRate,
		WriteConnRate:   connRate,
	}, tcpListner)

	// Now use the listener
	// e.g listener.Accept()

	err = listener.Close()
	if err != nil {
		fmt.Printf("error while closing listener %s", err)
	}
}

// ExampleNewRateLimitedConn wrap an existing net.Con and rate limit its read write operations
func ExampleNewRateLimitedConn() {

	// Setup a limiter with a 2kb/s limit, and max 2kb per read event.
	readLimiter := NewBandwidthLimiter(NewRateConfig(2048, 2048))

	// Setup a limiter infinite write limit
	writeLimiter := NewBandwidthLimiter(NewRateConfig(Inf, 0))

	ctx := context.Background()

	conn := &mockConn{} // get a connection

	rConn := NewRateLimitedConn(ctx, readLimiter, writeLimiter, conn)

	// Write is not rate limited
	_, _ = rConn.Write([]byte("test string"))

	// Read is reate limited
	_, _ = rConn.Read(make([]byte, 1))

	_ = rConn.Close()
}