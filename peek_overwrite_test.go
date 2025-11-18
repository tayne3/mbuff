// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPeek tests peeking scalar values without advancing position.
func TestPeek(t *testing.T) {
	b := NewBuffer(30)
	b.PutU8(0x01)
	b.PutU16(0x0203)
	b.PutU32(0x04050607)
	b.PutU64(0x08090A0B0C0D0E0F)
	b.Seek(1)

	// Peek from pos+offset
	assert.Equal(t, uint16(0x0203), b.PeekU16(0))
	assert.Equal(t, uint32(0x04050607), b.PeekU32(2))
	assert.Equal(t, uint64(0x08090A0B0C0D0E0F), b.PeekU64(6))

	// Verify position unchanged
	assert.Equal(t, 1, b.Pos())

	b.Clear()
	_, _ = b.Write([]byte{1, 2, 3, 4, 5})
	assert.Equal(t, 5, b.Count())
	b.Rewind()
	b.Seek(3)

	// Simple peek
	peekData := make([]byte, 2)
	n := b.Peek(peekData)
	assert.Equal(t, 2, n)
	assert.Equal(t, 3, b.Pos()) // Position should not change
	assert.Equal(t, []byte{4, 5}, peekData)

	// Peek more than readable
	peekData = make([]byte, 5)
	n = b.Peek(peekData)
	assert.Equal(t, 2, n) // Should only peek the remaining 2 bytes
	assert.Equal(t, []byte{4, 5}, peekData[:2])

	// Peek into empty slice
	n = b.Peek([]byte{})
	assert.Equal(t, 0, n)

	// Peek from empty buffer
	b.Clear()
	n = b.Peek(peekData)
	assert.Equal(t, 0, n)

}

// TestOverwrite tests overwriting values at specific offsets.
func TestOverwrite(t *testing.T) {
	b := NewBuffer(30)
	b.PutU32(0xAAAAAAAA)
	b.PutU32(0xBBBBBBBB)
	b.PutU16(0xCCCC)
	b.PutU64(0xDDDDDDDDDDDDDDDD)

	// Overwrite at absolute offsets
	b.OverwriteU8(0, 0xDE)
	b.OverwriteU16(2, 0xDEAD)
	b.OverwriteU32(4, 0xDEADBEEF)
	b.OverwriteU64(10, 0xCAFEBABEDEADBEAD)

	// Verify position unchanged
	assert.Equal(t, 18, b.Pos())

	expected := []byte{
		0xDE, 0xAA, 0xDE, 0xAD, // OverwriteU8(0, 0xDE), OverwriteU16(2, 0xDEAD)
		0xDE, 0xAD, 0xBE, 0xEF, // OverwriteU32(4, 0xDEADBEEF)
		0xCC, 0xCC, // from PutU16(0xCCCC)
		0xCA, 0xFE, 0xBA, 0xBE, 0xDE, 0xAD, 0xBE, 0xAD, // OverwriteU64(10, 0xCAFEBABEDEADBEAD)
	}
	assert.Equal(t, expected, b.Bytes())

	b.Rewind()
	assert.Equal(t, uint32(0xDEAADEAD), b.TakeU32())
	assert.Equal(t, uint32(0xDEADBEEF), b.TakeU32())
}

// TestPeekArr tests array peek operations.
func TestPeekArr(t *testing.T) {
	b := NewBuffer(64)
	b.PutU8(0xFF)
	b.PutArr8([]byte{0xAA, 0xBB, 0xCC})
	b.PutArr16([]uint16{0x1122, 0x3344})
	b.PutArr32([]uint32{0x11223344, 0x55667788})
	b.PutArr64([]uint64{0xAABBCCDDEEFF0011, 0x2233445566778899})
	b.Seek(1)

	// PeekArr8
	out8 := make([]byte, 3)
	b.PeekArr8(0, out8)
	assert.Equal(t, []byte{0xAA, 0xBB, 0xCC}, out8)
	assert.Equal(t, 1, b.Pos())

	// PeekArr16
	out16 := make([]uint16, 2)
	b.PeekArr16(3, out16)
	assert.Equal(t, []uint16{0x1122, 0x3344}, out16)
	assert.Equal(t, 1, b.Pos())

	// PeekArr32
	out32 := make([]uint32, 2)
	b.PeekArr32(7, out32)
	assert.Equal(t, []uint32{0x11223344, 0x55667788}, out32)
	assert.Equal(t, 1, b.Pos())

	// PeekArr64
	out64 := make([]uint64, 2)
	b.PeekArr64(15, out64)
	assert.Equal(t, []uint64{0xAABBCCDDEEFF0011, 0x2233445566778899}, out64)
	assert.Equal(t, 1, b.Pos())
}

// TestOverwriteArr tests array overwrite operations.
func TestOverwriteArr(t *testing.T) {
	b := NewBuffer(64)
	b.PutArr32([]uint32{0xAAAAAAAA, 0xBBBBBBBB, 0xCCCCCCCC, 0xDDDDDDDD, 0xEEEEEEEE, 0xFFFFFFFF})
	posAfterPut := b.Pos()

	// OverwriteArr8
	in8 := []byte{0xDE, 0xAD}
	b.OverwriteArr8(2, in8)

	// OverwriteArr16
	in16 := []uint16{0xDEAD, 0xBEEF}
	b.OverwriteArr16(4, in16)

	// OverwriteArr32
	in32 := []uint32{0xDEADBEEF, 0xCAFEBABE}
	b.OverwriteArr32(8, in32)

	// OverwriteArr64
	in64 := []uint64{0xDEADBEEFCAFEBABE}
	b.OverwriteArr64(16, in64)

	// Check that pos is unchanged
	assert.Equal(t, posAfterPut, b.Pos())

	b.Rewind()
	assert.Equal(t, uint32(0xAAAADEAD), b.TakeU32())
	assert.Equal(t, uint32(0xDEADBEEF), b.TakeU32())
	assert.Equal(t, uint32(0xDEADBEEF), b.TakeU32())
	assert.Equal(t, uint32(0xCAFEBABE), b.TakeU32())
	assert.Equal(t, uint64(0xDEADBEEFCAFEBABE), b.TakeU64())
}
