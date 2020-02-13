package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gerritjvv/tcpshaper/bandwidth"
	"github.com/rjeczalik/netxtest"
	"log"
	"net"
)

func tcpshaperListener(l net.Listener, limitGlobal, limitPerConn int) net.Listener {

	fmt.Printf("tcpshaperListener: limitGlobal: %d, limitPerConn: %d\n", limitGlobal, limitPerConn)

	ctx := context.Background()

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
