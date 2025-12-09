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

// Buffer is a buffer for efficient binary data processing.
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
type Buffer struct {
	data   []byte           // underlying byte array
	pos    int              // current position
	order  binary.ByteOrder // byte order
	hlswap bool             // whether high-low swap is enabled
}

// New creates a new Buffer with the specified initial capacity.
// The buffer will automatically grow as needed when writing data.
func New(capacity int) *Buffer {
	return NewBuffer(capacity)
}

// NewBuffer creates a new Buffer with the specified initial capacity.
// The buffer will automatically grow as needed when writing data.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		data:   make([]byte, 0, capacity),
		pos:    0,
		order:  binary.BigEndian,
		hlswap: false,
	}
}

// NewBufferFrom creates a new Buffer with the given buffer.
// The buffer's length and capacity are preserved, and position is set to 0.
// This is useful when you have a []byte with existing data that needs to be parsed.
func NewBufferFrom(buffer []byte) *Buffer {
	return &Buffer{
		data:   buffer,
		pos:    0,
		order:  binary.BigEndian,
		hlswap: false,
	}
}

// Capacity returns the total capacity of the buffer.
func (b *Buffer) Capacity() int { return cap(b.data) }

// Count returns the length of valid data in the buffer.
func (b *Buffer) Count() int { return len(b.data) }

// Pos returns the current position.
func (b *Buffer) Pos() int { return b.pos }

// Readable returns the length of readable data (count - pos).
func (b *Buffer) Readable() int { return len(b.data) - b.pos }

// Writable returns the length of writable space from current position (capacity - pos).
func (b *Buffer) Writable() int { return cap(b.data) - b.pos }

// Appendable returns the length of appendable space (capacity - len).
func (b *Buffer) Appendable() int { return cap(b.data) - len(b.data) }

// IsEmpty checks if the buffer is empty (len == 0).
func (b *Buffer) IsEmpty() bool { return len(b.data) == 0 }

// IsFull checks if the buffer is full (len == capacity).
func (b *Buffer) IsFull() bool { return len(b.data) == cap(b.data) }

// Data returns the underlying byte slice with its current capacity.
func (b *Buffer) Data() []byte { return b.data[:cap(b.data)] }

// Bytes returns the slice of valid data (from 0 to len).
func (b *Buffer) Bytes() []byte { return b.data }

// ReadableBytes returns the slice of readable data (from pos to len).
func (b *Buffer) ReadableBytes() []byte { return b.data[b.pos:] }

// WritableBytes returns the slice of writable data (from pos to capacity).
func (b *Buffer) WritableBytes() []byte { return b.data[b.pos:] }

// Rewind resets the position to 0.
func (b *Buffer) Rewind() { b.pos = 0 }

// Clear clears the buffer by resetting both position and length to 0.
func (b *Buffer) Clear() {
	b.pos = 0
	b.data = b.data[:0]
}

// SetEndian sets the byte order for reading/writing multi-byte values.
func (b *Buffer) SetEndian(e Endian) {
	if e == LittleEndian {
		b.order = binary.LittleEndian
	} else {
		b.order = binary.BigEndian
	}
}

// GetEndian returns the current byte order.
func (b *Buffer) GetEndian() Endian {
	if b.order == binary.LittleEndian {
		return LittleEndian
	}
	return BigEndian
}

// SetHLSwap enables or disables high-low byte swap for 32/64-bit types.
func (b *Buffer) SetHLSwap(enable bool) { b.hlswap = enable }

// Seek moves the position to the specified offset from the start.
// The offset must be within [0, len].
func (b *Buffer) Seek(offset int) error {
	if offset < 0 || offset > len(b.data) {
		return fmt.Errorf("mbuff.Buffer.Seek: seek offset %d out of bounds [0, %d]", offset, len(b.data))
	}
	b.pos = offset
	return nil
}

// Reseek moves the position to the specified offset from the end (len).
// The offset must be within [0, len].
func (b *Buffer) Reseek(offset int) error {
	if offset < 0 || offset > len(b.data) {
		return fmt.Errorf("mbuff.Buffer.Reseek: reseek offset %d out of bounds [0, %d]", offset, len(b.data))
	}
	b.pos = len(b.data) - offset
	return nil
}

