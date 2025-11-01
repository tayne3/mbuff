// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew tests buffer creation with specified capacity.
func TestNew(t *testing.T) {
	// Test with initial capacity
	m := New(100)

	assert.Equal(t, 100, m.Capacity())
	assert.Equal(t, 100, len(m.Data()))
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 0, m.Count())
	assert.True(t, m.IsEmpty())
	// Verify default byte order is BigEndian
	assert.Equal(t, binary.BigEndian, m.order)

	// Test writing to the buffer
	m.PutU32(0x12345678)
	assert.Equal(t, 4, m.Count())
	assert.Equal(t, 4, m.Pos())

	// Test with zero capacity
	m2 := New(0)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	// Should auto-grow when writing
	m2.PutU8(0xFF)
	assert.Equal(t, 1, m2.Count())
	assert.True(t, m2.Capacity() >= 1)
}

// TestNewMBuff tests wrapping existing buffer with data for parsing.
func TestNewMBuff(t *testing.T) {
	// Test with buffer containing data
	data := []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x01}
	m := NewMBuff(data)

	// Verify initial state
	assert.Equal(t, 8, m.Capacity())
	assert.Equal(t, 8, m.Count())
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 8, m.Readable())
	assert.False(t, m.IsEmpty())

	// Test reading the existing data
	v1 := m.TakeU16()
	assert.Equal(t, uint16(0x1234), v1)
	assert.Equal(t, 2, m.Pos())
	assert.Equal(t, 6, m.Readable())

	v2 := m.TakeU32()
	assert.Equal(t, uint32(0x5678ABCD), v2)
	assert.Equal(t, 6, m.Pos())
	assert.Equal(t, 2, m.Readable())

	// Test with empty buffer
	emptyData := []byte{}
	m2 := NewMBuff(emptyData)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	assert.True(t, m2.IsEmpty())

	// Test with buffer that has capacity > length
	b := make([]byte, 5, 10)
	b[0], b[1], b[2], b[3], b[4] = 0x01, 0x02, 0x03, 0x04, 0x05
	m3 := NewMBuff(b)
	assert.Equal(t, 10, m3.Capacity())
	assert.Equal(t, 5, m3.Count())
	assert.Equal(t, 5, m3.Readable())

	// Verify the actual data
	assert.Equal(t, byte(0x01), m3.TakeU8())
	assert.Equal(t, byte(0x02), m3.TakeU8())
	assert.Equal(t, byte(0x03), m3.TakeU8())
	assert.Equal(t, byte(0x04), m3.TakeU8())
	assert.Equal(t, byte(0x05), m3.TakeU8())
	assert.Equal(t, 0, m3.Readable())

	// Test that modifications affect original buffer
	original := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	m4 := NewMBuff(original)
	m4.OverwriteU8(0, 0xFF)
	assert.Equal(t, byte(0xFF), original[0]) // original should be modified
}

// TestStatusAndCapacity tests buffer status queries and capacity management.
func TestStatusAndCapacity(t *testing.T) {
	m := New(20)

	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 0, m.Count())
	assert.Equal(t, 20, m.Capacity())
	assert.Equal(t, 0, m.Readable())
	assert.Equal(t, 20, m.Writable())
	assert.Equal(t, 20, m.Appendable())
	assert.True(t, m.IsEmpty())
	assert.False(t, m.IsFull())

	m.PutU8(1)
	m.PutU8(2)
	m.PutU8(3)
	m.PutU8(4)
	m.PutU8(5)

	assert.Equal(t, 5, m.Pos())
	assert.Equal(t, 5, m.Count())
	assert.Equal(t, 0, m.Readable()) // position at end
	assert.Equal(t, 15, m.Writable())
	assert.Equal(t, 15, m.Appendable())
	assert.False(t, m.IsEmpty())

	m.Rewind()
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 5, m.Readable())
	assert.Equal(t, 20, m.Writable())
	assert.Equal(t, 15, m.Appendable())

	m.pos = 5
	m.data = m.data[:20]
	assert.Equal(t, 0, m.Appendable())
	assert.True(t, m.IsFull())
}

