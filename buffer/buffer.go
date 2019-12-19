package buffer

import "encoding/binary"

// Buffer 缓冲区
type Buffer struct {
	Internal         []byte // 存储
	readIndex        int    // 读索引
	writeIndex       int    // 写索引
	markedReadIndex  int    // 标记的读索引
	markedWriteIndex int    // 标记的写索引
}

// NewBuffer 新建缓存
func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		Internal: make([]byte, capacity),
	}
}

// growth 增长
func (object *Buffer) growth(needSize int) *Buffer {
	if object.writeIndex+needSize > cap(object.Internal) {
		size := 2 * cap(object.Internal)
		if 0 == size {
			size = needSize
		}
		copied := make([]byte, size)
		copy(copied, object.Internal)
		object.Internal = copied
	}
	return object
}

// Initialize 初始化
func (object *Buffer) Initialize(capacity int) *Buffer {
	object.Internal = make([]byte, capacity)
	return object
}

// IsEmpty 是否为空
func (object *Buffer) IsEmpty() bool {
	return object.readIndex == object.writeIndex
}

// WriteableBytes 可写入字节数
func (object *Buffer) WriteableBytes() int {
	return cap(object.Internal) - object.writeIndex
}

// ReadableBytes 可读取字节数
func (object *Buffer) ReadableBytes() int {
	return object.writeIndex - object.readIndex
}

// MarkReadIndex 标记读索引
func (object *Buffer) MarkReadIndex() *Buffer {
	object.markedReadIndex = object.readIndex
	return object
}

// MarkWriteIndex 标记写索引
func (object *Buffer) MarkWriteIndex() *Buffer {
	object.markedWriteIndex = object.writeIndex
	return object
}

// ResetReadIndex 重置读索引
func (object *Buffer) ResetReadIndex() *Buffer {
	object.readIndex = object.markedReadIndex
	object.markedReadIndex = 0
	return object
}

// ResetWriteIndex 重置写索引
func (object *Buffer) ResetWriteIndex() *Buffer {
	object.writeIndex = object.markedWriteIndex
	object.markedWriteIndex = 0
	return object
}

// GetReadIndex 获取读索引
func (object *Buffer) GetReadIndex() int {
	return object.readIndex
}

// GetWriteIndex 获取写索引
func (object *Buffer) GetWriteIndex() int {
	return object.writeIndex
}

// SetReadIndex 设置读索引
func (object *Buffer) SetReadIndex(index int) *Buffer {
	object.readIndex = index
	return object
}

// SetWriteIndex 设置写索引
func (object *Buffer) SetWriteIndex(index int) *Buffer {
	object.writeIndex = index
	return object
}

// WriteBytes 写字节
func (object *Buffer) WriteBytes(bytes []byte) *Buffer {
	object.growth(len(bytes))
	copy(object.Internal[object.writeIndex:], bytes)
	object.writeIndex += len(bytes)
	return object
}

// Slice 返回切片
func (object *Buffer) Slice(size int) []byte {
	return object.Internal[object.readIndex : object.readIndex+size]
}

// WriteUint32 写Uint32
func (object *Buffer) WriteUint32(v uint32) *Buffer {
	object.growth(4)
	binary.BigEndian.PutUint32(object.Internal[object.writeIndex:], v)
	object.writeIndex += 4
	return object
}

// ReadUint32 读Uint32
func (object *Buffer) ReadUint32() (v uint32) {
	v = binary.BigEndian.Uint32(object.Internal[object.readIndex:])
	object.readIndex += 4
	return
}

// DiscardReadBytes 丢弃已读的数据
func (object *Buffer) DiscardReadBytes() *Buffer {
	copy(object.Internal, object.Internal[object.readIndex:object.writeIndex])
	object.writeIndex -= object.readIndex
	object.readIndex = 0
	return object
}
