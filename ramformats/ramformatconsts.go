package ramformats

// Raw bytes for the export bundle headers
var (
	DATARAM_EXPORT_BUNDLE_HEADER_1 = []byte{0xda, 0x1a, 0xbe, 0x01} // DATA Bundle Export v1 header
	DATARAM_EXPORT_BUNDLE_HEADER_2 = []byte{0xda, 0x1a, 0xbe, 0x02} // DATA Bundle Export v2 header
)

// These are converted to ints
const (
	METADATA_HEADER = 10
	DATA_HEADER     = 12
)
