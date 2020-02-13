package main

import (
	"context"
	"flag"
	"github.com/gerritjvv/tcpshaper/bandwidth"
	"github.com/rjeczalik/netxtest"
	"log"
	"net"
)

func tcpshaperListener(l net.Listener, limitGlobal, limitPerConn int) net.Listener {

	ctx := context.Background()

	// For the test go run main.go -count 100 -limit-conn 0
	// The test itself calculates a range as a multiple of the connection count
	// e.g: want = 30s * t.Limit/t.Count => 30s * (26214400 / 100)
	//  I tried to dynamically adjust the connection rate based on the existing connection count
	//   but this did not pass the test either, because the conn count can be 1, 10, 100 etc at any moment.
	if limitPerConn == 0 {
		limitPerConn = limitGlobal / 100
	}

	serverRate := bandwidth.NewRateConfig(int64(limitGlobal), limitGlobal)
	connRate := bandwidth.NewRateConfig(int64(limitPerConn), limitPerConn)

	listener := bandwidth.NewListener(ctx, &bandwidth.ListenerConfig{
		ReadServerRate:  serverRate,
		WriteServerRate: serverRate,
		ReadConnRate:    connRate,
		WriteConnRate:   connRate,
	}, l)

	return listener
}

func main() {
	var test netxtest.LimitListenerTest

	test.RegisterFlags(flag.CommandLine)
	flag.Parse()

	if err := test.Run(tcpshaperListener); err != nil {
		log.Fatal(err)
	}
}
