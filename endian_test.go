// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEndian tests byte order handling.
func TestEndian(t *testing.T) {
	b := NewBuffer(8)

	// Big Endian (default)
	assert.Equal(t, BigEndian, b.GetEndian())
	b.PutU32(0x12345678)
	expectedBE := []byte{0x12, 0x34, 0x56, 0x78}
	assert.Equal(t, expectedBE, b.Bytes())

	// Little Endian
	b.Clear()
	b.SetEndian(LittleEndian)
	assert.Equal(t, LittleEndian, b.GetEndian())
	b.PutU32(0x12345678)
	expectedLE := []byte{0x78, 0x56, 0x34, 0x12}
	assert.Equal(t, expectedLE, b.Bytes())

	b.Rewind()
	v := b.TakeU32()
	assert.Equal(t, uint32(0x12345678), v)
}
