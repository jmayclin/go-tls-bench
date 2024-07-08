package susgobench

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func runServer(config *tls.Config) error {

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

// func TestLocalHostHandshake(t *testing.T) {
// 	serverConfig := tlsServerConfig()
// 	clientConfig := tlsClientConfig()

// 	listener, err := tls.Listen("tcp", "127.0.0.1:8444", serverConfig)
// 	if err != nil {
// 		t.Fatalf("failed to listen: %v", err)
// 	}
// 	defer listener.Close()

// 	go func() {
// 		for {
// 			conn, err := listener.Accept()
// 			if err != nil {
// 				log.Printf("failed to accept connection: %v", err)
// 				continue
// 			}
// 			go func() {
// 				tlsConn := conn.(*tls.Conn)
// 				if err := tlsConn.Handshake(); err != nil {
// 					log.Printf("server handshake failed: %v", err)
// 				}
// 				log.Println(tlsConn.ConnectionState())
// 				conn.Close()
// 			}()
// 		}
// 	}()

// 	time.Sleep(500 * time.Millisecond)

// 	conn, err := tls.Dial("tcp", "127.0.0.1:8444", clientConfig)
// 	if err != nil {
// 		t.Fatalf("client handshake failed: %v", err)
// 	}
// 	conn.Close()
// }

// func SharedMemTest(b *testing.B) {
// 	serverConfig := tlsServerConfig()
// 	clientConfig := tlsClientConfig()

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		clientToServer := &bytes.Buffer{}
// 		serverToClient := &bytes.Buffer{}

// 		clientConn := tls.Client(&dummyConn{r: serverToClient, w: clientToServer}, clientConfig)
// 		serverConn := tls.Server(&dummyConn{r: clientToServer, w: serverToClient}, serverConfig)

// 		b.Log("making the channel")
// 		done := make(chan bool)

// 		go func() {
// 			defer close(done)
// 			if err := serverConn.Handshake(); err != nil {
// 				b.Log("server handshake failed")
// 				//b.Fatalf("server handshake failed: %v", err)
// 			}
// 		}()

// 		if err := clientConn.Handshake(); err != nil {
// 			b.Fatalf("client handshake failed: %v", err)
// 		}

// 		<-done

// 		clientConn.Close()
// 		serverConn.Close()
// 	}
// }

func BenchmarkSharedMemHandshake(b *testing.B) {
	err := os.Setenv("GODEBUG", "tls=1")
	if err != nil {
		log.Fatalf("failed to set GODEBUG environment variable: %v", err)
	}

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

func TestSharedMemHandshake(t *testing.T) {
	err := os.Setenv("GODEBUG", "tls=1")
	if err != nil {
		log.Fatalf("failed to set GODEBUG environment variable: %v", err)
	}

	serverConfig := tlsServerConfig()
	clientConfig := tlsClientConfig()

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
		serverDone = true
		serverConn.Close()
	}()

	if err := clientConn.Handshake(); err != nil {
		t.Fatalf("client handshake failed: %v", err)
	}
	state := clientConn.ConnectionState()
	if !state.HandshakeComplete {
		t.Fatal("handshake was not complete")
	}

	clientDone = true
	clientConn.Close()
}

type dummyConn struct {
	name    string
	readCh  chan []byte
	writeCh chan []byte
	readBuf []byte
	closing *bool
}

func newDummyConn(name string, readCh, writeCh chan []byte, closing *bool) *dummyConn {
	return &dummyConn{
		name:    name,
		readCh:  readCh,
		writeCh: writeCh,
		closing: closing,
	}
}

func (c *dummyConn) Read(b []byte) (n int, err error) {
	//log.Printf("READ %s", c.name)
	if len(c.readBuf) == 0 {
		c.readBuf = <-c.readCh
	}
	n = copy(b, c.readBuf)
	c.readBuf = c.readBuf[n:]
	//log.Printf("READ %s done", c.name)
	return n, nil
}

func (c *dummyConn) Write(b []byte) (n int, err error) {
	//log.Printf("WRITE %s", c.name)
	if *c.closing {
		// pretend the write was successful
		//log.Printf("WRITE %s special done", c.name)
		return len(b), nil

	}
	c.writeCh <- b
	//log.Printf("WRITE %s done", c.name)

	return len(b), nil
}

func (c *dummyConn) Close() error {
	return nil
}

func (c *dummyConn) LocalAddr() net.Addr {
	return dummyAddr{}
}

func (c *dummyConn) RemoteAddr() net.Addr {
	return dummyAddr{}
}

func (c *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type dummyAddr struct{}

func (dummyAddr) Network() string { return "dummy" }
func (dummyAddr) String() string  { return "dummy" }
