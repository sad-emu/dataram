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
	tcpListener.Flush()
}

func TestTCPReuseStreamWithDummyStreams(t *testing.T) {
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

	dataToSendAgain := []byte("34567890-adfadsfa!@#$!@#%^!#$^$#")
	num, err = input.Write(dataToSendAgain)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if num != len(dataToSendAgain) {
		t.Fatalf("Write length mismatch: got %d, want %d", num, len(dataToSendAgain))
	}

	// append dataToSendAgain to dataToSend
	dataToSend = append(dataToSend, dataToSendAgain...)

	for i := 0; i < len(dataToSend); i++ {
		if output.Data[i] != dataToSend[i] {
			t.Errorf("Expected %c, got %c at index %d", dataToSend[i], output.Data[i], i)
		}
	}
	tcpListener.Flush()
}

// At the moment, a single TCPStream listener can only handle one input stream at a time.
// For more throughput use multiple TCPStream listeners.

// func TestTCPMultiStreamWithDummyStreams(t *testing.T) {
// 	// Create DummyStreams for input and output
// 	input := &DummyStream{StreamType: ramstream.DROutputStream}
// 	output := &DummyStream{StreamType: ramstream.DROutputStream}

// 	// Wrap DummyStreams with TCPStream for testing
// 	address := "127.0.0.1:9100"
// 	tcpSender := NewTCPStream(address, ramstream.DROutputStream, nil)
// 	tcpListener := NewTCPStream(address, ramstream.DROutputStream, output)

// 	input.SubStream = tcpSender

// 	go tcpListener.Listen(1024)

// 	// wait for listener to start
// 	time.Sleep(100 * time.Millisecond)

// 	// Send through n number of data blocks at once to check if the stream can handle multiple writes
// 	// in a for loop
// 	for i := 0; i < 100; i++ {
// 		// do this in a goroutine to simulate multiple writes
// 		go func(idx int) {
// 			// Sleep a random time
// 			time.Sleep(time.Duration(idx%10) * time.Millisecond)
// 			dataToSend := []byte(fmt.Sprintf("abcdefg%d", idx))
// 			num, err := input.Write(dataToSend)
// 			if err != nil {
// 				t.Fatalf("Write failed: %v", err)
// 			}
// 			if num != len(dataToSend) {
// 				t.Fatalf("Write length mismatch: got %d, want %d", num, len(dataToSend))
// 			}
// 			// Verify output.data equals dataToSend and error if not equal
// 			for j := 0; j < len(dataToSend); j++ {
// 				if output.Data[idx*len(dataToSend)+j] != dataToSend[j] {
// 					t.Errorf("Expected %c, got %c at index %d", dataToSend[j], output.Data[idx*len(dataToSend)+j], idx*len(dataToSend)+j)
// 				}
// 			}
// 		}(i)
// 	}

// 	tcpListener.Flush()
// }
