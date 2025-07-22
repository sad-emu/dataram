package ramformats

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

// The purpose of RamBundle is to take a collection of/a single RamFile(s)
// and bundle or split them into set chunks for transportation.

// The idea is a bundle will have two types of output
// 1. A Metadata map that contains information about the files in the bundle
// 2. A data block that contains the actual data of the files in the bundle
// Multiple data blocks may result from a single bundle. Only a single metadata map will be created for each bundle.

// State will need to be saved and stored for this class for restarts

type RamExportBundle struct {
	fileInboundQueue []RamFile
	exportBundle     []RamFile
	exportMeta       map[string]map[string]string // Metadata for the export bundle (map of string to map of strings)
	// exportBundleMeta []BundleMeta                 // Metadata of each package
	exportFinished bool  // Flag to indicate if the export is finished
	sentMetaData   bool  // Flag to indicate if metadata has been sent
	bundlesSent    int64 // Number of bundles sent
	totalBundles   int64 // Total number of bundles to be sent
	maxQueueSize   int   // Maximum size of the queue
	chunkSize      int64
	maxBundleCount int
	mu             sync.Mutex // Mutex to protect concurrent access
}

func NewRamExportBundle(chunkSize int64, maxBundleCount int, maxQueueSize int) *RamExportBundle {
	return &RamExportBundle{
		fileInboundQueue: make([]RamFile, 0),
		exportBundle:     make([]RamFile, 0),
		chunkSize:        chunkSize,
		maxBundleCount:   maxBundleCount,
		maxQueueSize:     maxQueueSize,
		exportFinished:   true,
	}
}

func (rb *RamExportBundle) PushFile(rf RamFile) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.fileInboundQueue) < rb.maxQueueSize {
		rb.fileInboundQueue = append(rb.fileInboundQueue, rf)
		return nil
	} else {
		return fmt.Errorf("RamBundle queue is full, cannot add more files until some are processed")
	}
}

func (rb *RamExportBundle) GetNextExportBundle() ([]byte, error) {
	// Get and return the next chunk
	if rb.bundlesSent >= rb.totalBundles {
		rb.exportFinished = true // All bundles have been sent
		// rb.exportBundleMeta = nil
		rb.exportBundle = rb.exportBundle[:0] // Clear the export bundle
	}

	// Check to see if the previous bundle is complete
	if rb.exportFinished {
		rb.mu.Lock()
		defer rb.mu.Unlock()
		newBundleSize := int64(0)
		newBundleCount := 0
		// loop over each file in the queue
		for i := 0; i < len(rb.fileInboundQueue); i++ {
			rf, ok := PopFront(&rb.fileInboundQueue)
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
			// I don't kow how to do division on 64bit ints
			for newBundleSize > 0 {
				rb.totalBundles += 1
				newBundleSize -= rb.chunkSize
			}

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
		headerBytes = append(DATARAM_EXPORT_BUNDLE_HEADER_1, headerBytes...)
		bytes = append(headerBytes, bytes...)

		return bytes, nil
	}

	bytesBundle := make([]byte, 0)
	headerBytes := IntToBytes(DATA_HEADER)
	headerBytes = append(DATARAM_EXPORT_BUNDLE_HEADER_1, headerBytes...)
	bytesBundle = append(headerBytes, bytesBundle...)
	// Bytes we have sent on the previous bundle
	bundleTotalPosition := int64(rb.chunkSize) * int64(rb.bundlesSent)
	bundleRelativePosition := int64(0)
	thisBundleBytes := int64(0)
	// // Bytes we have sent in the current bundle
	// currentFilePosition := int64(0)
	// What pos in the current file do we start to read from

	for i := 0; i < len(rb.exportBundle); i++ {
		// Iterate over the files until we reach sentBytes
		rf := rb.exportBundle[i]
		sizeVal, err := GetIntFromString(rf.MetaData[DRFileSizeKey])
		if err != nil || sizeVal <= 0 {
			return nil, fmt.Errorf("Error parsing file %s cannot add file to bundle: %v", rf.LocalPath, err)
		}

		// If we write this file out and we still havent caught up to our relative position
		// This file must be completed
		if bundleRelativePosition+sizeVal < bundleTotalPosition {
			bundleRelativePosition = bundleRelativePosition + sizeVal
			continue
		}

		currentFilePosition := bundleTotalPosition - bundleRelativePosition

		// Don't over-read if we are partway through a file
		amountToRead := rb.chunkSize - thisBundleBytes
		if (amountToRead + currentFilePosition) > sizeVal {
			amountToRead = (sizeVal - currentFilePosition)
		}

		bundleChunk := make([]byte, amountToRead)
		// open a filestream for rf

		fileHandle, err := os.Open(rf.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		n, err := fileHandle.ReadAt(bundleChunk, currentFilePosition)
		if err != nil {
			if err == io.EOF {
				return nil, nil // Return number of bytes read before EOF
			}
			return nil, fmt.Errorf("failed to read from file: %w", err)
		}

		// append bytesBundle to bundleChunk\
		// TODO format (UUID, start_pos, len, Datablob)
		uuidBytes := []byte(rf.UUID)
		if len(uuidBytes) != UUID_LEN {
			return nil, fmt.Errorf("Invalid UUID length: %d", len(uuidBytes))
		}
		totalFileSize := Int64ToBytes(sizeVal)
		startPos := Int64ToBytes(currentFilePosition)
		len := IntToBytes(n)
		bytesBundle = append(bytesBundle, uuidBytes...)
		bytesBundle = append(bytesBundle, totalFileSize...)
		bytesBundle = append(bytesBundle, startPos...)
		bytesBundle = append(bytesBundle, len...)
		bytesBundle = append(bytesBundle, bundleChunk...)
		thisBundleBytes += int64(n)

		// Update relative position
		bundleRelativePosition += int64(n)
		bundleTotalPosition += int64(n)
	}

	rb.bundlesSent += 1
	return bytesBundle, nil
}
