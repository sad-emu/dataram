package ramio

import (
	"data_ram/ramstream"
	"testing"
	"time"
)

// TCPStream backed by DummyStreams for testing
func TestTCPStreamWithDummyStreams(t *testing.T) {
	// Create DummyStreams for input and output
	input := &DummyStream{StreamType: ramstream.DROutputStream}
	output := &DummyStream{StreamType: ramstream.DROutputStream}

	// Wrap DummyStreams with TCPStream for testing
	address := "127.0.0.1:9100"
	tcpSender := NewTCPStream(address, ramstream.DROutputStream, nil)
	tcpListener := NewTCPStream(address, ramstream.DROutputStream, output)

	input.SubStream = tcpSender

	go tcpListener.Listen(1024)

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
