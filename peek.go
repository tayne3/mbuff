// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHavePeekable checks if the offset and length are within the count.
func (m *MBuff) mustHavePeekable(offset int, n int) int {
	absPos := m.pos + offset
	if absPos < 0 || absPos+n > len(m.data) {
		panic(fmt.Errorf("mbuff: peek at pos %d + offset %d exceeds count %d", m.pos, offset, len(m.data)))
	}
	return absPos
}

// PeekU8 reads a uint8 at pos+offset without advancing the position.
func (m *MBuff) PeekU8(offset int) uint8 {
	absPos := m.mustHavePeekable(offset, 1)
	return m.data[absPos]
}

// PeekU16 reads a uint16 at pos+offset without advancing the position.
func (m *MBuff) PeekU16(offset int) uint16 {
	absPos := m.mustHavePeekable(offset, 2)
	return m.order.Uint16(m.data[absPos : absPos+2])
}

// PeekU32 reads a uint32 at pos+offset without advancing the position.
func (m *MBuff) PeekU32(offset int) uint32 {
	absPos := m.mustHavePeekable(offset, 4)
	v := m.order.Uint32(m.data[absPos : absPos+4])
	return m.HLSwap32(v)
}

// PeekU64 reads a uint64 at pos+offset without advancing the position.
func (m *MBuff) PeekU64(offset int) uint64 {
	absPos := m.mustHavePeekable(offset, 8)
	v := m.order.Uint64(m.data[absPos : absPos+8])
	return m.HLSwap64(v)
}

// PeekArr8 reads bytes at pos+offset into slice v without advancing the position.
func (m *MBuff) PeekArr8(offset int, v []byte) {
	byteLen := len(v)
	absPos := m.mustHavePeekable(offset, byteLen)
	copy(v, m.data[absPos:absPos+byteLen])
}

// PeekArr16 reads uint16 values at pos+offset into slice v without advancing the position.
func (m *MBuff) PeekArr16(offset int, v []uint16) {
	byteLen := len(v) << 1
	absPos := m.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		v[i] = m.order.Uint16(m.data[readPos : readPos+2])
		readPos += 2
	}
}

// PeekArr32 reads uint32 values at pos+offset into slice v without advancing the position.
func (m *MBuff) PeekArr32(offset int, v []uint32) {
	byteLen := len(v) << 2
	absPos := m.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		val := m.order.Uint32(m.data[readPos : readPos+4])
		v[i] = m.HLSwap32(val)
		readPos += 4
	}
}

// PeekArr64 reads uint64 values at pos+offset into slice v without advancing the position.
func (m *MBuff) PeekArr64(offset int, v []uint64) {
	byteLen := len(v) << 3
	absPos := m.mustHavePeekable(offset, byteLen)
	readPos := absPos
	for i := range v {
		val := m.order.Uint64(m.data[readPos : readPos+8])
		v[i] = m.HLSwap64(val)
		readPos += 8
	}
}
