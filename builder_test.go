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

// TestNewBuilder tests buffer creation with specified capacity.
func TestNewBuilder(t *testing.T) {
	// Test with initial capacity
	b := NewBuilder(100)

	assert.Equal(t, 100, b.Capacity())
	assert.Equal(t, 100, len(b.Data()))
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 0, b.Count())
	assert.True(t, b.IsEmpty())
	// Verify default byte order is BigEndian
	assert.Equal(t, binary.BigEndian, b.order)

	// Test writing to the buffer
	b.PutU32(0x12345678)
	assert.Equal(t, 4, b.Count())
	assert.Equal(t, 4, b.Pos())

	// Test with zero capacity
	m2 := NewBuilder(0)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	// Should auto-grow when writing
	m2.PutU8(0xFF)
	assert.Equal(t, 1, m2.Count())
	assert.True(t, m2.Capacity() >= 1)
}

// TestNewBuilderFrom tests wrapping existing buffer with data for parsing.
func TestNewBuilderFrom(t *testing.T) {
	// Test with buffer containing data
	data := []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x01}
	b := NewBuilderFrom(data)

	// Verify initial state
	assert.Equal(t, 8, b.Capacity())
	assert.Equal(t, 8, b.Count())
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 8, b.Readable())
	assert.False(t, b.IsEmpty())

	// Test reading the existing data
	v1 := b.TakeU16()
	assert.Equal(t, uint16(0x1234), v1)
	assert.Equal(t, 2, b.Pos())
	assert.Equal(t, 6, b.Readable())

	v2 := b.TakeU32()
	assert.Equal(t, uint32(0x5678ABCD), v2)
	assert.Equal(t, 6, b.Pos())
	assert.Equal(t, 2, b.Readable())

	// Test with empty buffer
	emptyData := []byte{}
	m2 := NewBuilderFrom(emptyData)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	assert.True(t, m2.IsEmpty())

	// Test with buffer that has capacity > length
	buf := make([]byte, 5, 10)
	buf[0], buf[1], buf[2], buf[3], buf[4] = 0x01, 0x02, 0x03, 0x04, 0x05
	m3 := NewBuilderFrom(buf)
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
	m4 := NewBuilderFrom(original)
	m4.OverwriteU8(0, 0xFF)
	assert.Equal(t, byte(0xFF), original[0]) // original should be modified
}

// TestBuilder_StatusAndCapacity tests buffer status queries and capacity management.
func TestBuilder_StatusAndCapacity(t *testing.T) {
	b := NewBuilder(20)

	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 0, b.Count())
	assert.Equal(t, 20, b.Capacity())
	assert.Equal(t, 0, b.Readable())
	assert.Equal(t, 20, b.Writable())
	assert.Equal(t, 20, b.Appendable())
	assert.True(t, b.IsEmpty())
	assert.False(t, b.IsFull())

	b.PutU8(1)
	b.PutU8(2)
	b.PutU8(3)
	b.PutU8(4)
	b.PutU8(5)

	assert.Equal(t, 5, b.Pos())
	assert.Equal(t, 5, b.Count())
	assert.Equal(t, 0, b.Readable()) // position at end
	assert.Equal(t, 15, b.Writable())
	assert.Equal(t, 15, b.Appendable())
	assert.False(t, b.IsEmpty())

	b.Rewind()
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 5, b.Readable())
	assert.Equal(t, 20, b.Writable())
	assert.Equal(t, 15, b.Appendable())

	b.Seek(5)
	b.data = b.data[:20]
	assert.Equal(t, 0, b.Appendable())
	assert.True(t, b.IsFull())
}