// TestDataAndBytes tests data access methods.
func TestDataAndBytes(t *testing.T) {
	m := New(10)
	m.PutU8(0xAA)
	m.PutU8(0xBB)
	m.PutU8(0xCC)
	m.Rewind()
	m.PutU8(0xDD) // overwrite
	m.pos = 2

	// Data() returns the entire underlying slice
	assert.Equal(t, 10, len(m.Data()))
	// Bytes() returns valid data [0:count]
	assert.Equal(t, []byte{0xDD, 0xBB, 0xCC}, m.Bytes())
	// ReadableBytes() returns readable data [pos:count]
	assert.Equal(t, []byte{0xCC}, m.ReadableBytes())
}

// TestPointerAndConfig tests position control methods.
func TestPointerAndConfig(t *testing.T) {
	m := New(10)
	m.PutU8(1)
	m.PutU8(2)
	m.PutU8(3)

	m.Rewind()
	assert.Equal(t, 0, m.Pos())

	m.pos = 1

	// Seek
	assert.NoError(t, m.Seek(2))
	assert.Equal(t, 2, m.Pos())

	// Seek out of bounds
	assert.Error(t, m.Seek(4))
	assert.Error(t, m.Seek(-1))

	// Reseek
	assert.NoError(t, m.Reseek(1)) // move to count - 1 = 2
	assert.Equal(t, 2, m.Pos())

	assert.NoError(t, m.Reseek(3)) // move to count - 3 = 0
	assert.Equal(t, 0, m.Pos())

	// Reseek out of bounds
	assert.Error(t, m.Reseek(4))

	// Skip
	m.Rewind()
	n := m.Skip(2)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, m.Pos())

	n = m.Skip(5) // can only skip 1 more
	assert.Equal(t, 1, n)
	assert.Equal(t, 3, m.Pos())

	n = m.Skip(1) // nothing left to skip
	assert.Equal(t, 0, n)

	// Clear
	m.Clear()
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 0, m.Count())
}

// TestCompact tests buffer compaction.
func TestCompact(t *testing.T) {
	m := New(10)

	// Use Put operations to set up data
	m.PutU8(0xAA)
	m.PutU8(0xBB)
	m.PutU8(0xCC)
	m.PutU8(0xDD)
	m.PutU8(0xEE)
	assert.Equal(t, 5, len(m.data))

	// Case 1: pos = 0 (no operation needed)
	m.pos = 0
	m.Compact()
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 5, m.Count())
	assert.Equal(t, byte(0xBB), m.data[1])

	// Case 2: 0 < pos < count (should move data)
	m.pos = 2
	m.Compact()
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 3, m.Count())
	assert.Equal(t, byte(0xCC), m.data[0])
	assert.Equal(t, byte(0xDD), m.data[1])
	assert.Equal(t, byte(0xEE), m.data[2])

	// Case 3: pos = count (should clear)
	m.pos = 3
	m.Compact()
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 0, m.Count())
}

// TestReadWriteFillPeek tests io.Reader/Writer interface and Fill method.
func TestReadWriteFillPeek(t *testing.T) {
	m := New(10)
	writeData := []byte{1, 2, 3, 4, 5}

	// Write
	n, err := m.Write(writeData)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, m.Pos())
	assert.Equal(t, 5, m.Count())

	// Read
	m.Rewind()
	readData := make([]byte, 3)
	n, err = m.Read(readData)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, m.Pos())
	assert.Equal(t, []byte{1, 2, 3}, readData)

	// Peek
	peekData := make([]byte, 2)
	n = m.Peek(peekData)
	assert.Equal(t, 2, n)
	assert.Equal(t, 3, m.Pos())
	assert.Equal(t, []byte{4, 5}, peekData)

	// Read partial
	n, err = m.Read(readData) // can only read 2 bytes now
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 5, m.Pos())
	assert.Equal(t, []byte{4, 5}, readData[0:2])

	// Read EOF
	n, err = m.Read(readData)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)

	// Write with auto-grow (no longer returns ErrShortWrite)
	m.pos = 8
	m.data = m.data[:8]
	n, err = m.Write(writeData) // Will auto-grow to accommodate
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 13, m.Pos())
	assert.Equal(t, 13, m.Count())
	assert.True(t, m.Capacity() >= 13)

	// Fill
	m.Clear()
	n = m.Fill(0xAA, 4)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, m.Pos())
	assert.Equal(t, 4, m.Count())
	assert.Equal(t, []byte{0xAA, 0xAA, 0xAA, 0xAA}, m.Bytes())
}

