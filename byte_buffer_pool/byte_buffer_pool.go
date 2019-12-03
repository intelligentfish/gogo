package byte_buffer_pool

import (
	"bytes"
	"sync"
)

// 变量
var (
	byteBufferPoolOnce         sync.Once       // 执行一次
	byteBufferPoolOnceInstance *ByteBufferPool // 单实例
)

// 缓冲区池
type ByteBufferPool struct {
	*sync.Pool
}

// newByteBufferPool 工厂方法
func newByteBufferPool() *ByteBufferPool {
	return &ByteBufferPool{
		Pool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// 借用ByteBuffer
func (object *ByteBufferPool) BorrowByteBuffer() *bytes.Buffer {
	return object.Get().(*bytes.Buffer)
}

// 退还ByteBuffer
func (object *ByteBufferPool) ReturnByteBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	object.Put(buffer)
}

// GetByteBufferPoolInstance 获取实例
func GetByteBufferPoolInstance() *ByteBufferPool {
	byteBufferPoolOnce.Do(func() {
		byteBufferPoolOnceInstance = newByteBufferPool()
	})
	return byteBufferPoolOnceInstance
}