// TestBuilder_DataAndBytes tests data access methods.
func TestBuilder_DataAndBytes(t *testing.T) {
	b := NewBuilder(10)
	b.PutU8(0xAA)
	b.PutU8(0xBB)
	b.PutU8(0xCC)
	b.Rewind()
	b.PutU8(0xDD) // overwrite
	b.Seek(2)

	// Data() returns the entire underlying slice
	assert.Equal(t, 10, len(b.Data()))
	// Bytes() returns valid data [0:count]
	assert.Equal(t, []byte{0xDD, 0xBB, 0xCC}, b.Bytes())
	// ReadableBytes() returns readable data [pos:count]
	assert.Equal(t, []byte{0xCC}, b.ReadableBytes())
}

// TestBuilder_PointerAndConfig tests position control methods.
func TestBuilder_PointerAndConfig(t *testing.T) {
	b := NewBuilder(10)
	b.PutU8(1)
	b.PutU8(2)
	b.PutU8(3)

	b.Rewind()
	assert.Equal(t, 0, b.Pos())

	b.Seek(1)

	// Seek
	assert.NoError(t, b.Seek(2))
	assert.Equal(t, 2, b.Pos())

	// Seek out of bounds
	assert.Error(t, b.Seek(4))
	assert.Error(t, b.Seek(-1))

	// Reseek
	assert.NoError(t, b.Reseek(1)) // move to count - 1 = 2
	assert.Equal(t, 2, b.Pos())

	assert.NoError(t, b.Reseek(3)) // move to count - 3 = 0
	assert.Equal(t, 0, b.Pos())

	// Reseek out of bounds
	assert.Error(t, b.Reseek(4))

	// Skip
	b.Rewind()
	n := b.Skip(2)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, b.Pos())

	n = b.Skip(5) // can only skip 1 more
	assert.Equal(t, 1, n)
	assert.Equal(t, 3, b.Pos())

	n = b.Skip(1) // nothing left to skip
	assert.Equal(t, 0, n)

	// Clear
	b.Clear()
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 0, b.Count())
}

// TestBuilder_Compact tests buffer compaction.
func TestBuilder_Compact(t *testing.T) {
	b := NewBuilder(10)

	// Use Put operations to set up data
	b.PutU8(0xAA)
	b.PutU8(0xBB)
	b.PutU8(0xCC)
	b.PutU8(0xDD)
	b.PutU8(0xEE)
	assert.Equal(t, 5, len(b.data))

	// Case 1: pos = 0 (no operation needed)
	b.Seek(0)
	b.Compact()
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 5, b.Count())
	assert.Equal(t, byte(0xBB), b.data[1])

	// Case 2: 0 < pos < count (should move data)
	b.Seek(2)
	b.Compact()
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 3, b.Count())
	assert.Equal(t, byte(0xCC), b.data[0])
	assert.Equal(t, byte(0xDD), b.data[1])
	assert.Equal(t, byte(0xEE), b.data[2])

	// Case 3: pos = count (should clear)
	b.Seek(3)
	b.Compact()
	assert.Equal(t, 0, b.Pos())
	assert.Equal(t, 0, b.Count())
}

// TestBuilder_ReadPeek tests io.Reader interface and Peek method.
func TestBuilder_ReadPeek(t *testing.T) {
	b := NewBuilder(10)
	writeData := []byte{1, 2, 3, 4, 5}
	_, _ = b.Write(writeData)

	// Read
	b.Rewind()
	readData := make([]byte, 3)
	n, err := b.Read(readData)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, b.Pos())
	assert.Equal(t, []byte{1, 2, 3}, readData)

	// Peek
	peekData := make([]byte, 2)
	n = b.Peek(peekData)
	assert.Equal(t, 2, n)
	assert.Equal(t, 3, b.Pos())
	assert.Equal(t, []byte{4, 5}, peekData)

	// Read partial
	n, err = b.Read(readData) // can only read 2 bytes now
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 5, b.Pos())
	assert.Equal(t, []byte{4, 5}, readData[0:2])

	// Read EOF
	n, err = b.Read(readData)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

