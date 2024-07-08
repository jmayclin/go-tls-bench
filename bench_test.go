package susgobench

import (
	"testing"
)

func BenchmarkServerAuth(b *testing.B) {
	serverConfig := tlsServerConfig()
	clientConfig := tlsClientConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		harness_handshake(clientConfig, serverConfig)
	}

}

func BenchmarkResumption(b *testing.B) {
	serverConfig := resumptionServerConfig()
	clientConfig := resumptionClientConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		harness_handshake(clientConfig, serverConfig)
	}

}

func TestServerAuth(t *testing.T) {
	serverConfig := tlsServerConfig()
	clientConfig := tlsClientConfig()

	harness_handshake(clientConfig, serverConfig)
}

// TODO: bad test. make this return the connection state, and then assert on resumption
func TestResumption(t *testing.T) {
	serverConfig := resumptionServerConfig()
	clientConfig := resumptionClientConfig()

	harness_handshake(clientConfig, serverConfig)
}
