package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"syscall"
)

var (
	addr   = flag.String("listen", ":8080", "MP-TCP Proxy Listening IPv4_Address:Port")
	server = flag.String("server", "10.200.0.2:5001", "TCP Server IPv4_Address:Port")
	client = flag.String("client", "10.60.0.1", "MP-TCP Client IPv4_Address:Port")
	v      = flag.Bool("vebose", true, "Enable Verbose Logging on Received Packets")
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func handleMPTcp(rhconn net.Conn, conn net.Conn) {
	for {
		// Read MP-TCP Packets
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err == io.EOF {
			return
		}
		if *v {
			log.Printf("[Proxy][MP-TCP] Received %d bytes: ", n)
		}

		// Forward Payload to TCP Server
		n, err = rhconn.Write(buf[:n])
		if err == io.EOF {
			return
		}
	}
}

func handleTcp(rhconn net.Conn, conn net.Conn) {
	for {
		// Read TCP Packets
		buf := make([]byte, 1024)
		n, err := rhconn.Read(buf)
		if err == io.EOF {
			return
		}

		if *v {
			log.Printf("[Proxy][TCP] Received %d bytes:", n)
		}

		// Forward Payload to MP-TCP Client
		n, err = conn.Write(buf[:n])
		if err == io.EOF {
			return
		}
	}
}

func main() {
	flag.Parse()

	// Enable MP-TCP on a TCP Socket
	lc := &net.ListenConfig{}
	lc.SetMultipathTCP(true)
	ln, err := lc.Listen(context.Background(), "tcp", *addr)
	checkError(err)

	// Listen for incoming connections
	for {
		conn, err := ln.Accept()
		checkError(err)
		defer conn.Close()

		// Check if the connection supports mptcp
		isMultipathTCP, err := conn.(*net.TCPConn).MultipathTCP()
		checkError(err)
		if isMultipathTCP {
			log.Printf("[Proxy][MP-TCP] Accepted connection from %s", conn.RemoteAddr())
		} else {
			log.Printf("[Proxy][TCP][Warning] Accepted connection from %s", conn.RemoteAddr())
		}

		// Redirect Connection to Server
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP:   net.ParseIP(*client),
				Port: 0,
			},
			Control: setSockOptIPTransparent,
		}

		rhconn, err := dialer.Dial("tcp", *server)
		checkError(err)

		// Handle Packets
		go handleTcp(rhconn, conn)
		go handleMPTcp(rhconn, conn)
	}
}

func setSockOptIPTransparent(network string, address string, c syscall.RawConn) error {
	// Set IP Transparent Socket to Listen on an arbitrary address
	var fn = func(s uintptr) {
		err := syscall.SetsockoptInt(int(s), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
		checkError(err)
	}
	err := c.Control(fn)
	checkError(err)

	return nil
}