// TestBuilder_Write tests the Write method.
func TestBuilder_Write(t *testing.T) {
	b := NewBuilder(10)
	writeData := []byte{1, 2, 3, 4, 5}

	// Write to empty buffer
	n, err := b.Write(writeData)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, b.Pos())
	assert.Equal(t, 5, b.Count())
	assert.Equal(t, writeData, b.Bytes())

	n, err = b.Write(writeData)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())

	// Write with auto-grow
	b.Seek(8)
	n, err = b.Write(writeData) // Will auto-grow to accommodate
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 13, b.Pos())
	assert.Equal(t, 13, b.Count())
	assert.Equal(t, 20, b.Capacity())

	// Write empty slice
	n, err = b.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 13, b.Pos()) // Position should not change
}

// TestBuilder_Fill tests the Fill method.
func TestBuilder_Fill(t *testing.T) {
	// Fill empty buffer
	b := NewBuilder(10)
	n := b.Fill(0xAA, 4)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, b.Pos())
	assert.Equal(t, 4, b.Count())
	assert.Equal(t, []byte{0xAA, 0xAA, 0xAA, 0xAA}, b.Bytes())

	// Fill at non-zero position
	b.Fill(0xBB, 2)
	assert.Equal(t, 6, b.Pos())
	assert.Equal(t, 6, b.Count())
	assert.Equal(t, []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB}, b.Bytes())

	// Fill with zero length
	n = b.Fill(0xCC, 0)
	assert.Equal(t, 0, n)
	assert.Equal(t, 6, b.Pos()) // Position should not change

	// Fill with negative length
	n = b.Fill(0xCC, -1)
	assert.Equal(t, 0, n)
	assert.Equal(t, 6, b.Pos()) // Position should not change

	// Fill with auto-grow
	n = b.Fill(0xDD, 10) // Requires 6 + 10 = 16 bytes
	assert.Equal(t, 10, n)
	assert.Equal(t, 16, b.Pos())
	assert.Equal(t, 16, b.Count())
	assert.True(t, b.Capacity() >= 16)
	expected := []byte{
		0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB,
		0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD, 0xDD,
	}
	assert.Equal(t, expected, b.Bytes())
}

// TestBuilder_AutoGrow tests automatic buffer growth when writing beyond capacity.
func TestBuilder_AutoGrow(t *testing.T) {
	// Start with small buffer
	b := NewBuilder(4)
	assert.Equal(t, 4, b.Capacity())
	assert.Equal(t, 0, b.Count())

	// Write beyond initial capacity
	b.PutU32(0x01020304)
	assert.Equal(t, 4, b.Count())
	assert.Equal(t, 4, b.Capacity())

	// Write more - should trigger auto-grow
	b.PutU32(0x05060708)
	assert.Equal(t, 8, b.Count())
	assert.True(t, b.Capacity() >= 8)

	// Write large array - should grow again
	largeData := make([]byte, 100)
	for i := range largeData {
		largeData[i] = byte(i)
	}
	b.PutArr8(largeData)
	assert.Equal(t, 108, b.Count())
	assert.True(t, b.Capacity() >= 108)

	// Verify data integrity
	b.Rewind()
	assert.Equal(t, uint32(0x01020304), b.TakeU32())
	assert.Equal(t, uint32(0x05060708), b.TakeU32())
	outData := make([]byte, 100)
	b.TakeArr8(outData)
	assert.Equal(t, largeData, outData)
}

// TestBuilder_AutoGrowFromEmpty tests auto-grow from completely empty buffer.
func TestBuilder_AutoGrowFromEmpty(t *testing.T) {
	// Create buffer with zero capacity
	b := NewBuilder(0)
	assert.Equal(t, 0, b.Capacity())

	// Should auto-grow with minimum capacity
	b.PutU64(0x0102030405060708)
	assert.Equal(t, 8, b.Count())
	assert.True(t, b.Capacity() >= 8)

	b.Rewind()
	assert.Equal(t, uint64(0x0102030405060708), b.TakeU64())

	// Test auto-grow from empty with required capacity less than 64
	b = NewBuilder(0)
	assert.Equal(t, 0, b.Capacity())
	b.PutU8(0xFF) // Requires 1 byte, should trigger newCap = 64
	assert.Equal(t, 1, b.Count())
	assert.True(t, b.Capacity() >= 64)
	b.Rewind()
	assert.Equal(t, uint8(0xFF), b.TakeU8())
}

