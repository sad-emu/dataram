package main

import (
	"io"
	"os"
	"testing"
)

func TestTestStream_WriteReadSeek(t *testing.T) {
	ts := &TestStream{}
	data := []byte("hello world")
	// Write at start
	n, err := ts.Write(data)
	if err != nil || n != len(data) {
		t.Fatalf("Write failed: n=%d, err=%v", n, err)
	}
	// Seek to start and read
	ts.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, err = ts.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != string(data) {
		t.Errorf("Read data mismatch: got %q, want %q", buf[:n], data)
	}
	// Seek and overwrite
	ts.SeekAbsolute(6)
	ts.Write([]byte("Go!"))
	ts.SeekAbsolute(0)
	buf2 := make([]byte, 16)
	n, _ = ts.Read(buf2)
	if string(buf2[:n]) != "hello Go!ld" {
		t.Errorf("Overwrite failed: got %q", buf2[:n])
	}
}

func TestFileStream_WriteReadSeek(t *testing.T) {
	fname := "test_filestream.tmp"
	defer os.Remove(fname)
	fs, err := NewFileStream(fname)
	if err != nil {
		t.Fatalf("NewFileStream failed: %v", err)
	}
	defer fs.Close()
	data := []byte("file stream test")
	// Write at start
	fs.SeekAbsolute(0)
	n, err := fs.Write(data)
	if err != nil || n != len(data) {
		t.Fatalf("Write failed: n=%d, err=%v", n, err)
	}
	// Seek to start and read
	fs.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, err = fs.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if string(buf[:n]) != string(data) {
		t.Errorf("Read data mismatch: got %q, want %q", buf[:n], data)
	}
	// Seek and overwrite
	fs.SeekAbsolute(5)
	fs.Write([]byte("STREAM"))
	fs.SeekAbsolute(0)
	buf2 := make([]byte, 32)
	n, _ = fs.Read(buf2)
	if string(buf2[:n]) != "file STREAM test" {
		t.Errorf("Overwrite failed: got %q", buf2[:n])
	}
}
