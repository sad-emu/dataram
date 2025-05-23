package main

import (
	"bytes"
	"fmt"
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
		Transport: "TCP",
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
		// Wrap handleSender in a recover to catch panics and errors
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		handleSender(conn, dst)
		conn.Close()
		done <- nil
	}()

	time.Sleep(100 * time.Millisecond) // Give listener time to start

	// Sender (client) connects and sends data
	runSender(addr, meta, src)

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
	data := []byte("TCP test data for file transfer!")
	chunkSize := 16
	src := &TestStream{}
	src.Write(data)
	meta := FileMetadata{
		Filename:  "testfile.txt",
		FlowName:  "testflow",
		Size:      int64(len(data)),
		ChunkSize: chunkSize,
		Metadata:  map[string]interface{}{"desc": "tcp test"},
		Transport: "TCP",
	}
	dst := &TestStream{}
	addr := "127.0.0.1:9102"
	done := make(chan error, 1)
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
		// Wrap handleSender in a recover to catch panics and errors
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		handleSender(conn, dst)
		conn.Close()
		done <- nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src)
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
		Transport: "QUIC",
		QuicAddr:  "localhost:4243",
	}
	dst := &TestStream{}
	addr := "127.0.0.1:9103"
	done := make(chan error, 1)
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
		// Wrap handleSender in a recover to catch panics and errors
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		handleSender(conn, dst)
		conn.Close()
		done <- nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src)
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
