// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

// Endian represents byte order for multi-byte values.
type Endian bool

const (
	// BigEndian represents big-endian byte order.
	BigEndian Endian = false
	// LittleEndian represents little-endian byte order.
	LittleEndian Endian = true
)
