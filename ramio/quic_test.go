package ramio

import (
	"data_ram/ramstream"
	"os"
	"testing"
	"time"
)

// QUICStream backed by DummyStreams for testing
func TestQUICStreamWithDummyStreams(t *testing.T) {
	// Create DummyStreams for input and output
	input := &DummyStream{StreamType: ramstream.DROutputStream}
	output := &DummyStream{StreamType: ramstream.DROutputStream}

	// Use the test keys in data_ram/tools/cert.pem and data_ram/tools/key.pem
	// to generate a TLS config for QUIC
	certPEM, err := os.ReadFile("../tools/cert.pem")
	if err != nil {
		t.Fatalf("failed to read cert.pem: %v", err)
	}
	keyPEM, err := os.ReadFile("../tools/key.pem")
	if err != nil {
		t.Fatalf("failed to read key.pem: %v", err)
	}
	tlsConfig := generateTLSConfig(certPEM, keyPEM)

	// Wrap DummyStreams with TCPStream for testing
	address := "127.0.0.1:45454"
	quicSender := NewQUICStream(address, ramstream.DROutputStream, nil, tlsConfig)
	quicListener := NewQUICStream(address, ramstream.DROutputStream, output, tlsConfig)

	input.SubStream = quicSender

	go quicListener.Listen()

	// wait for listener to start
	time.Sleep(100 * time.Millisecond)

	dataToSend := []byte("abcdefg")
	num, err := input.Write(dataToSend)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if num != len(dataToSend) {
		t.Fatalf("Write length mismatch: got %d, want %d", num, len(dataToSend))
	}
	// Data is empty for some reason???
	// Verify output.data equals dataToSend and error if not equal
	for i := 0; i < len(dataToSend); i++ {
		if output.Data[i] != dataToSend[i] {
			t.Errorf("Expected %c, got %c at index %d", dataToSend[i], output.Data[i], i)
		}
	}
}
