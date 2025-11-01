# MBuff

**English** | [中文](README_zh.md)

A high-performance buffer library for efficient binary data processing in Go. MBuff provides a simple yet powerful API for reading and writing binary data with support for different byte orders and custom byte swapping.

## Installation

```bash
go get github.com/tayne3/mbuff
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/tayne3/mbuff"
)

func main() {
    // Create a buffer with initial capacity
    m := mbuff.New(1024)
    
    // Write data
    m.PutU8(0x01)
    m.PutU16(0x1234)
    m.PutU32(0x12345678)
    m.PutU64(0x123456789ABCDEF0)
    
    // Rewind to read from the beginning
    m.Rewind()
    
    // Read data
    v1 := m.TakeU8()    // 0x01
    v2 := m.TakeU16()   // 0x1234
    v3 := m.TakeU32()   // 0x12345678
    v4 := m.TakeU64()   // 0x123456789ABCDEF0
    
    fmt.Printf("Values: %x, %x, %x, %x\n", v1, v2, v3, v4)
}
```
