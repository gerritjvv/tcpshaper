# tcpshaper
Golang Middleware for rate limiting tcp traffic

[![Build Status](https://travis-ci.org/gerritjvv/tcpshaper.svg?branch=master)](https://travis-ci.org/gerritjvv/tcpshaper)

# Overview

This package provides middleware and utility functions for rate limiting net.Listener connections.  

It makes use of the https://godoc.org/golang.org/x/time/rate package to provide the rate limiting itself.  



# Usage


## Listener 

```go
package main
import (

"context"
"fmt"
"github.com/gerritjvv/tcpshaper/bandwidth"
"net"
)

func main() {
	// Get a Listener, e.g tcp on any port
	tcpListner, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Configure rate limit to:
	// total read+write traffic == 1mbs
	// Each connection read+write == 2kbs
	// The maximum that can be read or written at any one time is 2kb
	serverRate := bandwidth.NewRateConfig(1024*1024, 2048)
	connRate := bandwidth.NewRateConfig(2048, 2048)

	listener := bandwidth.NewListener(ctx, &bandwidth.ListenerConfig{
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

```

## Connection wrapping

```go

package main
import (

"context"
"github.com/gerritjvv/tcpshaper/bandwidth"
)

func main()  {
 
	// Setup a limiter with a 2kb/s limit, and max 2kb per read event.
	readLimiter := bandwidth.NewBandwidthLimiter(bandwidth.NewRateConfig(2048, 2048))

	// Setup a "noop" limiter with infinite rates
	writeLimiter := bandwidth.NewBandwidthLimiter(bandwidth.NewRateConfig(bandwidth.Inf, 0))

	ctx := context.Background()

	conn := getConn() // get a connection

	rConn := bandwidth.NewRateLimitedConn(ctx, readLimiter, writeLimiter, conn)
	
	// Write is not rate limited
	_, _ = rConn.Write([]byte("test string"))

	// Read is reate limited
	_, _ = rConn.Read(make([]byte, 1))

	_ = rConn.Close()

}
```
# Build and Test

```bash

## to build the library
./build.sh build

## run tests
./build.sh test

## show test coverage reports
./build.sh report

```