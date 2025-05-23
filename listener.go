package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Listener (Server)
func runListener(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Listener error:", err)
		return
	}
	fmt.Println("Listening on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		// Example usage with FileStream
		fileStream, err := NewFileStream("received_file.recv")
		if err != nil {
			fmt.Println("FileStream error:", err)
			conn.Close()
			continue
		}
		go handleSender(conn, fileStream)
	}
}

func handleSender(conn net.Conn, stream Stream) {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	var meta FileMetadata
	if err := dec.Decode(&meta); err != nil {
		fmt.Println("Metadata decode error:", err)
		return
	}
	fmt.Printf("Receiving file: %s (%d bytes)\n", meta.Filename, meta.Size)
	totalChunks := int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize))
	progress := &Progress{TotalChunks: totalChunks, Received: make(map[int]bool)}
	var mu sync.Mutex
	var wg sync.WaitGroup
	chunkResults := make(chan struct {
		idx  int
		data []byte
		ok   bool
	}, totalChunks)

	requestChunk := func(i int) {
		defer wg.Done()
		req := map[string]int{"request_chunk": i}
		mu.Lock()
		err := json.NewEncoder(conn).Encode(req)
		mu.Unlock()
		if err != nil {
			fmt.Printf("Request error for chunk %d: %v\n", i, err)
			chunkResults <- struct {
				idx  int
				data []byte
				ok   bool
			}{i, nil, false}
			return
		}
		var chunk struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
		}
		mu.Lock()
		err = dec.Decode(&chunk)
		mu.Unlock()
		if err != nil {
			fmt.Printf("Chunk decode error for chunk %d: %v\n", i, err)
			chunkResults <- struct {
				idx  int
				data []byte
				ok   bool
			}{i, nil, false}
			return
		}
		if err := stream.SeekAbsolute(int64(chunk.ChunkIndex * meta.ChunkSize)); err == nil {
			stream.Write(chunk.Data)
		}
		chunkResults <- struct {
			idx  int
			data []byte
			ok   bool
		}{chunk.ChunkIndex, chunk.Data, true}
	}

	for i := 0; i < totalChunks; i++ {
		wg.Add(1)
		go requestChunk(i)
	}

	receivedCount := 0
	for receivedCount < totalChunks {
		res := <-chunkResults
		if res.ok {
			mu.Lock()
			progress.Received[res.idx] = true
			mu.Unlock()
			fmt.Printf("Received chunk %d/%d\n", res.idx+1, totalChunks)
		}
		receivedCount++
	}
	wg.Wait()
	fmt.Println("File transfer complete.")
}
