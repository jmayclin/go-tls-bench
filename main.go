package main

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

func tlsServerConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
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
