package susgobench

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
)

const (
	certFile = "certs/rsa2048/server-chain.pem"
	keyFile  = "certs/rsa2048/server-key.pem"
	caFile   = "certs/rsa2048/ca-cert.pem"
)

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

func main() {
	config := tlsClientConfig()

	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("failed to read from server: %v", err)
	}

	log.Printf("Received from server: %s", buf[:n])
}
