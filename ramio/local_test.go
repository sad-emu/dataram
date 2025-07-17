package ramio

import (
	"data_ram/ramstream"
	"os"
	"testing"
)

func TestLocalStream_ReadWrite(t *testing.T) {
	os.MkdirAll("test_data", 0755)
	fileName := "test_data/test_localstream.txt"
	defer os.Remove(fileName)

	// Test Write
	stream := NewLocalStream(ramstream.DROutputStream, fileName)
	data := []byte("hello world")
	n, err := stream.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Write length mismatch: got %d, want %d", n, len(data))
	}

	// Test Read
	streamRead := NewLocalStream(ramstream.DRInputStream, fileName)
	buf := make([]byte, len(data))
	n2, err2 := streamRead.Read(buf)
	if err2 != nil {
		t.Fatalf("Read failed: %v", err2)
	}
	if n2 != len(data) {
		t.Fatalf("Read length mismatch: got %d, want %d", n2, len(data))
	}
	if string(buf) != string(data) {
		t.Fatalf("Read data mismatch: got %s, want %s", buf, data)
	}
}
func TestLocalStream_ResetLen(t *testing.T) {
	os.MkdirAll("test_data", 0755)
	fileName := "test_data/test_localstream_len.txt"
	defer os.Remove(fileName)
	defer os.Remove(fileName)

	stream := NewLocalStream(ramstream.DROutputStream, fileName)
	data := []byte("abc")
	_, err := stream.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if stream.Len() != len(data) {
		t.Fatalf("Len mismatch: got %d, want %d", stream.Len(), len(data))
	}
	if err := stream.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
	// After reset, position should be 0
	if stream.Position != 0 {
		t.Fatalf("Position after reset should be 0, got %d", stream.Position)
	}
}