// TestBuilder_Panics tests panic conditions for buffer operations.
func TestBuilder_Panics(t *testing.T) {
	b := NewBuilder(10)
	b.Fill(0xAA, 10)

	// Test that Put operations no longer panic - they auto-grow instead
	b.Seek(9)
	b.PutU16(0x1234) // Should NOT panic, will auto-grow
	assert.Equal(t, 11, b.Pos())
	assert.True(t, b.Capacity() >= 11)

	// Test readable panic (still applies - can't read what doesn't exist)
	b.Clear()
	b.PutU8(1)
	b.Rewind()
	assert.Panics(t, func() {
		b.TakeU16()
	}, "TakeU16 should panic when not enough readable data")

	// Test peekable panic
	b.Seek(0)
	assert.Panics(t, func() {
		b.PeekU8(2)
	}, "PeekU8(2) should panic when offset out of bounds")

	assert.Panics(t, func() {
		b.PeekU8(-1)
	}, "PeekU8(-1) should panic when offset is negative")

	// Test overwritable panic
	b.Seek(0)
	assert.Panics(t, func() {
		b.OverwriteU8(2, 0xFF)
	}, "OverwriteU8(2) should panic when offset out of bounds")
}

// TestBuilder_Grow tests the Grow method for pre-allocating space.
func TestBuilder_Grow(t *testing.T) {
	// Test growing from empty buffer
	b := NewBuilder(10)
	assert.Equal(t, 10, b.Capacity())
	assert.Equal(t, 0, b.Count())

	// Grow to ensure 50 more bytes can be written
	b.Grow(50)
	assert.True(t, b.Capacity() >= 50, "Capacity should be at least 50 after Grow(50)")

	// Write some data
	b.PutU32(0x12345678)
	assert.Equal(t, 4, b.Count())
	oldCap := b.Capacity()

	// Grow when there's already data
	b.Grow(100)
	newCap := b.Capacity()
	assert.True(t, newCap >= 104, "Capacity should be at least count(4) + grow(100) = 104")
	assert.True(t, newCap >= oldCap, "Capacity should not decrease")

	// Verify data integrity after grow
	b.Rewind()
	assert.Equal(t, uint32(0x12345678), b.TakeU32())

	// Test that subsequent writes don't trigger reallocation
	startCap := b.Capacity()
	for i := 0; i < 90; i++ {
		b.PutU8(byte(i))
	}
	assert.Equal(t, startCap, b.Capacity(), "Capacity should not change after pre-allocation")

	// Test Grow with 0
	b.Clear()
	oldCap = b.Capacity()
	b.Grow(0)
	assert.Equal(t, oldCap, b.Capacity(), "Grow(0) should not change capacity")

	// Test Grow panic on negative value
	assert.Panics(t, func() {
		b.Grow(-1)
	}, "Grow should panic on negative value")
}

// TestBuilder_Reserve tests the Reserve method for pre-allocating total capacity.
func TestBuilder_Reserve(t *testing.T) {
	// Test reserving from empty buffer
	b := NewBuilder(10)
	assert.Equal(t, 10, b.Capacity())

	// Reserve total capacity of 100
	b.Reserve(100)
	assert.True(t, b.Capacity() >= 100, "Capacity should be at least 100 after Reserve(100)")

	// Write some data
	b.PutU64(0x0102030405060708)
	assert.Equal(t, 8, b.Count())

	// Reserve with smaller value (should be no-op)
	oldCap := b.Capacity()
	b.Reserve(50)
	assert.Equal(t, oldCap, b.Capacity(), "Reserve with smaller value should not change capacity")

	// Reserve with larger value
	b.Reserve(200)
	assert.True(t, b.Capacity() >= 200, "Capacity should be at least 200 after Reserve(200)")

	// Verify data integrity
	b.Rewind()
	assert.Equal(t, uint64(0x0102030405060708), b.TakeU64())

	// Test Reserve with current capacity (no-op)
	b.Clear()
	currentCap := b.Capacity()
	b.Reserve(currentCap)
	assert.Equal(t, currentCap, b.Capacity(), "Reserve with current capacity should be no-op")

	// Test Reserve panic on negative value
	assert.Panics(t, func() {
		b.Reserve(-1)
	}, "Reserve should panic on negative value")
}

