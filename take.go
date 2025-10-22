// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"fmt"
)

// mustHaveReadable checks if the current position and length are within the count.
func (m *MBuff) mustHaveReadable(n int) {
	if m.pos+n > m.count {
		panic(fmt.Errorf("mbuff: read of %d bytes at pos %d exceeds count %d", n, m.pos, m.count))
	}
}

// TakeU8 reads and returns a uint8 at the current position, then advances the position.
func (m *MBuff) TakeU8() uint8 {
	m.mustHaveReadable(1)
	v := m.data[m.pos]
	m.pos += 1
	return v
}

// TakeU16 reads and returns a uint16 at the current position, then advances the position.
func (m *MBuff) TakeU16() uint16 {
	m.mustHaveReadable(2)
	v := m.order.Uint16(m.data[m.pos : m.pos+2])
	m.pos += 2
	return v
}

// TakeU32 reads and returns a uint32 at the current position, then advances the position.
func (m *MBuff) TakeU32() uint32 {
	m.mustHaveReadable(4)
	v := m.order.Uint32(m.data[m.pos : m.pos+4])
	m.pos += 4
	return m.HLSwap32(v)
}

// TakeU64 reads and returns a uint64 at the current position, then advances the position.
func (m *MBuff) TakeU64() uint64 {
	m.mustHaveReadable(8)
	v := m.order.Uint64(m.data[m.pos : m.pos+8])
	m.pos += 8
	return m.HLSwap64(v)
}

// TakeArr8 reads bytes at the current position into slice v, then advances the position.
func (m *MBuff) TakeArr8(v []byte) {
	m.mustHaveReadable(len(v))
	n := copy(v, m.data[m.pos:])
	m.pos += n
}

// TakeArr16 reads uint16 values at the current position into slice v, then advances the position.
func (m *MBuff) TakeArr16(v []uint16) {
	byteLen := len(v) << 1
	m.mustHaveReadable(byteLen)
	readPos := m.pos
	for i := range v {
		v[i] = m.order.Uint16(m.data[readPos : readPos+2])
		readPos += 2
	}
	m.pos += byteLen
}

// TakeArr32 reads uint32 values at the current position into slice v, then advances the position.
func (m *MBuff) TakeArr32(v []uint32) {
	byteLen := len(v) << 2
	m.mustHaveReadable(byteLen)
	readPos := m.pos
	for i := range v {
		val := m.order.Uint32(m.data[readPos : readPos+4])
		v[i] = m.HLSwap32(val)
		readPos += 4
	}
	m.pos += byteLen
}

// TakeArr64 reads uint64 values at the current position into slice v, then advances the position.
func (m *MBuff) TakeArr64(v []uint64) {
	byteLen := len(v) << 3
	m.mustHaveReadable(byteLen)
	readPos := m.pos
	for i := range v {
		val := m.order.Uint64(m.data[readPos : readPos+8])
		v[i] = m.HLSwap64(val)
		readPos += 8
	}
	m.pos += byteLen
}
