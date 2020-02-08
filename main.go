package main

import (
	"net"
	"os"
)

func main() {

	arg := os.Args[0]


	if arg == "server" {

		ln, err := net.Listen("tcp", ":7009")
		if err != nil {
			panic(err)
		}

		ln.Accept()


		return
	}

	if arg == "client" {

	}

}
