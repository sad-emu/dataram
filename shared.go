package main

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	TCP_S  = "TCP"
	QUIC_S = "QUIC"
)

// FileMetadata holds metadata about the file to transfer
type FileMetadata struct {
	Uuid         string                 `json:"uuid"` // Unique identifier for the file transfer
	Filename     string                 `json:"filename"`
	FlowName     string                 `json:"flow_name"`
	Size         int64                  `json:"size"`
	ChunkSize    int                    `json:"chunk_size"`
	Availability map[int]bool           `json:"availability"` // Map of chunk index to availability. Not all chunks may be available
	Metadata     map[string]interface{} `json:"metadata"`
	Transport    string                 `json:"transport"`           // "TCP" or "QUIC"
	QuicAddr     string                 `json:"quic_addr,omitempty"` // Only set if using QUIC
}

// Progress tracks transfer progress
type Progress struct {
	TotalChunks int
	Received    map[int]bool
	Sent        map[int]bool
}

type ChunkResult struct {
	Index int
	Data  []byte
	Ok    bool
}

// ChunkResultList represents a variable list of ChunkResult
// Useful for passing or collecting multiple chunk results
// e.g. for batch operations or result aggregation
type ChunkResultList []ChunkResult

// ProgressManager tracks progress for multiple file transfers by UUID
// Used by the receiver/listener to support resumable/partial transfers
type ProgressManager struct {
	progressMap map[string]*Progress
}

func NewProgressManager() *ProgressManager {
	return &ProgressManager{progressMap: make(map[string]*Progress)}
}

func (pm *ProgressManager) GetOrCreate(uuid string, totalChunks int) *Progress {
	if prog, ok := pm.progressMap[uuid]; ok {
		return prog
	}
	prog := &Progress{
		TotalChunks: totalChunks,
		Received:    make(map[int]bool),
		Sent:        make(map[int]bool),
	}
	pm.progressMap[uuid] = prog
	return prog
}

func (pm *ProgressManager) Remove(uuid string) {
	delete(pm.progressMap, uuid)
}

type ActiveTransfer struct {
	Meta   FileMetadata
	Stream Stream
}

type ActiveTransfers struct {
	mu        sync.Mutex
	transfers map[string]*ActiveTransfer
}

func NewActiveTransfers() *ActiveTransfers {
	return &ActiveTransfers{transfers: make(map[string]*ActiveTransfer)}
}

func (at *ActiveTransfers) Get(uuid string) (*ActiveTransfer, bool) {
	at.mu.Lock()
	defer at.mu.Unlock()
	tr, ok := at.transfers[uuid]
	return tr, ok
}

func (at *ActiveTransfers) Add(uuid string, meta FileMetadata, stream Stream) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.transfers[uuid] = &ActiveTransfer{Meta: meta, Stream: stream}
}

func (at *ActiveTransfers) Remove(uuid string) {
	at.mu.Lock()
	defer at.mu.Unlock()
	delete(at.transfers, uuid)
}

type Settings struct {
	StorageDir string `json:"storage_dir"`
}

var GlobalSettings Settings

func LoadSettings(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	return dec.Decode(&GlobalSettings)
}
