package susgobench

import (
	"crypto/tls"
	"log"
	"os"
	"testing"
)

func BenchmarkSharedMemHandshake(b *testing.B) {
	serverConfig := tlsServerConfig()
	clientConfig := tlsClientConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		clientToServer := make(chan []byte)
		serverToClient := make(chan []byte)

		clientDone := false
		serverDone := false

		clientConn := tls.Client(newDummyConn("client", serverToClient, clientToServer, &clientDone), clientConfig)
		serverConn := tls.Server(newDummyConn("server", clientToServer, serverToClient, &serverDone), serverConfig)

		//done := make(chan bool)

		go func() {
			if err := serverConn.Handshake(); err != nil {
				b.Logf("server handshake failed: %v", err)
			}
			serverDone = true
			serverConn.Close()
		}()

		if err := clientConn.Handshake(); err != nil {
			b.Fatalf("client handshake failed: %v", err)
		}
		state := clientConn.ConnectionState()
		if !state.HandshakeComplete {
			b.Fatal("handshake was not complete")
		}
		//log.Printf("Negotiated cipher suite: %s", tls.CipherSuiteName(state.CipherSuite))
		clientDone = true
		clientConn.Close()
	}

}

func BenchmarkSharedMemResumption(b *testing.B) {
	serverConfig := resumptionServerConfig()
	clientConfig := resumptionClientConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		clientToServer := make(chan []byte)
		serverToClient := make(chan []byte)

		clientDone := false
		serverDone := false

		clientConn := tls.Client(newDummyConn("client", serverToClient, clientToServer, &clientDone), clientConfig)
		serverConn := tls.Server(newDummyConn("server", clientToServer, serverToClient, &serverDone), serverConfig)

		//done := make(chan bool)

		go func() {
			if err := serverConn.Handshake(); err != nil {
				b.Logf("server handshake failed: %v", err)
			}
			serverDone = true
			serverConn.Close()
		}()

		if err := clientConn.Handshake(); err != nil {
			b.Fatalf("client handshake failed: %v", err)
		}
		state := clientConn.ConnectionState()
		if !state.HandshakeComplete {
			b.Fatal("handshake was not complete")
		}
		readBuf := make([]byte, 1)
		clientDone = true
		clientConn.Read(readBuf)
		//log.Println("client is closing")
		clientConn.Close()
	}

}

func TestSharedMemHandshake(t *testing.T) {
	err := os.Setenv("GODEBUG", "tls=1")
	if err != nil {
		log.Fatalf("failed to set GODEBUG environment variable: %v", err)
	}

	serverConfig := resumptionServerConfig()
	clientConfig := resumptionClientConfig()

	clientToServer := make(chan []byte)
	serverToClient := make(chan []byte)

	clientDone := false
	serverDone := false

	clientConn := tls.Client(newDummyConn("client", serverToClient, clientToServer, &clientDone), clientConfig)
	serverConn := tls.Server(newDummyConn("server", clientToServer, serverToClient, &serverDone), serverConfig)

	//done := make(chan bool)

	t.Log("making the channel")

	go func() {
		if err := serverConn.Handshake(); err != nil {
			t.Log("server handshake failed")
			t.Log((err))
			//b.Fatalf("server handshake failed: %v", err)
		}
		readBuf := make([]byte, 1)
		n, _ := serverConn.Read(readBuf)
		if n != 0 {
			t.Log("ERROR: unexpected read")
		}
		// writeBuf := make([]byte, 1)
		// clientConn.Write(writeBuf)

		serverDone = true
		log.Println("server is closing")
		serverConn.Close()
	}()

	if err := clientConn.Handshake(); err != nil {
		t.Fatalf("client handshake failed: %v", err)
	}
	state := clientConn.ConnectionState()
	if !state.HandshakeComplete {
		t.Fatal("handshake was not complete")
	}
	readBuf := make([]byte, 1)
	log.Println("getting the session ticket")
	clientDone = true
	clientConn.Read(readBuf)
	//log.Println("client is closing")
	clientConn.Close()
}
