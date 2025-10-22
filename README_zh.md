# MBuff

**English** | [中文](README_zh.md)

一个用于 Go 语言高效二进制数据处理的高性能缓冲区库。MBuff 提供了简单而强大的 API，用于读写二进制数据，支持不同的字节序和自定义字节交换。

## 安装

```bash
go get github.com/tayne3/mbuff
```

## 快速开始

```go
package main

import (
    "fmt"

    "github.com/tayne3/mbuff"
)

func main() {
    // Create a buffer with 1024 bytes
    buf := make([]byte, 1024)
    m := mbuff.NewMBuff(buf)
    
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
