package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
)

const (
	certFile = "certs/rsa2048/server-chain.pem"
	keyFile  = "certs/rsa2048/server-key.pem"
	caFile   = "certs/rsa2048/ca-cert.pem"
)

func main() {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("Server is listening on localhost:8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	rand.Read(buf) // Generate random data for demonstration

	n, err := conn.Write([]byte("Hello, World!\n"))
	if err != nil {
		log.Printf("failed to write to client: %v", err)
		return
	}

	log.Printf("Wrote %d bytes to client.", n)
}
