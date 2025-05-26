package main

import (
	"bytes"
	"fmt"
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
		Transport: TCP_S,
	}
	dst := &TestStream{}

	addr := "localhost:9101"
	done := make(chan error, 1)

	// Start listener (receiver) in a goroutine
	go func() {
		err := AcceptAndHandleOnce(addr, dst)
		if err != nil {
			fmt.Println("ffffffffffffffffffffffffffffffffffListener errorfdfdfdf: \n", err)
		}
		done <- err // Always send, even if err is nil
	}()

	time.Sleep(100 * time.Millisecond) // Give listener time to start

	// Sender (client) connects and sends data
	runSender(addr, meta, src, 3)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for transfer to complete")
	}

	dst.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := dst.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestFileTransferTCP(t *testing.T) {
	data := []byte("TCP test data for file transfer! TCP test data for file transfer! " +
		"TCP test data for file transfer! TCP test data for file transfer!")
	chunkSize := 16
	src := &TestStream{}
	src.Write(data)
	meta := FileMetadata{
		Filename:  "testfile.txt",
		FlowName:  "testflow",
		Size:      int64(len(data)),
		ChunkSize: chunkSize,
		Metadata:  map[string]interface{}{"desc": "tcp test"},
		Transport: TCP_S,
	}
	dst := &TestStream{}
	addr := "localhost:9102"
	done := make(chan error, 1)
	go func() {
		err := AcceptAndHandleOnce(addr, dst)
		if err != nil {
			done <- err
			return
		}
		done <- nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src, 3)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for transfer to complete")
	}
	dst.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := dst.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestFileTransferQUIC(t *testing.T) {
	data := []byte("QUIC test data for file transfer!")
	chunkSize := 16
	src := &TestStream{}
	src.Write(data)
	meta := FileMetadata{
		Filename:  "testfile_quic.txt",
		FlowName:  "testflow_quic",
		Size:      int64(len(data)),
		ChunkSize: chunkSize,
		Metadata:  map[string]interface{}{"desc": "quic test"},
		Transport: QUIC_S,
		QuicAddr:  "localhost:4243",
	}
	dst := &TestStream{}
	addr := "localhost:9103"
	done := make(chan error, 1)
	go func() {
		err := AcceptAndHandleOnce(addr, dst)
		fmt.Println("Listener error:", err)
		done <- err // Always send, even if err is nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src, 3)
	// Wait for either done or timeout, but also check for errors after the timeout
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	case <-time.After(5 * time.Second):
		// Try to read from done once more in case the goroutine finished just after the timeout
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Listener error (after timeout): %v", err)
			}
			t.Fatal("Test timed out waiting for transfer to complete (but goroutine finished after timeout)")
		default:
			t.Fatal("Test timed out waiting for transfer to complete")
		}
	}
	dst.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := dst.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}
