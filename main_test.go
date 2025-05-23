package main

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestMemoryStreamTransfer(t *testing.T) {
	data := []byte("Hello, this is a test file for transfer!")
	chunkSize := 8
	src := &TestStream{}
	src.Write(data)
	meta := FileMetadata{
		Filename:  "memory.txt",
		Size:      int64(len(data)),
		ChunkSize: chunkSize,
	}
	dst := &TestStream{}

	addr := "127.0.0.1:9101"
	done := make(chan error, 1)

	// Start listener (receiver) in a goroutine
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			done <- err
			return
		}
		defer ln.Close()
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		handleSender(conn, dst)
		conn.Close()
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond) // Give listener time to start

	// Sender (client) connects and sends data
	runSender(addr, meta, src)

	if err := <-done; err != nil {
		t.Fatalf("Listener error: %v", err)
	}

	dst.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := dst.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}
