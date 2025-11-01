package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPutTake tests writing and reading scalar values.
func TestPutTake(t *testing.T) {
	m := New(0)

	// Put various scalar types
	m.PutU8(0x01)
	m.PutU8(0xFE)
	m.PutU16(0x0304) // Big Endian
	m.PutU16(0xFFF6)
	m.PutU32(0x05060708)
	m.PutU32(0xFFFFFFEC)
	m.PutU64(0x090A0B0C0D0E0F10)
	m.PutU64(0xFFFFFFFFFFFFFFE2)

	// Check underlying bytes
	expected := []byte{
		0x01, 0xFE, // u8, u8
		0x03, 0x04, // u16
		0xFF, 0xF6, // u16
		0x05, 0x06, 0x07, 0x08, // u32
		0xFF, 0xFF, 0xFF, 0xEC, // u32
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // u64
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xE2, // u64
	}
	assert.Equal(t, expected, m.Bytes())

	// Take scalar types
	m.Rewind()
	assert.Equal(t, uint8(0x01), m.TakeU8())
	assert.Equal(t, uint8(0xFE), m.TakeU8())
	assert.Equal(t, uint16(0x0304), m.TakeU16())
	assert.Equal(t, uint16(0xFFF6), m.TakeU16())
	assert.Equal(t, uint32(0x05060708), m.TakeU32())
	assert.Equal(t, uint32(0xFFFFFFEC), m.TakeU32())
	assert.Equal(t, uint64(0x090A0B0C0D0E0F10), m.TakeU64())
	assert.Equal(t, uint64(0xFFFFFFFFFFFFFFE2), m.TakeU64())
}

// TestPutTakeArr tests array read/write operations.
func TestPutTakeArr(t *testing.T) {
	m := New(0)

	// Arr8
	arr8 := []byte{1, 2, 3, 4}
	m.PutArr8(arr8)
	assert.Equal(t, 4, m.Pos())
	assert.Equal(t, 4, m.Count())

	out8 := make([]byte, 4)
	m.Rewind()
	m.TakeArr8(out8)
	assert.Equal(t, arr8, out8)

	// Arr32 with endianness and HLSwap
	m.Clear()
	m.SetEndian(LittleEndian)
	m.SetHLSwap(true)
	arr32 := []uint32{0x11223344, 0xAABBCCDD}
	m.PutArr32(arr32)

	// With LittleEndian and HLSwap:
	// 0x11223344 -> HLSwap -> 0x22114433 -> LE -> [0x33, 0x44, 0x11, 0x22]
	// 0xAABBCCDD -> HLSwap -> 0xBBAADDCC -> LE -> [0xCC, 0xDD, 0xAA, 0xBB]
	expected32 := []byte{0x33, 0x44, 0x11, 0x22, 0xCC, 0xDD, 0xAA, 0xBB}
	assert.Equal(t, expected32, m.Bytes())

	out32 := make([]uint32, 2)
	m.Rewind()
	m.TakeArr32(out32)
	assert.Equal(t, arr32[0], out32[0])
	assert.Equal(t, arr32[1], out32[1])
}
