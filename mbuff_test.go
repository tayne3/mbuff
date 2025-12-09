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

// TestNewBufferCreation tests buffer creation with specified capacity.
func TestNewBufferCreation(t *testing.T) {
	// Test with initial capacity
	b := NewBuffer(100)

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
	m2 := NewBuffer(0)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	// Should panic when writing to zero-capacity buffer
	assert.Panics(t, func() {
		m2.PutU8(0xFF)
	}, "PutU8 should panic on zero-capacity buffer")

	// Test New alias
	mAlias := New(50)
	assert.IsType(t, &Buffer{}, mAlias)
	assert.Equal(t, 50, mAlias.Capacity())
	assert.Equal(t, 0, mAlias.Count())
}

// TestNewBufferFrom tests wrapping existing buffer with data for parsing.
func TestNewBufferFrom(t *testing.T) {
	// Test with buffer containing data
	data := []byte{0x12, 0x34, 0x56, 0x78, 0xAB, 0xCD, 0xEF, 0x01}
	b := NewBufferFrom(data)

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
	m2 := NewBufferFrom(emptyData)
	assert.Equal(t, 0, m2.Capacity())
	assert.Equal(t, 0, m2.Count())
	assert.True(t, m2.IsEmpty())

	// Test with buffer that has capacity > length
	buf := make([]byte, 5, 10)
	buf[0], buf[1], buf[2], buf[3], buf[4] = 0x01, 0x02, 0x03, 0x04, 0x05
	m3 := NewBufferFrom(buf)
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
	m4 := NewBufferFrom(original)
	m4.OverwriteU8(0, 0xFF)
	assert.Equal(t, byte(0xFF), original[0]) // original should be modified
}

// TestStatusAndCapacity tests buffer status queries and capacity management.
func TestStatusAndCapacity(t *testing.T) {
	b := NewBuffer(20)

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

// TestDataAndBytes tests data access methods.
func TestDataAndBytes(t *testing.T) {
	b := NewBuffer(10)
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
	// WritableBytes() returns the same as ReadableBytes in the current implementation
	assert.Equal(t, b.ReadableBytes(), b.WritableBytes())
}

// TestPointerAndConfig tests position control methods.
func TestPointerAndConfig(t *testing.T) {
	b := NewBuffer(10)
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

// TestSkip tests the Skip method of the Buffer.
func TestSkip(t *testing.T) {
	b := NewBuffer(10)
	_, _ = b.Write([]byte{1, 2, 3, 4, 5})
	b.Rewind()

	// Skip positive
	n := b.Skip(2)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, b.Pos())

	// Skip more than readable
	n = b.Skip(10)
	assert.Equal(t, 3, n) // Only 3 bytes were readable
	assert.Equal(t, 5, b.Pos())

	// Skip from end
	n = b.Skip(1)
	assert.Equal(t, 0, n)
	assert.Equal(t, 5, b.Pos())

	// Skip negative
	b.Rewind()
	n = b.Skip(-5)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, b.Pos())
}

// TestCompact tests buffer compaction.
func TestCompact(t *testing.T) {
	b := NewBuffer(10)

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

// TestRead tests the Read method of the Buffer.
func TestRead(t *testing.T) {
	b := NewBuffer(10)
	_, _ = b.Write([]byte{1, 2, 3, 4, 5})
	b.Rewind()

	// Simple read
	readData := make([]byte, 3)
	n, err := b.Read(readData)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, b.Pos())
	assert.Equal(t, []byte{1, 2, 3}, readData)

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

	// Read into empty slice
	n, err = b.Read([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestWrite tests the Write method of the Buffer.
func TestWrite(t *testing.T) {
	b := NewBuffer(10)
	writeData := []byte{1, 2, 3, 4, 5}

	// Simple write
	n, err := b.Write(writeData)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, b.Pos())
	assert.Equal(t, 5, b.Count())
	assert.Equal(t, writeData, b.Bytes())

	// Write empty slice
	n, err = b.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, 5, b.Pos())
	assert.Equal(t, 5, b.Count())

	n, err = b.Write(writeData)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())

	// Write with overflow
	b.Seek(8)
	n, err = b.Write(writeData) // writeData is 5 bytes, only 2 writable
	assert.Equal(t, io.ErrShortWrite, err)
	assert.Equal(t, 2, n)        // Should write 2 bytes (partial write)
	assert.Equal(t, 10, b.Pos()) // Position should advance to 10
}

