// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHaveOverwritable checks if the offset and length are within the count.
func (m *MBuff) mustHaveOverwritable(offset int, n int) {
	if offset < 0 || offset+n > len(m.data) {
		panic(fmt.Errorf("mbuff: overwrite at offset %d exceeds count %d", offset, len(m.data)))
	}
}

// OverwriteU8 overwrites a uint8 at the specified offset.
func (m *MBuff) OverwriteU8(offset int, v uint8) {
	m.mustHaveOverwritable(offset, 1)
	m.data[offset] = v
}

// OverwriteU16 overwrites a uint16 at the specified offset.
func (m *MBuff) OverwriteU16(offset int, v uint16) {
	m.mustHaveOverwritable(offset, 2)
	m.order.PutUint16(m.data[offset:offset+2], v)
}

// OverwriteU32 overwrites a uint32 at the specified offset.
func (m *MBuff) OverwriteU32(offset int, v uint32) {
	m.mustHaveOverwritable(offset, 4)
	v = m.HLSwap32(v)
	m.order.PutUint32(m.data[offset:offset+4], v)
}

// OverwriteU64 overwrites a uint64 at the specified offset.
func (m *MBuff) OverwriteU64(offset int, v uint64) {
	m.mustHaveOverwritable(offset, 8)
	v = m.HLSwap64(v)
	m.order.PutUint64(m.data[offset:offset+8], v)
}

// OverwriteArr8 overwrites bytes at the specified offset with slice v.
func (m *MBuff) OverwriteArr8(offset int, v []byte) {
	byteLen := len(v)
	m.mustHaveOverwritable(offset, byteLen)
	copy(m.data[offset:offset+byteLen], v)
}

// OverwriteArr16 overwrites uint16 values at the specified offset with slice v.
func (m *MBuff) OverwriteArr16(offset int, v []uint16) {
	byteLen := len(v) << 1
	m.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		m.order.PutUint16(m.data[writePos:writePos+2], val)
		writePos += 2
	}
}

// OverwriteArr32 overwrites uint32 values at the specified offset with slice v.
func (m *MBuff) OverwriteArr32(offset int, v []uint32) {
	byteLen := len(v) << 2
	m.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		val = m.HLSwap32(val)
		m.order.PutUint32(m.data[writePos:writePos+4], val)
		writePos += 4
	}
}

// OverwriteArr64 overwrites uint64 values at the specified offset with slice v.
func (m *MBuff) OverwriteArr64(offset int, v []uint64) {
	byteLen := len(v) << 3
	m.mustHaveOverwritable(offset, byteLen)
	writePos := offset
	for _, val := range v {
		val = m.HLSwap64(val)
		m.order.PutUint64(m.data[writePos:writePos+8], val)
		writePos += 8
	}
}
