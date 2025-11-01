// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

// PutU8 writes a uint8 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutU8(v uint8) {
	required := m.pos + 1
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	m.data[m.pos] = v
	m.pos += 1
}

// PutU16 writes a uint16 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutU16(v uint16) {
	required := m.pos + 2
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	m.order.PutUint16(m.data[m.pos:], v)
	m.pos += 2
}

// PutU32 writes a uint32 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutU32(v uint32) {
	required := m.pos + 4
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	v = m.HLSwap32(v)
	m.order.PutUint32(m.data[m.pos:], v)
	m.pos += 4
}

// PutU64 writes a uint64 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutU64(v uint64) {
	required := m.pos + 8
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	v = m.HLSwap64(v)
	m.order.PutUint64(m.data[m.pos:], v)
	m.pos += 8
}

// PutArr8 writes a byte slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutArr8(v []byte) {
	required := m.pos + len(v)
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	n := copy(m.data[m.pos:], v)
	m.pos += n
}

// PutArr16 writes a uint16 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutArr16(v []uint16) {
	byteLen := len(v) << 1
	required := m.pos + byteLen
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	writePos := m.pos
	for _, val := range v {
		m.order.PutUint16(m.data[writePos:], val)
		writePos += 2
	}
	m.pos += byteLen
}

// PutArr32 writes a uint32 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutArr32(v []uint32) {
	byteLen := len(v) << 2
	required := m.pos + byteLen
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	writePos := m.pos
	for _, val := range v {
		val = m.HLSwap32(val)
		m.order.PutUint32(m.data[writePos:], val)
		writePos += 4
	}
	m.pos += byteLen
}

// PutArr64 writes a uint64 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (m *MBuff) PutArr64(v []uint64) {
	byteLen := len(v) << 3
	required := m.pos + byteLen
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	writePos := m.pos
	for _, val := range v {
		val = m.HLSwap64(val)
		m.order.PutUint64(m.data[writePos:], val)
		writePos += 8
	}
	m.pos += byteLen
}
