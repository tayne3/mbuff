package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEndian tests byte order handling.
func TestEndian(t *testing.T) {
	m := New(8)

	// Big Endian (default)
	m.PutU32(0x12345678)
	expectedBE := []byte{0x12, 0x34, 0x56, 0x78}
	assert.Equal(t, expectedBE, m.Bytes())

	// Little Endian
	m.Clear()
	m.SetEndian(LittleEndian)
	m.PutU32(0x12345678)
	expectedLE := []byte{0x78, 0x56, 0x34, 0x12}
	assert.Equal(t, expectedLE, m.Bytes())

	m.Rewind()
	v := m.TakeU32()
	assert.Equal(t, uint32(0x12345678), v)
}
