package ramformats

var (
	DATARAM_EXPORT_BUNDLE_HEADER_1 = []byte{0xda, 0x1a, 0xbe, 0x01} // DATA Bundle Export as byte array // DATA Bundle 1
	DATARAM_EXPORT_BUNDLE_HEADER_2 = []byte{0xda, 0x1a, 0xbe, 0x02} // DATA Bundle Export as byte array // DATA Bundle 1
)

const (
	METADATA_HEADER = 10 // Size of the header for int64 response
	DATA_HEADER     = 12
)
