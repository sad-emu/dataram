package ramformats

import (
	"fmt"
	"sync"
)

type RamBundle struct {
	fileQueue      []RamFile
	exportBundle   []RamFile
	exportMeta     map[string]string // Metadata for the export bundle
	exportStatus   map[string]string // Status of each file in the export bundle
	maxQueueSize   int               // Maximum size of the queue
	chunkSize      int
	maxBundleCount int
	mu             sync.Mutex // Mutex to protect concurrent access
}

func NewRamBundle(chunkSize int, maxBundleCount int, maxQueueSize int) *RamBundle {
	return &RamBundle{
		fileQueue:      make([]RamFile, 0),
		exportBundle:   make([]RamFile, 0),
		chunkSize:      chunkSize,
		maxBundleCount: maxBundleCount,
	}
}

func (rb *RamBundle) AddFile(rf RamFile) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.fileQueue) < rb.maxQueueSize {
		rb.fileQueue = append(rb.fileQueue, rf)
		return nil
	} else {
		return fmt.Errorf("RamBundle queue is full, cannot add more files until some are processed")
	}
}

func (rb *RamBundle) GetNextBundle() ([]byte, error) {
	// Check to see if the previous bundle is complete

	// If old bundle is complete, from file queue to export queue

	// Build metadata if new bundle is being created

	// Get and return the next chunk

	return nil, fmt.Errorf("GetNextBundle not implemented yet")
}
