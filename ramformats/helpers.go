package ramformats

// Function to generate UUID
import (
	"encoding/binary"
	"encoding/json"
	"fmt"
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

func PopFront(files *[]RamFile) (RamFile, bool) {
	if len(*files) == 0 {
		return RamFile{}, false
	}
	first := (*files)[0]
	*files = (*files)[1:]
	return first, true
}
