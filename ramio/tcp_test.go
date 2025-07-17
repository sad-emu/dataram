package ramio

import (
	"data_ram/ramstream"
	"testing"
)

// TCPStream backed by DummyStreams for testing
func TestTCPStreamWithDummyStreams(t *testing.T) {
	input := &DummyStream{StreamType: ramstream.DRInputStream, Data: []byte("abc")}
	output := &DummyStream{StreamType: ramstream.DROutputStream}

	// Simulate TCPListener reading from input DummyStream
	buf := make([]byte, 3)
	n, err := input.Read(buf)
	if err != nil || n != 3 || string(buf) != "abc" {
		t.Fatalf("DummyStream input failed: %v, %d, %s", err, n, buf)
	}

	// Simulate TCPSender writing to output DummyStream
	n2, err2 := output.Write(buf)
	if err2 != nil || n2 != 3 {
		t.Fatalf("DummyStream output failed: %v, %d", err2, n2)
	}

	// Read back from output DummyStream
	output.StreamType = ramstream.DRInputStream
	output.Reset()
	buf2 := make([]byte, 3)
	n3, err3 := output.Read(buf2)
	if err3 != nil || n3 != 3 || string(buf2) != "abc" {
		t.Fatalf("DummyStream output read failed: %v, %d, %s", err3, n3, buf2)
	}
}