// Skip clamps advancement to Readable() to preserve the invariant pos <= len(data).
// Negative lengths are coerced to 0 so reads remain monotonic-forward even with
// untrusted or computed sizes. Returns the amount actually advanced.
func (b *Buffer) Skip(length int) int {
	if length < 0 {
		return 0
	}
	readable := b.Readable()
	if length > readable {
		length = readable
	}
	b.pos += length
	return length
}

// Commit clamps advancement to Writable() to prevent stepping beyond capacity.
// Unlike Skip, it extends len(data) when pos crosses current length so newly
// produced bytes become visible to readers. This separates capacity reservation
// from data visibility and avoids exposing uninitialized memory.
func (b *Buffer) Commit(length int) int {
	if length < 0 {
		return 0
	}
	writable := b.Writable()
	if length > writable {
		length = writable
	}
	b.pos += length
	if b.pos > len(b.data) {
		b.data = b.data[:b.pos]
	}
	return length
}

// Since creates a zero-copy view over [s, e) of the same backing array to
// avoid allocation and copying. s < 0 means "from current pos", e < 0 means
// "to len(data)". It preserves endianness and swap settings. Because the view
// shares storage, mutations via either buffer are reflected in both; invalid
// ranges fail fast with a panic.
func (b *Buffer) Since(s, e int) *Buffer {
	if s < 0 {
		s = b.pos
	}
	if e < 0 {
		e = len(b.data)
	}
	if s > e {
		panic("invalid range")
	}
	return &Buffer{
		data:   b.data[s:e:cap(b.data)],
		pos:    0,
		order:  b.order,
		hlswap: b.hlswap,
	}
}

// ReadableSince returns a zero-copy view over the current readable region to
// let downstream parsers operate on valid bytes without copying.
func (b *Buffer) ReadableSince() *Buffer { return b.Since(b.pos, len(b.data)) }

// WritableSince intentionally mirrors ReadableSince, exposing only the
// initialized region [pos:len]. Capacity management remains explicit via Grow
// or Reserve to avoid accidental writes past initialized data.
func (b *Buffer) WritableSince() *Buffer { return b.Since(b.pos, len(b.data)) }

// Compact compacts the buffer by moving readable data [pos:len] to [0:len-pos],
// then resets pos to 0 and adjusts data length.
func (b *Buffer) Compact() {
	if b.pos == 0 {
		return
	}
	if b.pos == len(b.data) {
		b.Clear()
		return
	}

	readableLen := b.Readable()
	if readableLen > 0 {
		copy(b.data[0:readableLen], b.data[b.pos:])
	}
	b.data = b.data[:readableLen]
	b.pos = 0
}

// Peek reads data from the current position into p without advancing the position.
// Returns the actual number of bytes read.
func (b *Buffer) Peek(p []byte) (n int) {
	if len(p) == 0 {
		return 0
	}
	readable := b.Readable()
	n = len(p)
	if n > readable {
		n = readable
	}
	if n == 0 {
		return 0
	}
	copy(p, b.data[b.pos:b.pos+n])
	return
}

// Fill fills the buffer with byte bt for the specified length.
// It fills up to the available writable space.
func (b *Buffer) Fill(bt byte, length int) int {
	if length <= 0 {
		return 0
	}

	writable := b.Writable()
	if writable == 0 {
		return 0
	}
	if length > writable {
		length = writable
	}
	if b.pos+length > len(b.data) {
		b.data = b.data[:b.pos+length]
	}

	end := b.pos + length
	for i := b.pos; i < end; i++ {
		b.data[i] = bt
	}

	b.pos += length
	return length
}

// Read reads data from the buffer into p.
// It implements the io.Reader interface.
// Reads at most len(p) or b.Readable() bytes.
func (b *Buffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	readable := b.Readable()
	if readable == 0 {
		err = io.EOF
		return
	}

	n = len(p)
	if n > readable {
		n = readable
	}
	copy(p, b.data[b.pos:])
	b.pos += n
	return
}

// Write writes data from p into the buffer.
// It implements the io.Writer interface.
// It writes up to the available writable space.
func (b *Buffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	writable := b.Writable()
	n = len(p)
	if n > writable {
		n = writable
		err = io.ErrShortWrite
	}
	copy(b.data[b.pos:b.pos+n], p[:n])
	b.pos += n
	return
}
