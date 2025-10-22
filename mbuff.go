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
//	  (data[0])                      (position)               (end of valid data)                       (cap(data))
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
	count  int              // end of valid data
	order  binary.ByteOrder // byte order
	hlswap bool             // whether high-low swap is enabled
}

// NewMBuff creates a new MBuff with the given buffer.
func NewMBuff(buffer []byte) *MBuff {
	return &MBuff{
		data:   buffer[0:cap(buffer)],
		pos:    0,
		count:  0,
		order:  binary.BigEndian,
		hlswap: false,
	}
}

// Capacity returns the total capacity of the buffer.
func (m *MBuff) Capacity() int { return cap(m.data) }

// Count returns the length of valid data in the buffer.
func (m *MBuff) Count() int { return m.count }

// Pos returns the current position.
func (m *MBuff) Pos() int { return m.pos }

// Readable returns the length of readable data (count - pos).
func (m *MBuff) Readable() int { return m.count - m.pos }

// Writable returns the length of writable space from current position (capacity - pos).
// Note: This may overwrite valid data beyond the "Appendable" region.
func (m *MBuff) Writable() int { return cap(m.data) - m.pos }

// Appendable returns the length of appendable space (capacity - count).
func (m *MBuff) Appendable() int { return cap(m.data) - m.count }

// IsEmpty checks if the buffer is empty (count == 0).
func (m *MBuff) IsEmpty() bool { return m.count == 0 }

// IsFull checks if the buffer is full (count == capacity).
func (m *MBuff) IsFull() bool { return m.count == cap(m.data) }

// Data returns the underlying byte slice (from 0 to capacity).
func (m *MBuff) Data() []byte { return m.data }

// Bytes returns the slice of valid data (from 0 to count).
func (m *MBuff) Bytes() []byte { return m.data[0:m.count] }

// ReadableBytes returns the slice of readable data (from pos to count).
func (m *MBuff) ReadableBytes() []byte { return m.data[m.pos:m.count] }

// Rewind resets the position to 0.
func (m *MBuff) Rewind() { m.pos = 0 }

// Clear clears the buffer by resetting both position and count to 0.
func (m *MBuff) Clear() { m.pos = 0; m.count = 0 }

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
// The offset must be within [0, count].
func (m *MBuff) Seek(offset int) error {
	if offset < 0 || offset > m.count {
		return fmt.Errorf("mbuff: seek offset %d out of bounds [0, %d]", offset, m.count)
	}
	m.pos = offset
	return nil
}

// Reseek moves the position to the specified offset from the end (count).
// The offset must be within [0, count].
func (m *MBuff) Reseek(offset int) error {
	if offset < 0 || offset > m.count {
		return fmt.Errorf("mbuff: reseek offset %d out of bounds [0, %d]", offset, m.count)
	}
	m.pos = m.count - offset
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

// Compact compacts the buffer by moving readable data [pos:count] to [0:count-pos],
// then resets pos to 0 and count to (count - pos).
func (m *MBuff) Compact() {
	if m.pos == 0 {
		return
	}
	if m.pos == m.count {
		m.Clear()
		return
	}

	readableLen := m.Readable()
	if readableLen > 0 {
		copy(m.data[0:readableLen], m.data[m.pos:m.count])
	}
	m.count = readableLen
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
// Returns the actual number of bytes filled.
func (m *MBuff) Fill(b byte, length int) int {
	if length <= 0 {
		return 0
	}
	writable := m.Writable()
	if length > writable {
		length = writable
	}
	if length == 0 {
		return 0
	}

	end := m.pos + length
	for i := m.pos; i < end; i++ {
		m.data[i] = b
	}

	m.pos += length
	if m.pos > m.count {
		m.count = m.pos
	}
	return length
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
	copy(p, m.data[m.pos:m.pos+n])
	m.pos += n
	return n, nil
}

// Write writes data from p into the buffer.
// It implements the io.Writer interface.
// Writes at most len(p) or m.Writable() bytes.
// Returns io.ErrShortWrite if fewer bytes are written than len(p).
func (m *MBuff) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	writable := m.Writable()
	n = len(p)
	if n > writable {
		n = writable
	}

	if n > 0 {
		copy(m.data[m.pos:m.pos+n], p[0:n])
		m.pos += n
		if m.pos > m.count {
			m.count = m.pos
		}
	}

	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}
