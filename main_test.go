package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestBasicTCPTransfer(t *testing.T) {
	setupSettings()
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

	addr := "localhost:9101"
	done := make(chan error, 1)

	tlsConf := makeTestTLSConfig()
	activeTransfers := NewActiveTransfers()
	// Start listener (receiver) in a goroutine
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 1)
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

	tr, ok := activeTransfers.Get(meta.Uuid)
	if !ok {
		t.Fatalf("No transfer found for uuid %s", meta.Uuid)
	}
	tr.Stream.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := tr.Stream.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestFileTransferTCP(t *testing.T) {
	setupSettings()
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
	addr := "localhost:9102"
	done := make(chan error, 1)

	tlsConf := makeTestTLSConfig()
	activeTransfers := NewActiveTransfers()
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 1)
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
	tr, ok := activeTransfers.Get(meta.Uuid)
	if !ok {
		t.Fatalf("No transfer found for uuid %s", meta.Uuid)
	}
	tr.Stream.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := tr.Stream.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestFileTransferQUIC(t *testing.T) {
	setupSettings()
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
	addr := "localhost:9103"
	done := make(chan error, 1)

	tlsConf := makeTestTLSConfig()
	activeTransfers := NewActiveTransfers()
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 1)
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
	tr, ok := activeTransfers.Get(meta.Uuid)
	if !ok {
		t.Fatalf("No transfer found for uuid %s", meta.Uuid)
	}
	tr.Stream.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := tr.Stream.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestFailBadTLSFileTransferQUIC(t *testing.T) {
	setupSettings()
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
	addr := "localhost:9103"
	done := make(chan error, 1)

	tlsConf := makeTestTLSConfig()
	activeTransfers := NewActiveTransfers()
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 1)
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

func TestResumablePartialFileTransferTCP(t *testing.T) {
	setupSettings()
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

	addr := "localhost:9104"
	done := make(chan error, 1)
	tlsConf := makeTestTLSConfig()
	// Start listener (receiver) in a goroutine
	activeTransfers := NewActiveTransfers()
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 2)
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

	tr, ok := activeTransfers.Get(meta.Uuid)
	if !ok {
		t.Fatalf("No transfer found for uuid %s", meta.Uuid)
	}
	tr.Stream.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := tr.Stream.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestResumablePartialFileTransferQUIC(t *testing.T) {
	setupSettings()
	data := []byte("QUIC resumable test data for file transfer! QUIC resumable test data for file transfer!")
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
		Metadata:     map[string]interface{}{"desc": "QUIC resumable test"},
		Transport:    QUIC_S,
		Availability: availability,
		Uuid:         "resumable-uuid-1",
		QuicAddr:     "localhost:4244", // Use a unique QUIC port
	}
	addr := "localhost:9105" // Use a unique TCP port for the listener
	done := make(chan error, 1)
	tlsConf := makeTestTLSConfig()
	// Start listener (receiver) in a goroutine
	activeTransfers := NewActiveTransfers()
	go func() {
		err := AcceptAndHandleLoop(addr, activeTransfers, tlsConf, 2)
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
	tr, ok := activeTransfers.Get(meta.Uuid)
	if !ok {
		t.Fatalf("No transfer found for uuid %s", meta.Uuid)
	}
	tr.Stream.SeekAbsolute(0)
	buf := make([]byte, len(data))
	n, _ := tr.Stream.Read(buf)
	if !bytes.Equal(buf[:n], data) {
		t.Errorf("Received data does not match. Got: %q, Want: %q", buf[:n], data)
	}
}

func TestConcurrentLargeFileTransfersTCP(t *testing.T) {
	setupSettings()
	const numFiles = 10
	const fileSize = 1 << 20 // 1MB
	chunkSize := 32 * 1024   // 32KB per chunk
	done := make(chan error, numFiles)
	addrs := make([]string, numFiles)
	files := make([][]byte, numFiles)
	metas := make([]FileMetadata, numFiles)
	tlsConf := makeTestTLSConfig()
	activeTransfers := NewActiveTransfers()

	// Prepare data and unique ports
	for i := 0; i < numFiles; i++ {
		files[i] = make([]byte, fileSize)
		for j := range files[i] {
			files[i][j] = byte('a')
		}
		addrs[i] = fmt.Sprintf("localhost:%d", 9200+i)
		metas[i] = FileMetadata{
			Filename:  fmt.Sprintf("concurrent_file_%d.dat", i),
			FlowName:  fmt.Sprintf("concurrent_flow_%d", i),
			Size:      int64(fileSize),
			ChunkSize: chunkSize,
			Transport: TCP_S,
			Uuid:      fmt.Sprintf("concurrent-uuid-%d", i),
		}
	}

	// Start listeners
	for i := 0; i < numFiles; i++ {
		go func(idx int) {
			err := AcceptAndHandleLoop(addrs[idx], activeTransfers, tlsConf, 1)
			done <- err
		}(i)
	}
	time.Sleep(200 * time.Millisecond)

	// Start senders
	for i := 0; i < numFiles; i++ {
		src := &TestStream{}
		src.Write(files[i])
		go runSender(addrs[i], metas[i], src, 10, tlsConf)
	}

	// Wait for all listeners to finish
	for i := 0; i < numFiles; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Listener %d error: %v", i, err)
			}
		case <-time.After(20 * time.Second):
			t.Fatalf("Timeout waiting for listener %d", i)
		}
	}

	time.Sleep(200 * time.Millisecond)

	// Verify the data for each file
	for i := 0; i < numFiles; i++ {
		tr, ok := activeTransfers.Get(metas[i].Uuid)
		if !ok {
			t.Errorf("No transfer found for uuid %s", metas[i].Uuid)
			continue
		}
		tr.Stream.SeekAbsolute(0)
		buf := make([]byte, fileSize)
		n, _ := tr.Stream.Read(buf)
		fmt.Printf("File %d: expected %d bytes, got %d\n", i, len(files[i]), n)
		if n != len(files[i]) {
			t.Errorf("File %d: expected %d bytes, got %d", i, len(files[i]), n)
		}
		for j := 0; j < len(buf); j++ {
			if buf[j] != files[i][j] {
				t.Errorf("File %d: data mismatch at byte %d", i, j)
			}
		}
		if !bytes.Equal(buf, files[i]) {
			t.Errorf("File %d: data mismatch", i)
		} else {
			fmt.Printf("File %d: data matches successfully!\n", i)
		}
	}
}
