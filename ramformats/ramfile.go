package ramformats

import (
	"fmt"
	"os"
)

const (
	DRInputFile       = "inputFile"
	DROutputFile      = "outputFile"
	DRFileNameKey     = "filename"
	DRFileSizeKey     = "filesize"
	DRUUIDKey         = "uuid"
	DRSendStartKey    = "sendStartTimestamp"
	DRSendEndKey      = "sendEndTimestamp"
	DRRecieveStartKey = "receiveStartTimestamp"
	DRRecieveEndKey   = "receiveEndTimestamp"
	DRChunkSizeKey    = "chunkSize"
	DRNumChunks       = "numChunks"
)

// Should we just give a stream here instead of path?

type RamFile struct {
	LocalPath   string
	RamFileType string
	MetaData    map[string]string
	UUID        string
}

func NewRamFileFromLocal(localPath string, fileName string) *RamFile {
	rf := &RamFile{
		LocalPath:   localPath,
		RamFileType: DRInputFile,
		MetaData:    make(map[string]string),
		UUID:        GenerateUUID(),
	}
	fileInfo, err := os.Stat(localPath)
	if err == nil {
		rf.MetaData[DRFileSizeKey] = fmt.Sprintf("%d", fileInfo.Size())
	} else {
		return nil
	}
	// Add filename to Metadata
	rf.MetaData[DRFileNameKey] = fileName
	rf.MetaData[DRUUIDKey] = rf.UUID
	return rf
}

func NewRamFileFromMeta(metaData map[string]string) *RamFile {
	rf := &RamFile{
		LocalPath:   "",
		RamFileType: DROutputFile,
		MetaData:    metaData,
		UUID:        metaData[DRUUIDKey],
	}
	return rf
}
