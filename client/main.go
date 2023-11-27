package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var (
	addr = flag.String("proxy_addr", "192.168.0.2:8080", "MP-TCP Explicit Proxy IPv4_Address:Port")
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	// Enable MP-TCP on the socket (this should be off by default)
	d := &net.Dialer{}
	d.SetMultipathTCP(true)

	// Connect to the Proxy
	c, err := d.Dial("tcp", *addr)
	checkError(err)
	defer c.Close()

	// Check if MP-TCP is supported by the proxy, else panic
	tcp, _ := c.(*net.TCPConn)
	mptcp, err := tcp.MultipathTCP()
	checkError(err)
	if !mptcp {
		log.Printf("[Client][MP-TCP][Warning]: MP-TCP Not Supported by host %s. Using TCP instead.", *addr)
	} else {
		log.Printf("[Client][MP-TCP] Connection established with %s", *addr)
	}

	// Send test packets periodically to Proxy
	count := 1
	for {
		payload := fmt.Sprintf("MP-TCP Test Packet # %d", count)
		n, err := c.Write([]byte(payload))
		checkError(err)

		log.Printf("[Client][MP-TCP] Packet #%d, Sent %d bytes to %s\n", count, n, *addr)
		count++
		time.Sleep(1 * time.Second)
	}
}