// TestAutoGrow tests automatic buffer growth when writing beyond capacity.
func TestAutoGrow(t *testing.T) {
	// Start with small buffer
	m := New(4)
	assert.Equal(t, 4, m.Capacity())
	assert.Equal(t, 0, m.Count())

	// Write beyond initial capacity
	m.PutU32(0x01020304)
	assert.Equal(t, 4, m.Count())
	assert.Equal(t, 4, m.Capacity())

	// Write more - should trigger auto-grow
	m.PutU32(0x05060708)
	assert.Equal(t, 8, m.Count())
	assert.True(t, m.Capacity() >= 8)

	// Write large array - should grow again
	largeData := make([]byte, 100)
	for i := range largeData {
		largeData[i] = byte(i)
	}
	m.PutArr8(largeData)
	assert.Equal(t, 108, m.Count())
	assert.True(t, m.Capacity() >= 108)

	// Verify data integrity
	m.Rewind()
	assert.Equal(t, uint32(0x01020304), m.TakeU32())
	assert.Equal(t, uint32(0x05060708), m.TakeU32())
	outData := make([]byte, 100)
	m.TakeArr8(outData)
	assert.Equal(t, largeData, outData)
}

// TestAutoGrowFromEmpty tests auto-grow from completely empty buffer.
func TestAutoGrowFromEmpty(t *testing.T) {
	// Create buffer with zero capacity
	m := New(0)
	assert.Equal(t, 0, m.Capacity())

	// Should auto-grow with minimum capacity
	m.PutU64(0x0102030405060708)
	assert.Equal(t, 8, m.Count())
	assert.True(t, m.Capacity() >= 8)

	m.Rewind()
	assert.Equal(t, uint64(0x0102030405060708), m.TakeU64())
}

// TestPanics tests panic conditions for buffer operations.
func TestPanics(t *testing.T) {
	m := New(10)

	// Test that Put operations no longer panic - they auto-grow instead
	m.pos = 9
	m.PutU16(0x1234) // Should NOT panic, will auto-grow
	assert.Equal(t, 11, m.pos)
	assert.True(t, m.Capacity() >= 11)

	// Test readable panic (still applies - can't read what doesn't exist)
	m.Clear()
	m.PutU8(1)
	m.Rewind()
	assert.Panics(t, func() {
		m.TakeU16()
	}, "TakeU16 should panic when not enough readable data")

	// Test peekable panic
	m.pos = 0
	assert.Panics(t, func() {
		m.PeekU8(2)
	}, "PeekU8(2) should panic when offset out of bounds")

	assert.Panics(t, func() {
		m.PeekU8(-1)
	}, "PeekU8(-1) should panic when offset is negative")

	// Test overwritable panic
	m.pos = 0
	assert.Panics(t, func() {
		m.OverwriteU8(2, 0xFF)
	}, "OverwriteU8(2) should panic when offset out of bounds")
}

