package ramformats

// Function to generate UUID
import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
)

func GenerateUUID() string {
	// Generate a new UUID
	return uuid.New().String()
}

func GetIntFromString(value string) (int64, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		// TODO log error
		return -1, fmt.Errorf("Error parsing value %s with error: %v", value, err)
	}
	i := int64(i64)
	return i, nil
}

func ExportMetaToBytes(meta map[string]map[string]string) ([]byte, error) {
	return json.Marshal(meta)
}

func BytesToExportMeta(data []byte) (map[string]map[string]string, error) {
	var meta map[string]map[string]string
	err := json.Unmarshal(data, &meta)
	return meta, err
}

func IntToBytes(intToConvert int) []byte {
	headerBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(headerBytes, uint32(intToConvert))
	return headerBytes
}

func BytesToInt(data []byte) int {
	return int(binary.BigEndian.Uint32(data))
}

func Int64ToBytes(intToConvert int64) []byte {
	headerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(headerBytes, uint64(intToConvert))
	return headerBytes
}

func BytesToInt64(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}

func PopFront(files *[]RamFile) (RamFile, bool) {
	if len(*files) == 0 {
		return RamFile{}, false
	}
	first := (*files)[0]
	*files = (*files)[1:]
	return first, true
}

func PrepareFile(path string, size int64) (*os.File, error) {
	// Open the file with read-write, create if not exists
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Get current file info
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Resize the file if it's smaller than desired
	if info.Size() < size {
		err = file.Truncate(size)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to resize file: %w", err)
		}
	}

	// Now file is ready for seeking and writing
	return file, nil
}
