/*
Package bandwidth provides rate limiting utilities and middleware for TCP connections.

The NewListener function returns a net.Listener that wraps an existing listener and rate limit globally
over all connections and each connection itself.

  // 1mbs
  serverRate := NewRateConfig(1024 * 1024, 1024 * 1024)
  // 200 kbs
  connRate := NewRateConfig(1024 * 200, 1024 * 200)

  ln, err := net.Listen("tcp", ":8080")
  if err != nil {
	// handle error
  }

  listener := bandwidth.NewListener(context, serverRate, connRate, ln)

  for {
   conn, err := ln.Accept()
   if err != nil {
	// handle error
   }
	go handleConnection(conn)
  }

*/
package bandwidth