// TestGrow tests the Grow method for pre-allocating space.
func TestGrow(t *testing.T) {
	// Test growing from empty buffer
	m := New(10)
	assert.Equal(t, 10, m.Capacity())
	assert.Equal(t, 0, m.Count())

	// Grow to ensure 50 more bytes can be written
	m.Grow(50)
	assert.True(t, m.Capacity() >= 50, "Capacity should be at least 50 after Grow(50)")

	// Write some data
	m.PutU32(0x12345678)
	assert.Equal(t, 4, m.Count())
	oldCap := m.Capacity()

	// Grow when there's already data
	m.Grow(100)
	newCap := m.Capacity()
	assert.True(t, newCap >= 104, "Capacity should be at least count(4) + grow(100) = 104")
	assert.True(t, newCap >= oldCap, "Capacity should not decrease")

	// Verify data integrity after grow
	m.Rewind()
	assert.Equal(t, uint32(0x12345678), m.TakeU32())

	// Test that subsequent writes don't trigger reallocation
	startCap := m.Capacity()
	for i := 0; i < 90; i++ {
		m.PutU8(byte(i))
	}
	assert.Equal(t, startCap, m.Capacity(), "Capacity should not change after pre-allocation")

	// Test Grow with 0
	m.Clear()
	oldCap = m.Capacity()
	m.Grow(0)
	assert.Equal(t, oldCap, m.Capacity(), "Grow(0) should not change capacity")

	// Test Grow panic on negative value
	assert.Panics(t, func() {
		m.Grow(-1)
	}, "Grow should panic on negative value")
}

// TestReserve tests the Reserve method for pre-allocating total capacity.
func TestReserve(t *testing.T) {
	// Test reserving from empty buffer
	m := New(10)
	assert.Equal(t, 10, m.Capacity())

	// Reserve total capacity of 100
	m.Reserve(100)
	assert.True(t, m.Capacity() >= 100, "Capacity should be at least 100 after Reserve(100)")

	// Write some data
	m.PutU64(0x0102030405060708)
	assert.Equal(t, 8, m.Count())

	// Reserve with smaller value (should be no-op)
	oldCap := m.Capacity()
	m.Reserve(50)
	assert.Equal(t, oldCap, m.Capacity(), "Reserve with smaller value should not change capacity")

	// Reserve with larger value
	m.Reserve(200)
	assert.True(t, m.Capacity() >= 200, "Capacity should be at least 200 after Reserve(200)")

	// Verify data integrity
	m.Rewind()
	assert.Equal(t, uint64(0x0102030405060708), m.TakeU64())

	// Test Reserve with current capacity (no-op)
	m.Clear()
	currentCap := m.Capacity()
	m.Reserve(currentCap)
	assert.Equal(t, currentCap, m.Capacity(), "Reserve with current capacity should be no-op")

	// Test Reserve panic on negative value
	assert.Panics(t, func() {
		m.Reserve(-1)
	}, "Reserve should panic on negative value")
}

// TestGrowVsReserve tests the difference between Grow and Reserve.
func TestGrowVsReserve(t *testing.T) {
	// Test Grow - relative to current length
	m1 := New(10)
	m1.PutU32(0x12345678) // count = 4
	m1.Grow(20)           // ensure can write 20 more bytes
	assert.True(t, m1.Capacity() >= 24, "After Grow(20) with count=4, capacity should be >= 24")

	// Test Reserve - absolute capacity
	m2 := New(10)
	m2.PutU32(0x12345678) // count = 4
	m2.Reserve(20)        // ensure total capacity is 20
	assert.True(t, m2.Capacity() >= 20, "After Reserve(20), capacity should be >= 20")
	assert.True(t, m2.Capacity() < m1.Capacity(), "Reserve(20) should result in smaller capacity than Grow(20)")

	// Practical use case: pre-allocate for known writes
	m3 := New(0)
	m3.Reserve(1024) // Pre-allocate 1KB
	startCap := m3.Capacity()

	// Write multiple times without reallocation
	for i := 0; i < 256; i++ {
		m3.PutU32(uint32(i))
	}
	assert.Equal(t, 1024, m3.Count())
	assert.Equal(t, startCap, m3.Capacity(), "Capacity should not change with pre-allocation")
}
