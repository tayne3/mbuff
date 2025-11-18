// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHavePeekable checks if the offset and length are within the count.
func (b *Buffer) mustHavePeekable(offset int, n int) int {
	absPos := b.pos + offset
	if absPos < 0 || absPos+n > len(b.data) {
		panic(fmt.Errorf("mbuff.Buffer.mustHavePeekable: peek at pos %d + offset %d exceeds count %d", b.pos, offset, len(b.data)))
	}
	return absPos
}

// PeekU8 reads a uint8 at pos+offset without advancing the position.
func (b *Buffer) PeekU8(offset int) uint8 {
	absPos := b.mustHavePeekable(offset, 1)
	return b.data[absPos]
}

// PeekU16 reads a uint16 at pos+offset without advancing the position.
func (b *Buffer) PeekU16(offset int) uint16 {
	absPos := b.mustHavePeekable(offset, 2)
	return b.order.Uint16(b.data[absPos : absPos+2])
}

// PeekU32 reads a uint32 at pos+offset without advancing the position.
func (b *Buffer) PeekU32(offset int) uint32 {
	absPos := b.mustHavePeekable(offset, 4)
	v := b.order.Uint32(b.data[absPos : absPos+4])
	return b.HLSwap32(v)
}

// PeekU64 reads a uint64 at pos+offset without advancing the position.
func (b *Buffer) PeekU64(offset int) uint64 {
	absPos := b.mustHavePeekable(offset, 8)
	v := b.order.Uint64(b.data[absPos : absPos+8])
	return b.HLSwap64(v)
}

// PeekArr8 reads bytes at pos+offset into slice v without advancing the position.
func (b *Buffer) PeekArr8(offset int, v []byte) {
	byteLen := len(v)
	absPos := b.mustHavePeekable(offset, byteLen)
	copy(v, b.data[absPos:absPos+byteLen])
}

// PeekArr16 reads uint16 values at pos+offset into slice v without advancing the position.
func (b *Buffer) PeekArr16(offset int, v []uint16) {
	byteLen := len(v) << 1
	absPos := b.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		v[i] = b.order.Uint16(b.data[readPos : readPos+2])
		readPos += 2
	}
}

// PeekArr32 reads uint32 values at pos+offset into slice v without advancing the position.
func (b *Buffer) PeekArr32(offset int, v []uint32) {
	byteLen := len(v) << 2
	absPos := b.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		val := b.order.Uint32(b.data[readPos : readPos+4])
		v[i] = b.HLSwap32(val)
		readPos += 4
	}
}

// PeekArr64 reads uint64 values at pos+offset into slice v without advancing the position.
func (b *Buffer) PeekArr64(offset int, v []uint64) {
	byteLen := len(v) << 3
	absPos := b.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		val := b.order.Uint64(b.data[readPos : readPos+8])
		v[i] = b.HLSwap64(val)
		readPos += 8
	}
}
