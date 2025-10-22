package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHLSwap tests high-low byte swap for 32-bit and 64-bit values.
func TestHLSwap(t *testing.T) {
	m := NewMBuff(make([]byte, 16))
	m.SetHLSwap(true)

	// PutU32 with HLSwap: 0xAABBCCDD -> 0xBBAADDCC
	m.PutU32(0xAABBCCDD)
	// With BigEndian and HLSwap, byte order becomes:
	// [0xBB, 0xAA, 0xDD, 0xCC]
	expected32 := []byte{0xBB, 0xAA, 0xDD, 0xCC}
	assert.Equal(t, expected32, m.Bytes())
	m.Rewind()
	v32 := m.TakeU32()
	assert.Equal(t, uint32(0xAABBCCDD), v32)

	// PutU64 with HLSwap: 0x1122334455667788 -> 0x2211443366558877
	m.Clear()
	m.PutU64(0x1122334455667788)
	// With BigEndian and HLSwap, byte order becomes:
	// [0x22, 0x11, 0x44, 0x33, 0x66, 0x55, 0x88, 0x77]
	expected64 := []byte{0x22, 0x11, 0x44, 0x33, 0x66, 0x55, 0x88, 0x77}
	assert.Equal(t, expected64, m.Bytes())
	m.Rewind()
	v64 := m.TakeU64()
	assert.Equal(t, uint64(0x1122334455667788), v64)
}
