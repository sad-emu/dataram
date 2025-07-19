package ramio

import (
	"data_ram/ramstream"
	"os"
	"testing"
	"time"
)

func TestLocalToTCPToLocalStream(t *testing.T) {

	os.MkdirAll("test_data", 0755)
	inputFile := "test_data/input.txt"
	outputFile := "test_data/output.txt"
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Write initial data to input file
	inputData := []byte("stream integration test")
	if err := os.WriteFile(inputFile, inputData, 0644); err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// LocalStream for reading
	localIn := NewLocalStream(ramstream.DRInputStream, inputFile)
	// LocalStream for writing
	localOut := NewLocalStream(ramstream.DROutputStream, outputFile)

	// TCPListener and TCPSender
	address := "127.0.0.1:9101"
	sender := NewTCPStream(address, ramstream.DROutputStream, nil)
	listener := NewTCPStream(address, ramstream.DROutputStream, localOut)

	go listener.Listen(1024)

	// wait for listener to start
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 1024)
	n, err := localIn.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	dataToSend := buf[:n]
	num, err := sender.Write(dataToSend)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if num != len(dataToSend) {
		t.Fatalf("Write length mismatch: got %d, want %d", num, len(dataToSend))
	}

	// Verify output file contains the same data
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(outputData) != string(dataToSend) {
		t.Fatalf("Output data mismatch: got %s, want %s", outputData, dataToSend)
	}

}
