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
	certFile = "certs/rsae_pkcs_2048_sha256/server-chain.pem"
	keyFile  = "certs/rsae_pkcs_2048_sha256/server-key.pem"
	caFile   = "certs/rsae_pkcs_2048_sha256/ca-cert.pem"
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

func resumptionServerConfig() *tls.Config {

	// Generate session ticket keys
	sessionTicketKeys := make([][32]byte, 1)
	if _, err := rand.Read(sessionTicketKeys[0][:]); err != nil {
		log.Fatalf("failed to generate session ticket key: %v", err)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %v", err)
	}

	// TODO: is explicitly disabling session ticket necessary?
	config := tls.Config{
		Certificates:           []tls.Certificate{cert},
		SessionTicketKey:       sessionTicketKeys[0],
		SessionTicketsDisabled: false,
	}
	//config.SetSessionTicketKeys(sessionTicketKeys)
	return &config
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

func resumptionClientConfig() *tls.Config {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		RootCAs:                caCertPool,
		ClientSessionCache:     NewLoggingSessionCache(128),
		ServerName:             "localhost",
		SessionTicketsDisabled: false,
	}
}

func harness_handshake(clientConfig, serverConfig *tls.Config) {
	clientToServer := make(chan []byte)
	serverToClient := make(chan []byte)

	clientDone := false
	serverDone := false

	clientConn := tls.Client(newDummyConn("client", serverToClient, clientToServer, &clientDone), clientConfig)
	serverConn := tls.Server(newDummyConn("server", clientToServer, serverToClient, &serverDone), serverConfig)

	//done := make(chan bool)

	go func() {
		if err := serverConn.Handshake(); err != nil {
			log.Println("server handshake failed")
			log.Println((err))
		}
		readBuf := make([]byte, 1)
		n, _ := serverConn.Read(readBuf)
		if n != 0 {
			log.Println("ERROR: unexpected read")
		}

		serverDone = true
		log.Println("server is closing")
		serverConn.Close()
	}()

	if err := clientConn.Handshake(); err != nil {
		log.Fatalf("client handshake failed: %v", err)
	}
	state := clientConn.ConnectionState()
	if !state.HandshakeComplete {
		log.Fatal("handshake was not complete")
	}
	readBuf := make([]byte, 1)
	clientDone = true
	// this read is necessary to get the session ticket with is a post handshake
	// message
	clientConn.Read(readBuf)
	clientConn.Close()
}

// TODO: remove the name, only used for debug messages/logging
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
	//log.Printf("READ %s with %d len", c.name, len(b))
	if *c.closing {
		// return fake 1 byte, to avoid fighting with their weird
		// io
		//log.Printf("READ %s special done", c.name)
		return 0, net.ErrClosed
	}
	if len(c.readBuf) == 0 {
		c.readBuf = <-c.readCh
	}
	n = copy(b, c.readBuf)
	c.readBuf = c.readBuf[n:]
	//log.Printf("READ %s done with %d bytes", c.name, n)
	return n, nil
}

func (c *dummyConn) Write(b []byte) (n int, err error) {
	//log.Printf("WRITE %s %d bytes", c.name, len(b))
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

// TODO: This was only used for debugging, when I forget that I have to call read to retrieve the NST bc TLS 1.3 shenanigans
// remove this and just use the underlying NewLRUClientSessionCache
type loggingSessionCache struct {
	cache tls.ClientSessionCache
}

func NewLoggingSessionCache(size int) tls.ClientSessionCache {
	return &loggingSessionCache{
		cache: tls.NewLRUClientSessionCache(size),
	}
}

func (l *loggingSessionCache) Put(key string, cs *tls.ClientSessionState) {
	//fmt.Printf("Adding session with key: %s\n", key)
	l.cache.Put(key, cs)
}

func (l *loggingSessionCache) Get(key string) (*tls.ClientSessionState, bool) {
	session, ok := l.cache.Get(key)
	if ok {
		//fmt.Printf("Retrieved session with key: %s\n", key)
	} else {
		//fmt.Printf("No session found with key: %s\n", key)
	}
	return session, ok
}

func main() {
	log.Println("hello world")
}
