package susgobench

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"os"
	"time"
)

const (
	certFile = "certs/rsa2048/server-chain.pem"
	keyFile  = "certs/rsa2048/server-key.pem"
	caFile   = "certs/rsa2048/ca-cert.pem"
)

func main() {
	// Start the server in a goroutine
	go func() {
		err := runServer()
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait a moment to ensure server starts before client
	time.Sleep(500 * time.Millisecond)

	// Start the client
	err := runClient()
	if err != nil {
		log.Fatalf("client error: %v", err)
	}
}

func tlsServerConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func runServer() error {

	config := tlsServerConfig()

	listener, err := tls.Listen("tcp", "localhost:8443", config)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("Server is listening on localhost:8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		tlsConn := conn.(*tls.Conn)
		tlsConn.Handshake()
		state := conn.(*tls.Conn).ConnectionState()
		log.Printf("Negotiated cipher suite: %s", tls.CipherSuiteName(state.CipherSuite))
		tlsConn.Close()
		//go handleConnection(conn)
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

func tlsClientConfig() *tls.Config {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		RootCAs:    caCertPool,
		ServerName: "localhost",
	}
}

func runClient() error {
	config := tlsClientConfig()

	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		return err
	}
	if !conn.ConnectionState().HandshakeComplete {
		log.Printf("something went very wrong")
	}
	defer conn.Close()

	// buf := make([]byte, 1024)
	// n, err := conn.Read(buf)
	// if err != nil {
	// 	return err
	// }

	// log.Printf("Received from server: %s", buf[:n])
	return nil
}
