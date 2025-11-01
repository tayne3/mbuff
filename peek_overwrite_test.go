package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPeek tests peeking scalar values without advancing position.
func TestPeek(t *testing.T) {
	m := New(16)
	m.PutU8(0x01)
	m.PutU8(0x02)
	m.PutU32(0x12345678)
	m.pos = 1

	// Peek from pos+offset
	assert.Equal(t, uint8(0x02), m.PeekU8(0))
	assert.Equal(t, uint32(0x12345678), m.PeekU32(1))

	// Verify position unchanged
	assert.Equal(t, 1, m.Pos())
}

// TestOverwrite tests overwriting values at specific offsets.
func TestOverwrite(t *testing.T) {
	m := New(16)
	m.PutU32(0xAAAAAAAA)
	m.PutU32(0xBBBBBBBB)

	// Overwrite at absolute offsets
	m.OverwriteU8(0, 0xDE)
	m.OverwriteU32(4, 0xDEADBEEF)

	// Verify position unchanged
	assert.Equal(t, 8, m.Pos())

	expected := []byte{0xDE, 0xAA, 0xAA, 0xAA, 0xDE, 0xAD, 0xBE, 0xEF}
	assert.Equal(t, expected, m.Bytes())

	m.Rewind()
	assert.Equal(t, uint32(0xDEAAAAAA), m.TakeU32())
	assert.Equal(t, uint32(0xDEADBEEF), m.TakeU32())
}

// TestPeekArr tests array peek operations.
func TestPeekArr(t *testing.T) {
	m := New(64)
	m.PutU8(0xFF)
	m.PutU32(0x11223344)
	m.PutU32(0x55667788)
	m.pos = 1

	out32 := make([]uint32, 2)
	m.PeekArr32(0, out32)

	assert.Equal(t, 1, m.Pos())
	assert.Equal(t, uint32(0x11223344), out32[0])
	assert.Equal(t, uint32(0x55667788), out32[1])

	// Test with offset
	m.PeekArr32(4, out32[0:1])
	assert.Equal(t, uint32(0x55667788), out32[0])
}

// TestOverwriteArr tests array overwrite operations.
func TestOverwriteArr(t *testing.T) {
	m := New(64)
	m.PutArr32([]uint32{0xAAAAAAAA, 0xBBBBBBBB, 0xCCCCCCCC})

	// Overwrite at absolute offset
	in32 := []uint32{0xDEADBEEF, 0xCAFEBABE}
	m.OverwriteArr32(4, in32)

	assert.Equal(t, 12, m.Pos())

	expected := []byte{
		0xAA, 0xAA, 0xAA, 0xAA, // 0xAAAAAAAA
		0xDE, 0xAD, 0xBE, 0xEF, // 0xDEADBEEF
		0xCA, 0xFE, 0xBA, 0xBE, // 0xCAFEBABE
	}
	assert.Equal(t, expected, m.Bytes())
}
