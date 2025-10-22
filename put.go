// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHaveWritable checks if the current position and length are within the capacity.
func (m *MBuff) mustHaveWritable(n int) {
	if m.pos+n > cap(m.data) {
		panic(fmt.Errorf("mbuff: write of %d bytes at pos %d exceeds capacity %d", n, m.pos, cap(m.data)))
	}
}

// PutU8 writes a uint8 at the current position and advances the position.
func (m *MBuff) PutU8(v uint8) {
	m.mustHaveWritable(1)
	m.data[m.pos] = v
	m.pos += 1
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutU16 writes a uint16 at the current position and advances the position.
func (m *MBuff) PutU16(v uint16) {
	m.mustHaveWritable(2)
	m.order.PutUint16(m.data[m.pos:m.pos+2], v)
	m.pos += 2
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutU32 writes a uint32 at the current position and advances the position.
func (m *MBuff) PutU32(v uint32) {
	m.mustHaveWritable(4)
	v = m.HLSwap32(v)
	m.order.PutUint32(m.data[m.pos:m.pos+4], v)
	m.pos += 4
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutU64 writes a uint64 at the current position and advances the position.
func (m *MBuff) PutU64(v uint64) {
	m.mustHaveWritable(8)
	v = m.HLSwap64(v)
	m.order.PutUint64(m.data[m.pos:m.pos+8], v)
	m.pos += 8
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutArr8 writes a byte slice at the current position and advances the position.
func (m *MBuff) PutArr8(v []byte) {
	m.mustHaveWritable(len(v))
	n := copy(m.data[m.pos:], v)
	m.pos += n
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutArr16 writes a uint16 slice at the current position and advances the position.
func (m *MBuff) PutArr16(v []uint16) {
	byteLen := len(v) << 1
	m.mustHaveWritable(byteLen)
	writePos := m.pos
	for _, val := range v {
		m.order.PutUint16(m.data[writePos:writePos+2], val)
		writePos += 2
	}
	m.pos += byteLen
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutArr32 writes a uint32 slice at the current position and advances the position.
func (m *MBuff) PutArr32(v []uint32) {
	byteLen := len(v) << 2
	m.mustHaveWritable(byteLen)
	writePos := m.pos
	for _, val := range v {
		val = m.HLSwap32(val)
		m.order.PutUint32(m.data[writePos:writePos+4], val)
		writePos += 4
	}
	m.pos += byteLen
	if m.pos > m.count {
		m.count = m.pos
	}
}

// PutArr64 writes a uint64 slice at the current position and advances the position.
func (m *MBuff) PutArr64(v []uint64) {
	byteLen := len(v) << 3
	m.mustHaveWritable(byteLen)
	writePos := m.pos
	for _, val := range v {
		val = m.HLSwap64(val)
		m.order.PutUint64(m.data[writePos:writePos+8], val)
		writePos += 8
	}
	m.pos += byteLen
	if m.pos > m.count {
		m.count = m.pos
	}
}
