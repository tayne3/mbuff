package mbuff

// Endian represents byte order for multi-byte values.
type Endian bool

const (
	// BigEndian represents big-endian byte order.
	BigEndian Endian = false
	// LittleEndian represents little-endian byte order.
	LittleEndian Endian = true
)
