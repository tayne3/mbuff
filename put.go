// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

// PutU8 writes a uint8 at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutU8(v uint8) {
	required := b.pos + 1
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutU8: buffer overflow")
		}
		b.data = b.data[:required]
	}

	b.data[b.pos] = v
	b.pos += 1
}

// PutU16 writes a uint16 at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutU16(v uint16) {
	required := b.pos + 2
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutU16: buffer overflow")
		}
		b.data = b.data[:required]
	}

	b.order.PutUint16(b.data[b.pos:], v)
	b.pos += 2
}

// PutU32 writes a uint32 at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutU32(v uint32) {
	required := b.pos + 4
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutU32: buffer overflow")
		}
		b.data = b.data[:required]
	}

	b.order.PutUint32(b.data[b.pos:], b.HLSwap32(v))
	b.pos += 4
}

// PutU64 writes a uint64 at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutU64(v uint64) {
	required := b.pos + 8
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutU64: buffer overflow")
		}
		b.data = b.data[:required]
	}

	b.order.PutUint64(b.data[b.pos:], b.HLSwap64(v))
	b.pos += 8
}

// PutArr8 writes a byte slice at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutArr8(v []byte) {
	required := b.pos + len(v)
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutArr8: buffer overflow")
		}
		b.data = b.data[:required]
	}

	n := copy(b.data[b.pos:], v)
	b.pos += n
}

// PutArr16 writes a uint16 slice at the current position and advances the position.
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutArr16(v []uint16) {
	byteLen := len(v) << 1
	required := b.pos + byteLen
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutArr16: buffer overflow")
		}
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
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutArr32(v []uint32) {
	byteLen := len(v) << 2
	required := b.pos + byteLen
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutArr32: buffer overflow")
		}
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
// Panics if the write would exceed the buffer's capacity.
func (b *Buffer) PutArr64(v []uint64) {
	byteLen := len(v) << 3
	required := b.pos + byteLen
	if required > len(b.data) {
		if required > cap(b.data) {
			panic("mbuff.Buffer.PutArr64: buffer overflow")
		}
		b.data = b.data[:required]
	}

	writePos := b.pos
	for _, val := range v {
		b.order.PutUint64(b.data[writePos:], b.HLSwap64(val))
		writePos += 8
	}
	b.pos += byteLen
}
