// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"encoding/binary"
)

type Builder struct {
	Buffer
}

func NewBuilder(capacity int) *Builder {
	return &Builder{
		Buffer{
			data:   make([]byte, 0, capacity),
			pos:    0,
			order:  binary.BigEndian,
			hlswap: false,
		}}
}

func NewBuilderFrom(buffer []byte) *Builder {
	return &Builder{
		Buffer{
			data:   buffer,
			pos:    0,
			order:  binary.BigEndian,
			hlswap: false,
		},
	}
}

// ensure ensures the underlying data slice has at least the required capacity.
// If not, it grows the slice.
func (b *Builder) ensure(required int) {
	if required <= cap(b.data) {
		return
	}

	var newCap int
	if cap(b.data) == 0 {
		// If current capacity is 0, handle initial allocation separately.
		newCap = required
		if newCap < 64 {
			newCap = 64 // Minimum initial capacity
		}
	} else {
		// Calculate new capacity: double the current capacity or use required, whichever is larger
		newCap = cap(b.data) * 2
		if newCap < required {
			newCap = required
		}
	}

	// Allocate new slice and copy existing data
	newData := make([]byte, len(b.data), newCap)
	copy(newData, b.data)
	b.data = newData
}

// Grow grows the buffer's capacity to guarantee space for n more bytes.
// After Grow(n), at least n bytes can be written to the buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Builder) Grow(n int) {
	if n < 0 {
		panic("mbuff.Builder.Grow: negative count")
	}
	required := len(b.data) + n
	b.ensure(required)
}

// Reserve reserves space to guarantee the buffer can hold at least capacity bytes
// without another allocation.
// If capacity is less than or equal to current capacity, this is a no-op.
// If capacity is negative, Reserve will panic.
func (b *Builder) Reserve(capacity int) {
	if capacity < 0 {
		panic("mbuff.Builder.Reserve: negative capacity")
	}
	b.ensure(capacity)
}

// Fill fills the buffer with byte b for the specified length.
// The buffer will automatically grow if necessary.
func (b *Builder) Fill(bt byte, length int) int {
	if length <= 0 {
		return 0
	}

	required := b.pos + length
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	end := b.pos + length
	for i := b.pos; i < end; i++ {
		b.data[i] = bt
	}

	b.pos += length
	return length
}

// Write writes data from p into the buffer.
// It implements the io.Writer interface.
// The buffer will automatically grow if necessary to accommodate all data.
func (b *Builder) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	required := b.pos + len(p)
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	n = copy(b.data[b.pos:], p)
	b.pos += n
	return n, nil
}

// PutU8 writes a uint8 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutU8(v uint8) {
	required := b.pos + 1
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	b.data[b.pos] = v
	b.pos += 1
}

// PutU16 writes a uint16 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutU16(v uint16) {
	required := b.pos + 2
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	b.order.PutUint16(b.data[b.pos:], v)
	b.pos += 2
}

// PutU32 writes a uint32 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutU32(v uint32) {
	required := b.pos + 4
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	b.order.PutUint32(b.data[b.pos:], b.HLSwap32(v))
	b.pos += 4
}

// PutU64 writes a uint64 at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutU64(v uint64) {
	required := b.pos + 8
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	b.order.PutUint64(b.data[b.pos:], b.HLSwap64(v))
	b.pos += 8
}

// PutArr8 writes a byte slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutArr8(v []byte) {
	required := b.pos + len(v)
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	n := copy(b.data[b.pos:], v)
	b.pos += n
}

// PutArr16 writes a uint16 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutArr16(v []uint16) {
	byteLen := len(v) << 1
	required := b.pos + byteLen
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	writePos := b.pos
	for _, val := range v {
		b.order.PutUint16(b.data[writePos:], val)
		writePos += 2
	}
	b.pos += byteLen
}

// PutArr32 writes a uint32 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutArr32(v []uint32) {
	byteLen := len(v) << 2
	required := b.pos + byteLen
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	writePos := b.pos
	for _, val := range v {
		b.order.PutUint32(b.data[writePos:], b.HLSwap32(val))
		writePos += 4
	}
	b.pos += byteLen
}

// PutArr64 writes a uint64 slice at the current position and advances the position.
// The buffer will automatically grow if necessary.
func (b *Builder) PutArr64(v []uint64) {
	byteLen := len(v) << 3
	required := b.pos + byteLen
	b.ensure(required)

	// Extend data slice if needed
	if required > len(b.data) {
		b.data = b.data[:required]
	}

	writePos := b.pos
	for _, val := range v {
		b.order.PutUint64(b.data[writePos:], b.HLSwap64(val))
		writePos += 8
	}
	b.pos += byteLen
}
