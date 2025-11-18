// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHaveReadable checks if the current position and length are within the count.
func (b *Buffer) mustHaveReadable(n int) {
	if b.pos+n > len(b.data) {
		panic(fmt.Errorf("mbuff.Buffer.mustHaveReadable: read of %d bytes at pos %d exceeds count %d", n, b.pos, len(b.data)))
	}
}

// TakeU8 reads and returns a uint8 at the current position, then advances the position.
func (b *Buffer) TakeU8() uint8 {
	b.mustHaveReadable(1)
	v := b.data[b.pos]
	b.pos += 1
	return v
}

// TakeU16 reads and returns a uint16 at the current position, then advances the position.
func (b *Buffer) TakeU16() uint16 {
	b.mustHaveReadable(2)
	v := b.order.Uint16(b.data[b.pos : b.pos+2])
	b.pos += 2
	return v
}

// TakeU32 reads and returns a uint32 at the current position, then advances the position.
func (b *Buffer) TakeU32() uint32 {
	b.mustHaveReadable(4)
	v := b.order.Uint32(b.data[b.pos : b.pos+4])
	b.pos += 4
	return b.HLSwap32(v)
}

// TakeU64 reads and returns a uint64 at the current position, then advances the position.
func (b *Buffer) TakeU64() uint64 {
	b.mustHaveReadable(8)
	v := b.order.Uint64(b.data[b.pos : b.pos+8])
	b.pos += 8
	return b.HLSwap64(v)
}

// TakeArr8 reads bytes at the current position into slice v, then advances the position.
func (b *Buffer) TakeArr8(v []byte) {
	b.mustHaveReadable(len(v))
	n := copy(v, b.data[b.pos:])
	b.pos += n
}

// TakeArr16 reads uint16 values at the current position into slice v, then advances the position.
func (b *Buffer) TakeArr16(v []uint16) {
	byteLen := len(v) << 1
	b.mustHaveReadable(byteLen)
	readPos := b.pos
	for i := range v {
		v[i] = b.order.Uint16(b.data[readPos : readPos+2])
		readPos += 2
	}
	b.pos += byteLen
}

// TakeArr32 reads uint32 values at the current position into slice v, then advances the position.
func (b *Buffer) TakeArr32(v []uint32) {
	byteLen := len(v) << 2
	b.mustHaveReadable(byteLen)
	readPos := b.pos
	for i := range v {
		val := b.order.Uint32(b.data[readPos : readPos+4])
		v[i] = b.HLSwap32(val)
		readPos += 4
	}
	b.pos += byteLen
}

// TakeArr64 reads uint64 values at the current position into slice v, then advances the position.
func (b *Buffer) TakeArr64(v []uint64) {
	byteLen := len(v) << 3
	b.mustHaveReadable(byteLen)
	readPos := b.pos
	for i := range v {
		val := b.order.Uint64(b.data[readPos : readPos+8])
		v[i] = b.HLSwap64(val)
		readPos += 8
	}
	b.pos += byteLen
}
