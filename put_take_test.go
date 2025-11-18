// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPut tests that Put operations correctly extend the slice length
// when capacity is sufficient.
func TestPut(t *testing.T) {
	// Create a buffer with length 2 and capacity 10
	buf := make([]byte, 2, 10)
	buf[0] = 0x01
	buf[1] = 0x02
	b := NewBufferFrom(buf)
	assert.Equal(t, 2, b.Count())
	assert.Equal(t, 10, b.Capacity())

	// Test PutU32 - should extend length from 2 to 6
	b.Seek(2)
	b.PutU32(0x03040506)
	assert.Equal(t, 6, b.Pos())
	assert.Equal(t, 6, b.Count())
	assert.Equal(t, 10, b.Capacity()) // Capacity should not change
	expected32 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	assert.Equal(t, expected32, b.Bytes())

	// Test PutU16 - should extend length from 6 to 8
	b.Seek(6)
	b.PutU16(0x0708)
	assert.Equal(t, 8, b.Pos())
	assert.Equal(t, 8, b.Count())
	expected16 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	assert.Equal(t, expected16, b.Bytes())

	b.Seek(2)
	assert.NotPanics(t, func() {
		b.PutU64(0x0708090A0B0C0D0E)
	}, "PutU64 should not panic on buffer overflow")
	b.Seek(3)
	assert.Panics(t, func() {
		b.PutU64(0x0708090A0B0C0D0E)
	}, "PutU64 should panic on buffer overflow")

	b.Seek(6)
	assert.NotPanics(t, func() {
		b.PutU32(0x0708090A)
	}, "PutU32 should not panic on buffer overflow")
	b.Seek(7)
	assert.Panics(t, func() {
		b.PutU32(0x0708090A)
	}, "PutU32 should panic on buffer overflow")

	b.Seek(8)
	assert.NotPanics(t, func() {
		b.PutU16(0x0708)
	}, "PutU16 should not panic on buffer overflow")
	b.Seek(9)
	assert.Panics(t, func() {
		b.PutU16(0x0708)
	}, "PutU16 should panic on buffer overflow")

	b.Seek(9)
	assert.NotPanics(t, func() {
		b.PutU8(0x07)
	}, "PutU8 should not panic on buffer overflow")
	b.Seek(10)
	assert.Panics(t, func() {
		b.PutU8(0x07)
	}, "PutU8 should panic on buffer overflow")
}

// TestTake tests that Take operations correctly read values
// from the buffer and update the position.
func TestTake(t *testing.T) {
	// Pre-allocate enough space for all puts
	b := NewBuffer(30)

	// Put various scalar types
	b.PutU8(0x01)
	b.PutU8(0xFE)
	b.PutU16(0x0304) // Big Endian
	b.PutU16(0xFFF6)
	b.PutU32(0x05060708)
	b.PutU32(0xFFFFFFEC)
	b.PutU64(0x090A0B0C0D0E0F10)
	b.PutU64(0xFFFFFFFFFFFFFFE2)

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
	assert.Equal(t, expected, b.Bytes())

	// Take scalar types
	b.Rewind()
	assert.Equal(t, uint8(0x01), b.TakeU8())
	assert.Equal(t, uint8(0xFE), b.TakeU8())
	assert.Equal(t, uint16(0x0304), b.TakeU16())
	assert.Equal(t, uint16(0xFFF6), b.TakeU16())
	assert.Equal(t, uint32(0x05060708), b.TakeU32())
	assert.Equal(t, uint32(0xFFFFFFEC), b.TakeU32())
	assert.Equal(t, uint64(0x090A0B0C0D0E0F10), b.TakeU64())
	assert.Equal(t, uint64(0xFFFFFFFFFFFFFFE2), b.TakeU64())
}

// TestPutTakeArr tests array read/write operations.
func TestPutTakeArr(t *testing.T) {
	b := NewBuffer(32)

	// Arr8
	arr8 := []byte{1, 2, 3, 4}
	b.PutArr8(arr8)
	assert.Equal(t, 4, b.Pos())
	assert.Equal(t, 4, b.Count())

	out8 := make([]byte, 4)
	b.Rewind()
	b.TakeArr8(out8)
	assert.Equal(t, arr8, out8)

	// Arr16
	b.Clear()
	arr16 := []uint16{0x1122, 0x3344}
	b.PutArr16(arr16)
	expected16 := []byte{0x11, 0x22, 0x33, 0x44}
	assert.Equal(t, expected16, b.Bytes())
	out16 := make([]uint16, 2)
	b.Rewind()
	b.TakeArr16(out16)
	assert.Equal(t, arr16, out16)

	// Arr32 with endianness and HLSwap
	b.Clear()
	b.SetEndian(LittleEndian)
	b.SetHLSwap(true)
	arr32 := []uint32{0x11223344, 0xAABBCCDD}
	b.PutArr32(arr32)

	// With LittleEndian and HLSwap:
	// 0x11223344 -> HLSwap -> 0x22114433 -> LE -> [0x33, 0x44, 0x11, 0x22]
	// 0xAABBCCDD -> HLSwap -> 0xBBAADDCC -> LE -> [0xCC, 0xDD, 0xAA, 0xBB]
	expected32 := []byte{0x33, 0x44, 0x11, 0x22, 0xCC, 0xDD, 0xAA, 0xBB}
	assert.Equal(t, expected32, b.Bytes())

	out32 := make([]uint32, 2)
	b.Rewind()
	b.TakeArr32(out32)
	assert.Equal(t, arr32[0], out32[0])
	assert.Equal(t, arr32[1], out32[1])
	b.SetEndian(BigEndian)
	b.SetHLSwap(false)

	// Arr64
	b.Clear()
	arr64 := []uint64{0x1122334455667788, 0xAABBCCDDEEFF0011}
	b.PutArr64(arr64)
	expected64 := []byte{
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
		0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11,
	}
	assert.Equal(t, expected64, b.Bytes())
	out64 := make([]uint64, 2)
	b.Rewind()
	b.TakeArr64(out64)
	assert.Equal(t, arr64, out64)

	// Test panic on overflow
	b.Clear()
	b.PutArr8(make([]byte, 32)) // Fill the buffer
	assert.Equal(t, 0, b.Writable())
	assert.Panics(t, func() {
		b.PutArr8([]byte{0x00})
	}, "PutArr8 should panic on buffer overflow")

	b.Clear()
	b.PutArr16(make([]uint16, 16)) // Fill the buffer
	assert.Panics(t, func() {
		b.PutArr16([]uint16{0x00})
	}, "PutArr16 should panic on buffer overflow")

	b.Clear()
	b.PutArr32(make([]uint32, 8)) // Fill the buffer
	assert.Panics(t, func() {
		b.PutArr32([]uint32{0x00})
	}, "PutArr32 should panic on buffer overflow")

	b.Clear()
	b.PutArr64(make([]uint64, 4)) // Fill the buffer
	assert.Panics(t, func() {
		b.PutArr64([]uint64{0x00})
	}, "PutArr64 should panic on buffer overflow")
}
