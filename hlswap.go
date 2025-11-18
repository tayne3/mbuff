// Copyright 2025 The Gromb Authors. All rights reserved.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mbuff

// HLSwap32 swaps high and low bytes within each 16-bit word of a uint32.
// Example: 0x11223344 -> 0x22114433
func (b *Buffer) HLSwap32(v uint32) uint32 {
	if !b.hlswap {
		return v
	}
	part1 := (v & 0xFF00FF00) >> 8
	part2 := (v & 0x00FF00FF) << 8
	return part1 | part2
}

// HLSwap64 swaps high and low bytes within each 16-bit word of a uint64.
// Example: 0x1122334455667788 -> 0x2211443366558877
func (b *Buffer) HLSwap64(v uint64) uint64 {
	if !b.hlswap {
		return v
	}
	part1 := (v & 0xFF00FF00FF00FF00) >> 8
	part2 := (v & 0x00FF00FF00FF00FF) << 8
	return part1 | part2
}