// TestBuilder_GrowVsReserve tests the difference between Grow and Reserve.
func TestBuilder_GrowVsReserve(t *testing.T) {
	// Test Grow - relative to current length
	m1 := NewBuilder(10)
	m1.PutU32(0x12345678) // count = 4
	m1.Grow(20)           // ensure can write 20 more bytes
	assert.True(t, m1.Capacity() >= 24, "After Grow(20) with count=4, capacity should be >= 24")

	// Test Reserve - absolute capacity
	m2 := NewBuilder(10)
	m2.PutU32(0x12345678) // count = 4
	m2.Reserve(20)        // ensure total capacity is 20
	assert.True(t, m2.Capacity() >= 20, "After Reserve(20), capacity should be >= 20")
	assert.True(t, m2.Capacity() < m1.Capacity(), "Reserve(20) should result in smaller capacity than Grow(20)")

	// Practical use case: pre-allocate for known writes
	m3 := NewBuilder(0)
	m3.Reserve(1024) // Pre-allocate 1KB
	startCap := m3.Capacity()

	// Write multiple times without reallocation
	for i := 0; i < 256; i++ {
		m3.PutU32(uint32(i))
	}
	assert.Equal(t, 1024, m3.Count())
	assert.Equal(t, startCap, m3.Capacity(), "Capacity should not change with pre-allocation")
}

// TestBuilder_PutArr tests the PutArr methods for writing slices of numeric types.
func TestBuilder_PutArr(t *testing.T) {
	t.Run("PutArr16", func(t *testing.T) {
		b := NewBuilder(0)
		data := []uint16{0x0102, 0x0304, 0x0506}
		b.PutArr16(data)

		assert.Equal(t, 6, b.Pos())
		assert.Equal(t, 6, b.Count())
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, b.Bytes())

		// Test with pre-existing data and auto-grow
		b.Clear()
		b.PutU8(0xFF)
		b.PutArr16(data)
		assert.Equal(t, 7, b.Count())
		assert.Equal(t, []byte{0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, b.Bytes())
	})

	t.Run("PutArr32", func(t *testing.T) {
		b := NewBuilder(0)
		data := []uint32{0x01020304, 0x05060708}
		b.PutArr32(data)

		assert.Equal(t, 8, b.Pos())
		assert.Equal(t, 8, b.Count())
		assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, b.Bytes())

		// Test with pre-existing data and auto-grow
		b.Clear()
		b.PutU16(0xFFEE)
		b.PutArr32(data)
		assert.Equal(t, 10, b.Count())
		assert.Equal(t, []byte{0xFF, 0xEE, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, b.Bytes())
	})

	t.Run("PutArr64", func(t *testing.T) {
		b := NewBuilder(0)
		data := []uint64{0x0102030405060708, 0x090A0B0C0D0E0F10}
		b.PutArr64(data)

		assert.Equal(t, 16, b.Pos())
		assert.Equal(t, 16, b.Count())
		expected := []byte{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
		}
		assert.Equal(t, expected, b.Bytes())

		// Test with pre-existing data and auto-grow
		b.Clear()
		b.PutU32(0xFFEEDDCC)
		b.PutArr64(data)
		assert.Equal(t, 20, b.Count())
		expected = []byte{
			0xFF, 0xEE, 0xDD, 0xCC,
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
		}
		assert.Equal(t, expected, b.Bytes())
	})
}
