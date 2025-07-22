package ramformats

import (
	"bytes"
	"fmt"
	"sync"
)

// Define const header bytes

// The purpose of RamBundle is to take chunks of ramexportbundles
// and piece them back into ramfiles

// The bundles have two types of output
// 1. A Metadata map that contains information about the files in the bundle
// 2. A data block that contains the actual data of the files in the bundle
// Multiple data blocks may result from a single bundle. Only a single metadata map will be created for each set of bundles.

// State will need to be saved and stored for this class for restarts

type RamImportBundle struct {
	processingDirectory string
	filePartsQueue      []byte
	processBundles      map[string]RamFile
	completedFiles      []RamFile

	maxQueueSize   int // Maximum size of the queue
	chunkSize      int64
	maxBundleCount int
	mu             sync.Mutex // Mutex to protect concurrent access
}

func NewRamImportBundle(chunkSize int64, maxBundleCount int, maxQueueSize int, processingDir string) *RamImportBundle {
	return &RamImportBundle{
		processBundles:      make(map[string]RamFile),
		completedFiles:      make([]RamFile, 0),
		filePartsQueue:      make([]byte, 0),
		chunkSize:           chunkSize,
		maxBundleCount:      maxBundleCount,
		maxQueueSize:        maxQueueSize,
		processingDirectory: processingDir,
	}
}

func (rb *RamImportBundle) PopFile(rf RamFile) *RamFile {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if len(rb.completedFiles) != 0 {
		rf, _ := PopFront(&rb.completedFiles)
		return &rf
	} else {
		return nil
	}
}

func (rb *RamImportBundle) ProcessNextExportBundle(dataIn []byte) error {
	// Verify data has a valid header
	blockHeader := dataIn[0:4]
	if !bytes.Equal(blockHeader, DATARAM_EXPORT_BUNDLE_HEADER_1) {
		return fmt.Errorf("Error parsing data. Unrecognised block header: %d", blockHeader)
	}

	typeHeader := BytesToInt(dataIn[4:8])
	if typeHeader == METADATA_HEADER {
		// if it's a metadata bundle
		metadataHeader, err := BytesToExportMeta(dataIn[8:])
		if err != nil {
			return fmt.Errorf("Error parsing ram export meta map, %s", err)
		}

		for k, v := range metadataHeader {
			fmt.Printf("key[%s] value[%s]\n", k, v)
			_, exists := rb.processBundles[k]
			if !exists {
				ramFile := *NewRamFileFromMeta(v)
				// TODO this
				ramFile.LocalPath = "/tmp/" + k
				rb.processBundles[k] = ramFile
			} else {
				// Already exists from a data packet first, just update metadata
				for kk, vv := range v {
					rb.processBundles[k].MetaData[kk] = vv
				}
			}

		}
		// create RamFiles and add into process bundles
		// Check to see if they exist first with uuid checks, if they exist just update metadata

	} else if typeHeader == DATA_HEADER {
		// First get the uuid
		// check to see if exists
		// Create RamFile if not with uuid, tmp path = ramFile.LocalPath = "/tmp/" + uuid
		// read data start point
		// read data length
		// open file on disk, create if doesnt exist
		// write from start point for length
		return nil
	} else {
		return fmt.Errorf("Error parsing data. Unrecognised type header: %d", typeHeader)
	}
	return fmt.Errorf("Unreachable code reached.")
}
