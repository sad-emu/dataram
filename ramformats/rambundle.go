package ramformats

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

// Define const header bytes
const (
	METADATA_HEADER = 10 // Size of the header for int64 response
	DATA_HEADER     = 12
)

// The purpose of RamBundle is to take a collection of/a single RamFile(s)
// and bundle or split them into set chunks for transportation.

// The idea is a bundle will have two types of output
// 1. A Metadata map that contains information about the files in the bundle
// 2. A data block that contains the actual data of the files in the bundle
// Multiple data blocks may result from a single bundle. Only a single metadata map will be created for each bundle.

type RamBundle struct {
	fileQueue      []RamFile
	exportBundle   []RamFile
	exportMeta     map[string]map[string]string // Metadata for the export bundle (map of string to map of strings)
	exportStatus   map[string]string            // Status of each file in the export bundle
	exportFinished bool                         // Flag to indicate if the export is finished
	sentMetaData   bool                         // Flag to indicate if metadata has been sent
	bundlesSent    int64                        // Number of bundles sent
	totalBundles   int64                        // Total number of bundles to be sent
	maxQueueSize   int                          // Maximum size of the queue
	chunkSize      int64
	maxBundleCount int
	mu             sync.Mutex // Mutex to protect concurrent access
}

func popFront(files *[]RamFile) (RamFile, bool) {
	if len(*files) == 0 {
		return RamFile{}, false
	}
	first := (*files)[0]
	*files = (*files)[1:]
	return first, true
}

func NewRamBundle(chunkSize int64, maxBundleCount int, maxQueueSize int) *RamBundle {
	return &RamBundle{
		fileQueue:      make([]RamFile, 0),
		exportBundle:   make([]RamFile, 0),
		chunkSize:      chunkSize,
		maxBundleCount: maxBundleCount,
		exportFinished: true,
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
	// Get and return the next chunk
	if rb.bundlesSent >= rb.totalBundles {
		rb.exportFinished = true // All bundles have been sent
		rb.exportStatus = make(map[string]string)
		rb.exportBundle = rb.exportBundle[:0] // Clear the export bundle
	}

	// Check to see if the previous bundle is complete
	if rb.exportFinished {
		rb.mu.Lock()
		defer rb.mu.Unlock()
		newBundleSize := int64(0)
		newBundleCount := 0
		// loop over each file in the queue
		for i := 0; i < len(rb.fileQueue); i++ {
			rf, ok := popFront(&rb.fileQueue)
			if !ok {
				break // No more files to process
			}
			sizeVal, err := GetIntFromString(rf.MetaData[DRFileSizeKey])
			if err != nil || sizeVal <= 0 {
				return nil, fmt.Errorf("Error parsing file %s cannot add file to queue: %v", rf.LocalPath, err)
			}

			newBundleSize += sizeVal
			newBundleCount += 1

			rb.exportBundle = append(rb.exportBundle, rf)

			if newBundleSize >= rb.chunkSize || newBundleCount >= rb.maxBundleCount {
				// If the bundle is full, break out of the loop
				break
			}
		}

		if newBundleCount != 0 {
			rb.exportFinished = false // We have a new bundle to export
			rb.bundlesSent = 0

			// Build metadata for the new bundle
			rb.exportMeta = make(map[string]map[string]string)
			for i := 0; i < len(rb.exportBundle); i++ {
				rf := rb.exportBundle[i]
				rb.exportMeta[rf.UUID] = make(map[string]string)
				rb.exportMeta[rf.UUID][DRFileNameKey] = rf.MetaData[DRFileNameKey]
				rb.exportMeta[rf.UUID][DRFileSizeKey] = rf.MetaData[DRFileSizeKey]
				rb.exportMeta[rf.UUID][DRUUIDKey] = rf.UUID
				rb.exportMeta[rf.UUID][DRSendStartKey] = time.Now().Format(time.RFC3339)
				rb.exportMeta[rf.UUID][DRChunkSizeKey] = strconv.FormatInt(rb.chunkSize, 10)
			}
			rb.sentMetaData = false

			// How many bundles do we need to send
			rb.totalBundles = newBundleSize / rb.chunkSize
		} else {
			// No more files to process, return nil
			return nil, nil
		}
	}

	// Return bytes of the exportMeta.
	if !rb.sentMetaData {
		// Convert metadata to bytes
		bytes, err := ExportMetaToBytes(rb.exportMeta)
		if err != nil {
			return nil, fmt.Errorf("Error converting metadata to bytes: %v", err)
		}
		rb.sentMetaData = true // Metadata has been sent

		// Append MetaData header to the start of the bytes
		headerBytes := IntToBytes(METADATA_HEADER)
		bytes = append(headerBytes, bytes...)

		return bytes, nil
	}

	bytesBundle := make([]byte, 0)
	headerBytes := IntToBytes(DATA_HEADER)
	bytesBundle = append(headerBytes, bytesBundle...)
	sentBytes := int64(rb.chunkSize) * int64(rb.bundlesSent)
	countBytes := int64(0)
	lastFileBytes := int64(0)
	for i := 0; i < len(rb.exportBundle); i++ {
		// Iterate over the files until we reach sentBytes
		rf := rb.exportBundle[i]
		sizeVal, err := GetIntFromString(rf.MetaData[DRFileSizeKey])
		if err != nil || sizeVal <= 0 {
			return nil, fmt.Errorf("Error parsing file %s cannot add file to bundle: %v", rf.LocalPath, err)
		}
		lastFileBytes = countBytes
		countBytes += int64(sizeVal)
		if countBytes <= sentBytes {
			// Skip this file, we have already sent it
			continue
		}

		relativePosition := countBytes - lastFileBytes
		amountToRead := lastFileBytes + rb.chunkSize
		if amountToRead > sizeVal {
			amountToRead = sizeVal
		}

		bundleChunk := make([]byte, amountToRead)
		// open a filestream for rf

		fileHandle, err := os.Open(rf.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		n, err := fileHandle.ReadAt(bundleChunk, relativePosition)
		if err != nil {
			if err == io.EOF {
				return nil, nil // Return number of bytes read before EOF
			}
			return nil, fmt.Errorf("failed to read from file: %w", err)
		}
		countBytes += int64(n)

		// append bytesBundle to bundleChunk\
		// TODO format (UUID, start_pos, len, Datablob)
		uuidBytes := []byte(rf.UUID)
		startPos := Int64ToBytes(relativePosition)
		len := IntToBytes(n)
		bytesBundle = append(bytesBundle, uuidBytes...)
		bytesBundle = append(bytesBundle, startPos...)
		bytesBundle = append(bytesBundle, len...)
		bytesBundle = append(bytesBundle, bundleChunk...)
	}

	return nil, fmt.Errorf("GetNextBundle not implemented yet")
}