// TestFill tests the Fill method of the Buffer.
func TestFill(t *testing.T) {
	b := NewBuffer(10)

	// Simple fill
	n := b.Fill(0xAA, 10)
	assert.Equal(t, 10, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())
	for i := 0; i < 10; i++ {
		assert.Equal(t, byte(0xAA), b.Bytes()[i])
	}

	// Fill with zero length
	n = b.Fill(0xBB, 0)
	assert.Equal(t, 0, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())

	// Fill with negative length
	n = b.Fill(0xBB, -1)
	assert.Equal(t, 0, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())

	// Fill with overflow
	b.Seek(8)
	n = b.Fill(0xCC, 5) // 5 bytes requested, only 2 writable
	assert.Equal(t, 2, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())
	assert.Equal(t, []byte{0xCC, 0xCC}, b.Bytes()[8:10])

	// Fill with no writable space
	n = b.Fill(0xDD, 1)
	assert.Equal(t, 0, n)
	assert.Equal(t, 10, b.Pos())
	assert.Equal(t, 10, b.Count())
}

// TestPanics tests panic conditions for buffer operations.
func TestPanics(t *testing.T) {
	b := NewBuffer(10)
	b.Fill(0xAA, 10)

	// Test that Put operations now panic on overflow
	b.Seek(9)
	assert.Panics(t, func() {
		b.PutU16(0x1234) // Should panic, needs 2 bytes, only 1 writable
	}, "PutU16 should panic on buffer overflow")

	// Reset for next test
	b.Clear()

	// Test readable panic (still applies - can't read what doesn't exist)
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

// TestCommitAndSince tests Commit and Since operations.
func TestCommitAndSince(t *testing.T) {
	t.Run("Commit", func(t *testing.T) {
		b := NewBuffer(20)
		n := b.Commit(10)
		assert.Equal(t, 10, n)
		assert.Equal(t, 10, b.Pos())
		assert.Equal(t, 10, b.Count())

		n = b.Commit(20) // try to commit more than writable
		assert.Equal(t, 10, n)
		assert.Equal(t, 20, b.Pos())
		assert.Equal(t, 20, b.Count())

		b.Clear()
		n = b.Commit(-1) // negative commit
		assert.Equal(t, 0, n)
		assert.Equal(t, 0, b.Pos())
		assert.Equal(t, 0, b.Count())
	})

	t.Run("Since", func(t *testing.T) {
		b := NewBuffer(20)
		b.PutU32(0x01020304)
		b.PutU32(0x05060708)
		b.PutU32(0x090A0B0C)
		b.PutU32(0x0D0E0F10)
		assert.Equal(t, 16, b.Count())

		b.Seek(4)
		s := b.Since(0, 8)
		assert.Equal(t, 8, s.Count())
		assert.Equal(t, 0, s.Pos())
		assert.Equal(t, uint32(0x01020304), s.TakeU32())
		assert.Equal(t, uint32(0x05060708), s.TakeU32())

		// test since with negative values for default behavior
		b.Seek(4)
		s2 := b.Since(-1, -1) // should be from current pos to end
		assert.Equal(t, 12, s2.Count())
		assert.Equal(t, 0, s2.Pos())
		assert.Equal(t, uint32(0x05060708), s2.TakeU32())

		assert.Panics(t, func() {
			b.Since(8, 4)
		}, "Since with s > e should panic")
	})

	t.Run("ReadableSince", func(t *testing.T) {
		b := NewBuffer(16)
		b.PutU32(0x01020304)
		b.PutU32(0x05060708)
		b.Seek(2)
		s := b.ReadableSince()
		assert.Equal(t, 6, s.Count())
		assert.Equal(t, []byte{0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, s.Bytes())
	})

	t.Run("WritableSince", func(t *testing.T) {
		b := NewBuffer(16)
		b.PutU32(0x01020304)
		b.PutU32(0x05060708)
		b.Seek(2)
		s := b.WritableSince()
		assert.Equal(t, 6, s.Count())
		assert.Equal(t, []byte{0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, s.Bytes())
	})
}
