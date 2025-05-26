package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"testing"
	"time"
)

func makeTestTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic("failed to load test TLS cert: " + err.Error())
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		NextProtos:         []string{"dataram"},
		ServerName:         "localhost",
	}
}

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

	tlsConf := makeTestTLSConfig()
	// Start listener (receiver) in a goroutine
	go func() {
		err := AcceptAndHandleLoop(addr, dst, tlsConf, 1)
		if err != nil {
			fmt.Println("Listener error: \n", err)
		}
		done <- err // Always send, even if err is nil
	}()

	time.Sleep(100 * time.Millisecond) // Give listener time to start

	// Sender (client) connects and sends data
	runSender(addr, meta, src, 3, tlsConf)

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

	tlsConf := makeTestTLSConfig()
	go func() {
		err := AcceptAndHandleLoop(addr, dst, tlsConf, 1)
		if err != nil {
			done <- err
			return
		}
		done <- nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src, 3, tlsConf)
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

	tlsConf := makeTestTLSConfig()
	go func() {
		err := AcceptAndHandleLoop(addr, dst, tlsConf, 1)
		fmt.Println("Listener error:", err)
		done <- err // Always send, even if err is nil
	}()
	time.Sleep(100 * time.Millisecond)
	runSender(addr, meta, src, 3, tlsConf)
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

func TestFailBadTLSFileTransferQUIC(t *testing.T) {
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

	tlsConf := makeTestTLSConfig()
	go func() {
		err := AcceptAndHandleLoop(addr, dst, tlsConf, 1)
		fmt.Println("Listener error:", err)
		done <- err // Always send, even if err is nil
	}()
	time.Sleep(100 * time.Millisecond)
	tlsConf.InsecureSkipVerify = false // Disable insecure skip verify to test TLS failure
	runSender(addr, meta, src, 3, tlsConf)
	// Wait for either done or timeout, but also check for errors after the timeout
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("Listener should have error: %v", err)
		}
	case <-time.After(5 * time.Second):
		// Try to read from done once more in case the goroutine finished just after the timeout
		select {
		case err := <-done:
			if err == nil {
				t.Fatalf("Listener error should happen (after timeout): %v", err)
			}
		default:
			// Pass as we expect an error
			t.Log("Test timed out waiting for transfer to complete, as expected due to TLS failure")
		}
	}

}

// func TestPartialFileTransferTCP(t *testing.T) {
// 	data := []byte("TCP test data for file transfer! TCP test data for file transfer! " +
// 		"TCP test data for file transfer! TCP test data for file transfer!")
// 	chunkSize := 16
// 	src := &TestStream{}
// 	src.Write(data)
// 	meta := FileMetadata{
// 		Filename:  "testfile.txt",
// 		FlowName:  "testflow",
// 		Size:      int64(len(data)),
// 		ChunkSize: chunkSize,
// 		Metadata:  map[string]interface{}{"desc": "tcp test"},
// 		Transport: TCP_S,
// 	}
// 	dst := &TestStream{}
// 	addr := "localhost:9102"
// 	done := make(chan error, 1)

// 	tlsConf := makeTestTLSConfig()
// 	go func() {
// 		err := AcceptAndHandleOnce(addr, dst, tlsConf)
// 		if err != nil {
// 			done <- err
// 			return
// 		}
// 		done <- nil
// 	}()
// 	time.Sleep(100 * time.Millisecond)
// 	runSender(addr, meta, src, 3, tlsConf)
// 	select {
// 	case err := <-done:
// 		if err != nil {
// 			t.Fatalf("Listener error: %v", err)
// 		}
// 	case <-time.After(5 * time.Second):
// 		t.Fatal("Test timed out waiting for transfer to complete")
// 	}
// 	dst.SeekAbsolute(0)
// 	buf := make([]byte, len(data))
// 	n, _ := dst.Read(buf)
// 	if !bytes.Equal(buf[:n], data) {
// 		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
// 	}
// }

// func TestPartialFileTransferQUIC(t *testing.T) {
// 	data := []byte("QUIC test data for file transfer!")
// 	chunkSize := 16
// 	src := &TestStream{}
// 	src.Write(data)
// 	meta := FileMetadata{
// 		Filename:  "testfile_quic.txt",
// 		FlowName:  "testflow_quic",
// 		Size:      int64(len(data)),
// 		ChunkSize: chunkSize,
// 		Metadata:  map[string]interface{}{"desc": "quic test"},
// 		Transport: QUIC_S,
// 		QuicAddr:  "localhost:4243",
// 	}
// 	dst := &TestStream{}
// 	addr := "localhost:9103"
// 	done := make(chan error, 1)

// 	tlsConf := makeTestTLSConfig()
// 	go func() {
// 		err := AcceptAndHandleOnce(addr, dst, tlsConf)
// 		fmt.Println("Listener error:", err)
// 		done <- err // Always send, even if err is nil
// 	}()
// 	time.Sleep(100 * time.Millisecond)
// 	runSender(addr, meta, src, 3, tlsConf)
// 	// Wait for either done or timeout, but also check for errors after the timeout
// 	select {
// 	case err := <-done:
// 		if err != nil {
// 			t.Fatalf("Listener error: %v", err)
// 		}
// 	case <-time.After(5 * time.Second):
// 		// Try to read from done once more in case the goroutine finished just after the timeout
// 		select {
// 		case err := <-done:
// 			if err != nil {
// 				t.Fatalf("Listener error (after timeout): %v", err)
// 			}
// 			t.Fatal("Test timed out waiting for transfer to complete (but goroutine finished after timeout)")
// 		default:
// 			t.Fatal("Test timed out waiting for transfer to complete")
// 		}
// 	}
// 	dst.SeekAbsolute(0)
// 	buf := make([]byte, len(data))
// 	n, _ := dst.Read(buf)
// 	if !bytes.Equal(buf[:n], data) {
// 		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
// 	}
// }

func TestResumablePartialFileTransferTCP(t *testing.T) {
	data := []byte("TCP resumable test data for file transfer! TCP resumable test data for file transfer!")
	chunkSize := 16
	src := &TestStream{}
	src.Write(data)
	totalChunks := (len(data) + chunkSize - 1) / chunkSize
	// Only first half of chunks are available initially
	availability := make(map[int]bool)
	for i := 0; i < totalChunks/2; i++ {
		availability[i] = true
	}
	meta := FileMetadata{
		Filename:     "testfile_resumable.txt",
		FlowName:     "testflow_resumable",
		Size:         int64(len(data)),
		ChunkSize:    chunkSize,
		Metadata:     map[string]interface{}{"desc": "tcp resumable test"},
		Transport:    TCP_S,
		Availability: availability,
		Uuid:         "resumable-uuid-1",
	}
	dst := &TestStream{}
	addr := "localhost:9104"
	done := make(chan error, 1)
	tlsConf := makeTestTLSConfig()
	// Start listener (receiver) in a goroutine
	go func() {
		err := AcceptAndHandleLoop(addr, dst, tlsConf, 2)
		done <- err
	}()
	time.Sleep(100 * time.Millisecond)
	// First transfer: only partial chunks available
	runSender(addr, meta, src, 3, tlsConf)
	// Simulate sender now has all chunks available
	for i := totalChunks / 2; i < totalChunks; i++ {
		meta.Availability[i] = true
	}
	time.Sleep(200 * time.Millisecond) // Give listener time to process
	// Re-run sender with updated availability (same UUID)
	runSender(addr, meta, src, 3, tlsConf)
	// Wait for completion
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Listener error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for resumable transfer to complete")
	}
	dst.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := dst.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}
