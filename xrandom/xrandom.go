package xrandom

import (
	"bytes"
	"encoding/base64"
	"math/rand"
	"time"
)

// 变量
var (
	privateSeed int64 // 种子
)

// NewSource 新建Source
func NewSource() rand.Source {
	seed := time.Now().UnixNano()
	for 0 != privateSeed && seed == privateSeed {
		seed = time.Now().UnixNano()
		time.Sleep(1 * time.Nanosecond)
	}
	privateSeed = seed
	return rand.NewSource(seed)
}

// Bytes 随机字节数组
func Bytes(size int) []byte {
	r := NewSource()
	sb := &bytes.Buffer{}
	for i := 0; i < size; i++ {
		sb.WriteByte(byte(r.Int63()))
	}
	return sb.Bytes()
}

// Base64String 随机Base64字符串
func Base64String(size int) string {
	return base64.StdEncoding.EncodeToString(Bytes(size))
}

// Bool 随机bool
func Bool() bool {
	r := NewSource()
	return 0 == r.Int63()%2
}
