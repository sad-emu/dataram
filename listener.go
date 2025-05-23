package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	quic "github.com/quic-go/quic-go"
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

// AcceptAndHandleOnce accepts a single connection and handles it, for use in tests.
func AcceptAndHandleOnce(addr string, stream Stream) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	conn, err := ln.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	handleSender(conn, stream)
	return nil
}

func handleQuic(meta FileMetadata, mu *sync.Mutex, wg *sync.WaitGroup, chunkResultsChan chan ChunkResult, stream Stream, progress *Progress) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"dataram"},
		ServerName:         "localhost",
	}
	session, err := quic.DialAddr(context.Background(), meta.QuicAddr, tlsConf, nil)
	if err != nil {
		fmt.Println("QUIC dial error:", err)
		return
	}
	defer session.CloseWithError(0, "")
	totalChunks := progress.TotalChunks
	requestChunk := func(i int) {
		defer wg.Done()
		streamSend, err := session.OpenStreamSync(context.Background())
		if err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		json.NewEncoder(streamSend).Encode(map[string]int{"request_chunk": i})
		streamSend.Close()
		streamRecv, err := session.AcceptStream(context.Background())
		if err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		var chunk struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
		}
		if err := json.NewDecoder(streamRecv).Decode(&chunk); err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if err := stream.SeekAbsolute(int64(chunk.ChunkIndex * meta.ChunkSize)); err == nil {
			stream.Write(chunk.Data)
		}
		chunkResultsChan <- ChunkResult{Index: chunk.ChunkIndex, Data: chunk.Data, Ok: true}
	}
	for i := 0; i < totalChunks; i++ {
		wg.Add(1)
		go requestChunk(i)
	}
	receivedCount := 0
	for receivedCount < totalChunks {
		res := <-chunkResultsChan
		if res.Ok {
			mu.Lock()
			progress.Received[res.Index] = true
			mu.Unlock()
		}
		receivedCount++
	}
	wg.Wait()
	fmt.Println("File transfer complete.")
}

func handleTcp(meta FileMetadata, mu *sync.Mutex, wg *sync.WaitGroup, chunkResultsChan chan ChunkResult, stream Stream, progress *Progress, dec *json.Decoder, conn net.Conn) {
	totalChunks := progress.TotalChunks
	requestChunk := func(i int) {
		defer wg.Done()
		req := map[string]int{"request_chunk": i}
		mu.Lock()
		err := json.NewEncoder(conn).Encode(req)
		mu.Unlock()
		if err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
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
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if err := stream.SeekAbsolute(int64(chunk.ChunkIndex * meta.ChunkSize)); err == nil {
			stream.Write(chunk.Data)
		}
		chunkResultsChan <- ChunkResult{Index: chunk.ChunkIndex, Data: chunk.Data, Ok: true}
	}
	for i := 0; i < totalChunks; i++ {
		wg.Add(1)
		go requestChunk(i)
	}
	receivedCount := 0
	for receivedCount < totalChunks {
		res := <-chunkResultsChan
		if res.Ok {
			mu.Lock()
			progress.Received[res.Index] = true
			mu.Unlock()
		}
		receivedCount++
	}
	wg.Wait()
	fmt.Println("File transfer complete.")
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
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Received: make(map[int]bool)}
	var mu sync.Mutex
	var wg sync.WaitGroup
	chunkResultsChan := make(chan ChunkResult, progress.TotalChunks)

	if meta.Transport == QUIC_S {
		handleQuic(meta, &mu, &wg, chunkResultsChan, stream, progress)
	} else if meta.Transport == TCP_S {
		handleTcp(meta, &mu, &wg, chunkResultsChan, stream, progress, dec, conn)
	}
}
