// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MBuff is a buffer for efficient binary data processing.
// Buffer structure:
//
//	     <----------------------------------------------[Capacity]------------------------------------------>
//	                                      <-----------------------------[Writable]-------------------------->
//	     |--------------------------------+--------------------------+--------------------------------------|
//	     |***********[Processed]**********|========[Readable]========|.............[Appendable].............|
//	     |--------------------------------+--------------------------+--------------------------------------|
//	     ^                                ^                          ^                                      ^
//	     0                               pos                       count                                 capacity
//	  (data[0])                      (position)                 (len(data))                             (cap(data))
//
//	Region descriptions:
//	  - Processed:
//	    Description: From buffer start to current position.
//	    Range: 0~{pos-1}
//
//	  - Readable:
//	    Description: From current position to end of valid data.
//	    Range: {pos}~{count-1}
//
//	  - Writable:
//	    Description: From current position to buffer end.
//	    Range: {pos}~{capacity-1}
//
//	  - Appendable:
//	    Description: From end of valid data to buffer end.
//	    Range: {count}~{capacity-1}
//
//	Byte order and high-low swap:
//	  - order:  Byte order for handling different endianness.
//	  - hlswap: Flag to enable/disable high-low byte swap for 32-bit and 64-bit types.
type MBuff struct {
	data   []byte           // underlying byte array
	pos    int              // current position
	order  binary.ByteOrder // byte order
	hlswap bool             // whether high-low swap is enabled
}

// New creates a new MBuff with the specified initial capacity.
// The buffer will automatically grow as needed when writing data.
func New(capacity int) *MBuff {
	return &MBuff{
		data:   make([]byte, 0, capacity),
		pos:    0,
		order:  binary.BigEndian,
		hlswap: false,
	}
}

// NewMBuff creates a new MBuff with the given buffer.
// The buffer's length and capacity are preserved, and position is set to 0.
// This is useful when you have a []byte with existing data that needs to be parsed.
func NewMBuff(buffer []byte) *MBuff {
	return &MBuff{
		data:   buffer,
		pos:    0,
		order:  binary.BigEndian,
		hlswap: false,
	}
}

// Capacity returns the total capacity of the buffer.
func (m *MBuff) Capacity() int { return cap(m.data) }

// Count returns the length of valid data in the buffer.
func (m *MBuff) Count() int { return len(m.data) }

// Pos returns the current position.
func (m *MBuff) Pos() int { return m.pos }

// Readable returns the length of readable data (count - pos).
func (m *MBuff) Readable() int { return len(m.data) - m.pos }

// Writable returns the length of writable space from current position (capacity - pos).
func (m *MBuff) Writable() int { return cap(m.data) - m.pos }

// Appendable returns the length of appendable space (capacity - len).
func (m *MBuff) Appendable() int { return cap(m.data) - len(m.data) }

// IsEmpty checks if the buffer is empty (len == 0).
func (m *MBuff) IsEmpty() bool { return len(m.data) == 0 }

// IsFull checks if the buffer is full (len == capacity).
func (m *MBuff) IsFull() bool { return len(m.data) == cap(m.data) }

// Data returns the underlying byte slice with its current capacity.
func (m *MBuff) Data() []byte { return m.data[:cap(m.data)] }

// Bytes returns the slice of valid data (from 0 to len).
func (m *MBuff) Bytes() []byte { return m.data }

// ReadableBytes returns the slice of readable data (from pos to len).
func (m *MBuff) ReadableBytes() []byte { return m.data[m.pos:] }

// Rewind resets the position to 0.
func (m *MBuff) Rewind() { m.pos = 0 }

// Clear clears the buffer by resetting both position and length to 0.
func (m *MBuff) Clear() {
	m.pos = 0
	m.data = m.data[:0]
}

// SetEndian sets the byte order for reading/writing multi-byte values.
func (m *MBuff) SetEndian(e Endian) {
	if e == LittleEndian {
		m.order = binary.LittleEndian
	} else {
		m.order = binary.BigEndian
	}
}

// GetEndian returns the current byte order.
func (m *MBuff) GetEndian() Endian {
	if m.order == binary.LittleEndian {
		return LittleEndian
	}
	return BigEndian
}

// SetHLSwap enables or disables high-low byte swap for 32/64-bit types.
func (m *MBuff) SetHLSwap(enable bool) { m.hlswap = enable }

