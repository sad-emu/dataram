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

// AcceptAndHandleLoop accepts up to maxLoops connections and handles them. If maxLoops is 0, loops forever.
func AcceptAndHandleLoop(addr string, stream Stream, tlsConf *tls.Config, maxLoops int) error {
	ln, err := tls.Listen("tcp", addr, tlsConf)
	if err != nil {
		return err
	}
	defer ln.Close()
	count := 0
	for {
		if maxLoops > 0 && count >= maxLoops {
			break
		}
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		err = handleSender(conn, stream, tlsConf)
		conn.Close()
		if err != nil {
			return err
		}
		count++
	}
	return nil
}

func handleQuic(meta FileMetadata, mu *sync.Mutex, wg *sync.WaitGroup, chunkResultsChan chan ChunkResult, stream Stream, progress *Progress, tlsConf *tls.Config) error {
	session, err := quic.DialAddr(context.Background(), meta.QuicAddr, tlsConf, nil)
	if err != nil {
		return fmt.Errorf("listener QUIC dial error: %w", err)
	}
	defer session.CloseWithError(0, "")
	totalChunks := progress.TotalChunks
	availableCount := 0
	for i := 0; i < totalChunks; i++ {
		if meta.Availability == nil || meta.Availability[i] {
			availableCount++
		}
	}
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
			Ok         bool   `json:"ok"`
		}
		if err := json.NewDecoder(streamRecv).Decode(&chunk); err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if !chunk.Ok {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if err := stream.SeekAbsolute(int64(chunk.ChunkIndex * meta.ChunkSize)); err == nil {
			stream.Write(chunk.Data)
		}
		chunkResultsChan <- ChunkResult{Index: chunk.ChunkIndex, Data: chunk.Data, Ok: true}
	}
	// Only request available and not-yet-received chunks
	for i := 0; i < totalChunks; i++ {
		if meta.Availability != nil && !meta.Availability[i] {
			continue
		}
		mu.Lock()
		if progress.Received != nil && progress.Received[i] {
			mu.Unlock()
			continue
		}
		mu.Unlock()
		wg.Add(1)
		go requestChunk(i)
	}
	receivedCount := 0
	for receivedCount < availableCount {
		res := <-chunkResultsChan
		if res.Ok {
			mu.Lock()
			progress.Received[res.Index] = true
			mu.Unlock()
		}
		receivedCount++
	}
	wg.Wait()
	fmt.Println("Listener QUIC File transfer complete.")
	return nil
}

func handleTcp(meta FileMetadata, mu *sync.Mutex, wg *sync.WaitGroup, chunkResultsChan chan ChunkResult, stream Stream, progress *Progress, dec *json.Decoder, conn net.Conn, tlsConf *tls.Config) error {
	totalChunks := progress.TotalChunks
	availableCount := 0
	for i := 0; i < totalChunks; i++ {
		if meta.Availability == nil || meta.Availability[i] {
			availableCount++
		}
	}
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
			Ok         bool   `json:"ok"`
		}
		mu.Lock()
		err = dec.Decode(&chunk)
		mu.Unlock()
		if err != nil {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if !chunk.Ok {
			chunkResultsChan <- ChunkResult{Index: i, Data: nil, Ok: false}
			return
		}
		if err := stream.SeekAbsolute(int64(chunk.ChunkIndex * meta.ChunkSize)); err == nil {
			stream.Write(chunk.Data)
		}
		chunkResultsChan <- ChunkResult{Index: chunk.ChunkIndex, Data: chunk.Data, Ok: true}
	}
	// Only request available and not-yet-received chunks
	for i := 0; i < totalChunks; i++ {
		if meta.Availability != nil && !meta.Availability[i] {
			continue
		}
		mu.Lock()
		if progress.Received != nil && progress.Received[i] {
			mu.Unlock()
			continue
		}
		mu.Unlock()
		wg.Add(1)
		go requestChunk(i)
	}
	receivedCount := 0
	for receivedCount < availableCount {
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
	return nil
}

// Helper to load the server certificate
func loadListenerCert() tls.Certificate {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic("failed to load server TLS cert: " + err.Error())
	}
	return cert
}

func handleSender(conn net.Conn, stream Stream, tlsConf *tls.Config) error {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	var meta FileMetadata
	if err := dec.Decode(&meta); err != nil {
		return fmt.Errorf("metadata decode error: %w", err)
	}
	fmt.Printf("Receiving file: %s (%d bytes)\n", meta.Filename, meta.Size)
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Received: make(map[int]bool)}
	var mu sync.Mutex
	var wg sync.WaitGroup
	chunkResultsChan := make(chan ChunkResult, progress.TotalChunks)

	if meta.Transport == QUIC_S {
		return handleQuic(meta, &mu, &wg, chunkResultsChan, stream, progress, tlsConf)
	} else if meta.Transport == TCP_S {
		return handleTcp(meta, &mu, &wg, chunkResultsChan, stream, progress, dec, conn, tlsConf)
	}
	return nil
}
