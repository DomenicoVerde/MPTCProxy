package main

import (
	"flag"
	"log"
	"net"
)

var (
	addr = flag.String("listen", "10.200.0.2:5001", "TCP Server Listening IPv4_Address:Port")
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	// Listen for TCP Connections
	ln, err := net.Listen("tcp", *addr)
	checkError(err)
	defer ln.Close()

	// Listen for incoming connections
	for {
		conn, err := ln.Accept()
		checkError(err)
		go func() {
			defer conn.Close()

			log.Printf("[Server][TCP] Accepted connection from %s", conn.RemoteAddr())

			// Read Packets
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				checkError(err)

				log.Printf("[Server][TCP] Received %d bytes from %s", n, conn.RemoteAddr())
			}
		}()
	}
}
