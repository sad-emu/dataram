package ramio

import (
	"crypto/tls"
	"net"
)

// Helper to read a big-endian int64 from a connection
func readInt64(conn net.Conn) (int64, error) {
	buf := make([]byte, 8)
	_, err := conn.Read(buf)
	if err != nil {
		return 0, err
	}
	val := int64(buf[0])<<56 | int64(buf[1])<<48 | int64(buf[2])<<40 | int64(buf[3])<<32 |
		int64(buf[4])<<24 | int64(buf[5])<<16 | int64(buf[6])<<8 | int64(buf[7])
	return val, nil
}

// Helper to write a big-endian int64 to a connection
func writeInt64(conn net.Conn, val int64) error {
	buf := []byte{
		byte(val >> 56), byte(val >> 48), byte(val >> 40), byte(val >> 32),
		byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val),
	}
	_, err := conn.Write(buf)
	return err
}

// generateTLSConfig returns a self-signed TLS config for testing
func generateTLSConfig(certIn []byte, keyIn []byte) *tls.Config {
	// For demo purposes only: use a proper cert in production
	cert, err := tls.X509KeyPair(certIn, keyIn)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}
