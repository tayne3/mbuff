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

// TestNewMBuff tests buffer creation and initialization.
func TestNewMBuff(t *testing.T) {
	b := make([]byte, 100)
	m := NewMBuff(b)

	assert.Equal(t, 100, m.Capacity())
	assert.Equal(t, 100, len(m.Data()))
	assert.Equal(t, 0, m.Pos())
	assert.Equal(t, 0, m.Count())
	assert.True(t, m.IsEmpty())
	// Verify default byte order is BigEndian
	assert.Equal(t, binary.BigEndian, m.order)

	// Test slice with different length and capacity
	b2 := make([]byte, 10, 50)
	m2 := NewMBuff(b2)
	assert.Equal(t, 50, m2.Capacity())
	assert.Equal(t, 50, len(m2.Data()))
}

// TestStatusAndCapacity tests buffer status queries and capacity management.
func TestStatusAndCapacity(t *testing.T) {
	m := NewMBuff(make([]byte, 20))

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
	m.count = 20
	assert.Equal(t, 0, m.Appendable())
	assert.True(t, m.IsFull())
}

// TestDataAndBytes tests data access methods.
func TestDataAndBytes(t *testing.T) {
	m := NewMBuff(make([]byte, 10))
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
	m := NewMBuff(make([]byte, 10))
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
	m := NewMBuff(make([]byte, 10))
	m.data[0] = 0xAA
	m.data[1] = 0xBB
	m.data[2] = 0xCC
	m.data[3] = 0xDD
	m.data[4] = 0xEE
	m.count = 5

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
	m := NewMBuff(make([]byte, 10))
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

	// Write short
	m.pos = 8
	m.count = 8
	n, err = m.Write(writeData) // can only write 2 bytes
	assert.Equal(t, io.ErrShortWrite, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 10, m.Pos())
	assert.Equal(t, 10, m.Count())

	// Fill
	m.Clear()
	n = m.Fill(0xAA, 4)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, m.Pos())
	assert.Equal(t, 4, m.Count())
	assert.Equal(t, []byte{0xAA, 0xAA, 0xAA, 0xAA}, m.Bytes())
}

// TestPanics tests panic conditions for buffer operations.
func TestPanics(t *testing.T) {
	m := NewMBuff(make([]byte, 10))

	// Test writable panic
	m.pos = 9
	assert.Panics(t, func() {
		m.PutU16(0x1234)
	}, "PutU16 should panic when not enough writable space")

	// Test readable panic
	m.Clear()
	m.PutU8(1)
	m.Rewind()
	assert.Panics(t, func() {
		m.TakeU16()
	}, "TakeU16 should panic when not enough readable space")

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
