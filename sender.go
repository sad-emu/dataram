package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Sender (Client)
func runSender(addr string, meta FileMetadata, stream Stream) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(meta); err != nil {
		fmt.Println("Metadata send error:", err)
		return
	}
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	progress := &Progress{TotalChunks: int((meta.Size + int64(meta.ChunkSize) - 1) / int64(meta.ChunkSize)), Sent: make(map[int]bool)}
	for {
		var req map[string]int
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Request decode error:", err)
			return
		}
		idx := req["request_chunk"]
		buf := make([]byte, meta.ChunkSize)
		stream.SeekAbsolute(int64(idx * meta.ChunkSize))
		n, _ := stream.Read(buf)
		chunk := struct {
			ChunkIndex int    `json:"chunk_index"`
			Data       []byte `json:"data"`
		}{ChunkIndex: idx, Data: buf[:n]}
		if err := enc.Encode(chunk); err != nil {
			fmt.Println("Chunk send error:", err)
			return
		}
		progress.Sent[idx] = true
		fmt.Printf("Sent chunk %d/%d\n", idx+1, progress.TotalChunks)
	}
	fmt.Println("File send complete.")
}