// Seek moves the position to the specified offset from the start.
// The offset must be within [0, len].
func (m *MBuff) Seek(offset int) error {
	if offset < 0 || offset > len(m.data) {
		return fmt.Errorf("mbuff: seek offset %d out of bounds [0, %d]", offset, len(m.data))
	}
	m.pos = offset
	return nil
}

// Reseek moves the position to the specified offset from the end (len).
// The offset must be within [0, len].
func (m *MBuff) Reseek(offset int) error {
	if offset < 0 || offset > len(m.data) {
		return fmt.Errorf("mbuff: reseek offset %d out of bounds [0, %d]", offset, len(m.data))
	}
	m.pos = len(m.data) - offset
	return nil
}

// Skip advances the position forward by the specified length.
// Returns the actual number of bytes skipped.
func (m *MBuff) Skip(length int) int {
	if length < 0 {
		return 0
	}
	readable := m.Readable()
	if length > readable {
		length = readable
	}
	m.pos += length
	return length
}

// Compact compacts the buffer by moving readable data [pos:len] to [0:len-pos],
// then resets pos to 0 and adjusts data length.
func (m *MBuff) Compact() {
	if m.pos == 0 {
		return
	}
	if m.pos == len(m.data) {
		m.Clear()
		return
	}

	readableLen := m.Readable()
	if readableLen > 0 {
		copy(m.data[0:readableLen], m.data[m.pos:])
	}
	m.data = m.data[:readableLen]
	m.pos = 0
}

// Peek reads data from the current position into p without advancing the position.
// Returns the actual number of bytes read.
func (m *MBuff) Peek(p []byte) (n int) {
	if len(p) == 0 {
		return 0
	}
	readable := m.Readable()
	n = len(p)
	if n > readable {
		n = readable
	}
	if n == 0 {
		return 0
	}
	copy(p, m.data[m.pos:m.pos+n])
	return
}

// Fill fills the buffer with byte b for the specified length.
// The buffer will automatically grow if necessary.
func (m *MBuff) Fill(b byte, length int) int {
	if length <= 0 {
		return 0
	}

	required := m.pos + length
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	end := m.pos + length
	for i := m.pos; i < end; i++ {
		m.data[i] = b
	}

	m.pos += length
	return length
}

// Grow grows the buffer's capacity to guarantee space for n more bytes.
// After Grow(n), at least n bytes can be written to the buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (m *MBuff) Grow(n int) {
	if n < 0 {
		panic("mbuff.MBuff.Grow: negative count")
	}
	required := len(m.data) + n
	m.ensure(required)
}

// Reserve reserves space to guarantee the buffer can hold at least capacity bytes
// without another allocation.
// If capacity is less than or equal to current capacity, this is a no-op.
// If capacity is negative, Reserve will panic.
func (m *MBuff) Reserve(capacity int) {
	if capacity < 0 {
		panic("mbuff.MBuff.Reserve: negative capacity")
	}
	m.ensure(capacity)
}

// ensure ensures the underlying data slice has at least the required capacity.
// If not, it grows the slice using a strategy similar to append (doubling capacity).
func (m *MBuff) ensure(required int) {
	if required <= cap(m.data) {
		return
	}

	// Calculate new capacity: double the current capacity or use required, whichever is larger
	newCap := cap(m.data) * 2
	if newCap < required {
		newCap = required
	}

	// Handle the case when current capacity is 0
	if newCap == 0 {
		newCap = required
		if newCap < 64 {
			newCap = 64 // Minimum initial capacity
		}
	}

	// Allocate new slice and copy existing data
	newData := make([]byte, len(m.data), newCap)
	copy(newData, m.data)
	m.data = newData
}

// Read reads data from the buffer into p.
// It implements the io.Reader interface.
// Reads at most len(p) or m.Readable() bytes.
func (m *MBuff) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	readable := m.Readable()
	if readable == 0 {
		return 0, io.EOF
	}

	n = len(p)
	if n > readable {
		n = readable
	}
	copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// Write writes data from p into the buffer.
// It implements the io.Writer interface.
// The buffer will automatically grow if necessary to accommodate all data.
func (m *MBuff) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	required := m.pos + len(p)
	m.ensure(required)

	// Extend data slice if needed
	if required > len(m.data) {
		m.data = m.data[:required]
	}

	n = copy(m.data[m.pos:], p)
	m.pos += n
	return n, nil
}
