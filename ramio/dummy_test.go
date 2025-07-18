package ramio

import (
	"data_ram/ramstream"
	"testing"
)

func TestDummyStream_WriteRead(t *testing.T) {
	stream := NewDummyStream(ramstream.DROutputStream)
	data := []byte("hello world")
	n, err := stream.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Write length mismatch: got %d, want %d", n, len(data))
	}

	stream.StreamType = ramstream.DRInputStream
	buf := make([]byte, len(data))
	stream.Reset()
	n2, err2 := stream.Read(buf)
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

func TestDummyStream_ResetLen(t *testing.T) {
	stream := NewDummyStream(ramstream.DROutputStream)
	data := []byte("abc")
	_, err := stream.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if stream.Len() != len(data) {
		t.Fatalf("Len mismatch: got %d, want %d", stream.Len(), len(data))
	}
	stream.Reset()
	if stream.Position != 0 {
		t.Fatalf("Position after reset should be 0, got %d", stream.Position)
	}
}

func TestDummyStream_StreamTypeEnforcement(t *testing.T) {
	stream := NewDummyStream(ramstream.DRInputStream)
	data := []byte("should not write")
	n, err := stream.Write(data)
	if n != 0 || err != nil {
		t.Fatalf("Write should not succeed for input stream")
	}

	stream = NewDummyStream(ramstream.DROutputStream)
	buf := make([]byte, 10)
	stream.Reset()
	n2, err2 := stream.Read(buf)
	if n2 != 0 || err2 != nil {
		t.Fatalf("Read should not succeed for output stream")
	}
}
