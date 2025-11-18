// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHaveOverwritable checks if the offset and length are within the count.
func (b *Buffer) mustHaveOverwritable(offset int, n int) {
	if offset < 0 || offset+n > len(b.data) {
		panic(fmt.Errorf("mbuff.Buffer.mustHaveOverwritable: overwrite at offset %d exceeds count %d", offset, len(b.data)))
	}
}

// OverwriteU8 overwrites a uint8 at the specified offset.
func (b *Buffer) OverwriteU8(offset int, v uint8) {
	b.mustHaveOverwritable(offset, 1)
	b.data[offset] = v
}

// OverwriteU16 overwrites a uint16 at the specified offset.
func (b *Buffer) OverwriteU16(offset int, v uint16) {
	b.mustHaveOverwritable(offset, 2)
	b.order.PutUint16(b.data[offset:offset+2], v)
}

// OverwriteU32 overwrites a uint32 at the specified offset.
func (b *Buffer) OverwriteU32(offset int, v uint32) {
	b.mustHaveOverwritable(offset, 4)
	b.order.PutUint32(b.data[offset:offset+4], b.HLSwap32(v))
}

// OverwriteU64 overwrites a uint64 at the specified offset.
func (b *Buffer) OverwriteU64(offset int, v uint64) {
	b.mustHaveOverwritable(offset, 8)
	b.order.PutUint64(b.data[offset:offset+8], b.HLSwap64(v))
}

// OverwriteArr8 overwrites bytes at the specified offset with slice v.
func (b *Buffer) OverwriteArr8(offset int, v []byte) {
	byteLen := len(v)
	b.mustHaveOverwritable(offset, byteLen)
	copy(b.data[offset:offset+byteLen], v)
}

// OverwriteArr16 overwrites uint16 values at the specified offset with slice v.
func (b *Buffer) OverwriteArr16(offset int, v []uint16) {
	byteLen := len(v) << 1
	b.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		b.order.PutUint16(b.data[writePos:writePos+2], val)
		writePos += 2
	}
}

// OverwriteArr32 overwrites uint32 values at the specified offset with slice v.
func (b *Buffer) OverwriteArr32(offset int, v []uint32) {
	byteLen := len(v) << 2
	b.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		b.order.PutUint32(b.data[writePos:writePos+4], b.HLSwap32(val))
		writePos += 4
	}
}

// OverwriteArr64 overwrites uint64 values at the specified offset with slice v.
func (b *Buffer) OverwriteArr64(offset int, v []uint64) {
	byteLen := len(v) << 3
	b.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		b.order.PutUint64(b.data[writePos:writePos+8], b.HLSwap64(val))
		writePos += 8
	}
}
