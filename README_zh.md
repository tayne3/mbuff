# MBuff

[English](README.md) | **中文**

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
    // 创建指定初始容量的缓冲区
    b := mbuff.NewBuffer(1024)

    // 写入数据
    b.PutU8(0x01)
    b.PutU16(0x1234)
    b.PutU32(0x12345678)
    b.PutU64(0x123456789ABCDEF0)
    
    // 重置到开头以读取数据
    b.Rewind()
    
    // 读取数据
    v1 := b.TakeU8()    // 0x01
    v2 := b.TakeU16()   // 0x1234
    v3 := b.TakeU32()   // 0x12345678
    v4 := b.TakeU64()   // 0x123456789ABCDEF0
    
    fmt.Printf("Values: %x, %x, %x, %x\n", v1, v2, v3, v4)
}
```
