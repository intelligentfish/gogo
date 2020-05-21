package byte_buf

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"unsafe"
)

const (
	ByteSize    = 1
	Int16Size   = 2
	Int24Size   = 3
	Int32Size   = 4
	Int64Size   = 8
	Float32Size = 4
	Float64Size = 8
)

var (
	// BufferTooLarge 错误
	BufferTooLarge = errors.New("buffer tool large")
)

// ByteBuf 缓冲区
type ByteBuf struct {
	initCap int    // 初始容量
	maxCap  int    // 最大容量
	buf     []byte // 存储
	rIndex  int    // 读索引
	wIndex  int    // 写索引
	mrIndex int    // 标记读索引
	mwIndex int    // 标记写索引
}

// ByteBufOption 缓冲区选项
type ByteBufOption func(object *ByteBuf)

// InitCapOption 初始化容量选项
func InitCapOption(cap int) ByteBufOption {
	return func(object *ByteBuf) {
		object.initCap = cap
	}
}

// MaxCapOption 最大容量选项
func MaxCapOption(cap int) ByteBufOption {
	return func(object *ByteBuf) {
		object.maxCap = cap
	}
}

// WrappedOption 包装选项
func WrappedOption(slice []byte) ByteBufOption {
	return func(object *ByteBuf) {
		object.buf = slice
		object.initCap = len(slice)
	}
}

// New 工厂方法
func New(options ...ByteBufOption) *ByteBuf {
	object := &ByteBuf{}
	return object.Initialize(options...)
}

// move 移动内存
func (object *ByteBuf) move() *ByteBuf {
	copy(object.buf, object.buf[object.rIndex:object.wIndex])
	object.wIndex -= object.rIndex
	object.rIndex = 0
	return object
}

// Internal 内部缓冲区
func (object *ByteBuf) Internal() []byte {
	return object.buf
}

// Initialize 初始化
func (object *ByteBuf) Initialize(options ...ByteBufOption) *ByteBuf {
	if nil != object.buf {
		return object
	}

	for _, option := range options {
		option(object)
	}
	if nil == object.buf {
		if 0 >= object.initCap {
			object.initCap = 1 << 12
		}
		object.buf = make([]byte, object.initCap)
	}
	if 0 >= object.maxCap {
		object.maxCap = math.MaxInt32
	}
	return object
}

// InitCap 初始容量
func (object *ByteBuf) InitCap() int {
	return object.initCap
}

// MaxCap 最大容量
func (object *ByteBuf) MaxCap() int {
	return object.maxCap
}

// ReaderIndex 读索引
func (object *ByteBuf) ReaderIndex() int {
	return object.rIndex
}

// SetReaderIndex 设置读索引
func (object *ByteBuf) SetReaderIndex(index int) *ByteBuf {
	object.rIndex = index
	return object
}

// WriterIndex 写索引
func (object *ByteBuf) WriterIndex() int {
	return object.wIndex
}

// SetWriterIndex 设置写索引
func (object *ByteBuf) SetWriterIndex(index int) *ByteBuf {
	object.wIndex = index
	return object
}

// SetIndex 设置索引
func (object *ByteBuf) SetIndex(readIndex, writeIndex int) *ByteBuf {
	object.rIndex = readIndex
	object.wIndex = writeIndex
	return object
}

// ReadableBytes 可读字节数
func (object *ByteBuf) ReadableBytes() int {
	return object.wIndex - object.rIndex
}

// WriteableBytes 可写字节数
func (object *ByteBuf) WriteableBytes() int {
	return object.initCap - object.wIndex
}

// MaxWriteableBytes 最大可写字节数
func (object *ByteBuf) MaxWriteableBytes() int {
	return object.maxCap - object.wIndex
}

// IsReadable 会否可读
func (object *ByteBuf) IsReadable() bool {
	return object.wIndex > object.rIndex
}

// IsReadableBytes 是否可读指定大小字节
func (object *ByteBuf) IsReadableBytes(bytes int) bool {
	return bytes >= object.wIndex-object.rIndex
}

// IsWriteable 是否可写
func (object *ByteBuf) IsWriteable() bool {
	return cap(object.buf) > object.wIndex
}

// IsWriteableBytes 是否可写指定大小字节
func (object *ByteBuf) IsWriteableBytes(bytes int) bool {
	return bytes > object.maxCap-object.wIndex
}

// Clear 清空
func (object *ByteBuf) Clear() *ByteBuf {
	object.rIndex = 0
	object.wIndex = 0
	object.buf = make([]byte, object.initCap)
	return object
}

// MarkReaderIndex 标记读索引
func (object *ByteBuf) MarkReaderIndex() *ByteBuf {
	object.mrIndex = object.rIndex
	return object
}

// ResetReaderIndex 重置读索引
func (object *ByteBuf) ResetReaderIndex() *ByteBuf {
	object.rIndex = object.mrIndex
	object.mrIndex = 0
	return object
}

// MarkWriterIndex 标记写索引
func (object *ByteBuf) MarkWriterIndex() *ByteBuf {
	object.mwIndex = object.wIndex
	return object
}

// ResetWriterIndex 重置写索引
func (object *ByteBuf) ResetWriterIndex() *ByteBuf {
	object.wIndex = object.mwIndex
	object.mwIndex = 0
	return object
}

// DiscardReadBytes 丢弃读字节
func (object *ByteBuf) DiscardReadBytes() *ByteBuf {
	if 0 != object.rIndex {
		if object.rIndex != object.wIndex {
			copy(object.buf, object.buf[object.rIndex:object.wIndex])
		}
		object.wIndex -= object.rIndex
		object.rIndex = 0
	}
	return object
}

// DiscardAllBytes 丢去所有字节
func (object *ByteBuf) DiscardAllBytes() *ByteBuf {
	return object.SetReaderIndex(object.WriterIndex()).DiscardReadBytes()
}

// EnsureWriteable 确定可写
func (object *ByteBuf) EnsureWriteable(size int) *ByteBuf {
	// 容量低于4KB时2倍扩容，否则按照当前容量的1/16递增
	const BK4 = 1 << 12
	if size > object.WriteableBytes() {
		readableBytes := object.ReadableBytes()
		writeableBytes := object.WriteableBytes()
		if size <= object.rIndex+writeableBytes {
			object.move()
		} else {
			var newCap int
			remain := size - object.rIndex - writeableBytes
			if BK4 > object.initCap {
				newCap = object.initCap * 2
				for newCap < readableBytes+remain {
					newCap *= 2
				}
			} else {
				step := object.initCap / 16
				newCap = object.initCap + step
				for newCap < readableBytes+remain {
					newCap += step
				}
			}

			buf := make([]byte, newCap)
			copy(buf, object.buf[object.rIndex:object.wIndex])
			object.initCap = newCap
			object.buf = buf
			object.wIndex -= object.rIndex
			object.rIndex = 0
		}
	}
	return object
}

// PeekByte 查看Byte值
func (object *ByteBuf) PeekByte(index int) byte {
	return object.buf[index]
}

// GetBool 获取Bool值
func (object *ByteBuf) GetBool() bool {
	v := 1 == object.buf[object.rIndex]
	object.rIndex++
	return v
}

// GetByte 获取Byte
func (object *ByteBuf) GetByte() byte {
	v := object.buf[object.rIndex]
	object.rIndex++
	return v
}

// GetUint8 获取Uint8
func (object *ByteBuf) GetUint8() uint8 {
	return object.GetByte()
}

// GetUint16 获取Uint16
func (object *ByteBuf) GetUint16() uint16 {
	v := binary.BigEndian.Uint16(object.buf[object.rIndex:])
	object.rIndex += Int16Size
	return v
}

// GetUint16LE 获取小端Uint16
func (object *ByteBuf) GetUint16LE() uint16 {
	v := binary.LittleEndian.Uint16(object.buf[object.rIndex:])
	object.rIndex += Int16Size
	return v
}

// GetMedium 获取中等值
func (object *ByteBuf) GetMedium() uint32 {
	v := uint32(object.buf[object.rIndex+2]) << 16
	v |= uint32(object.buf[object.rIndex+1]) << 8
	v |= uint32(object.buf[object.rIndex])
	object.rIndex += Int24Size
	return v
}

// GetMediumLE 获取小端中等值
func (object *ByteBuf) GetMediumLE() uint32 {
	v := uint32(object.buf[object.rIndex+2]) << 16
	v |= uint32(object.buf[object.rIndex+1]) << 8
	v |= uint32(object.buf[object.rIndex])
	object.rIndex += Int24Size
	return v
}

// GetUint32 获取大端uint32
func (object *ByteBuf) GetUint32() uint32 {
	v := binary.BigEndian.Uint32(object.buf[object.rIndex:])
	object.rIndex += Int32Size
	return v
}

// GetUint32LE 获取小端uin32
func (object *ByteBuf) GetUint32LE() uint32 {
	v := binary.LittleEndian.Uint32(object.buf[object.rIndex:])
	object.rIndex += Int32Size
	return v
}

// GetUint64 获取大端uint64
func (object *ByteBuf) GetUint64() uint64 {
	v := binary.BigEndian.Uint64(object.buf[object.rIndex:])
	object.rIndex += Int64Size
	return v
}

// GetUint64LE 获取小端uint64
func (object *ByteBuf) GetUint64LE() uint64 {
	v := binary.LittleEndian.Uint64(object.buf[object.rIndex:])
	object.rIndex += Int64Size
	return v
}

// GetFloat32 获取大端float32
func (object *ByteBuf) GetFloat32() float32 {
	v := binary.BigEndian.Uint32(object.buf[object.rIndex:])
	object.rIndex += Float32Size
	return math.Float32frombits(v)
}

// GetFloat32LE 获取小端float32
func (object *ByteBuf) GetFloat32LE() float32 {
	v := binary.LittleEndian.Uint32(object.buf[object.rIndex:])
	object.rIndex += Float32Size
	return math.Float32frombits(v)
}

// GetFloat64 获取大端float64
func (object *ByteBuf) GetFloat64() float64 {
	v := binary.BigEndian.Uint64(object.buf[object.rIndex:])
	object.rIndex += Float64Size
	return math.Float64frombits(v)
}

// GetFloat64LE 获取小端float64
func (object *ByteBuf) GetFloat64LE() float64 {
	v := binary.LittleEndian.Uint64(object.buf[object.rIndex:])
	object.rIndex += Float64Size
	return math.Float64frombits(v)
}

// GetBytes 获取字节
func (object *ByteBuf) GetBytes(length int) []byte {
	if length > object.ReadableBytes() {
		length = object.ReadableBytes()
	}
	v := make([]byte, length)
	copy(v, object.buf[object.rIndex:])
	object.rIndex += length
	return v
}

// WriteBool 写Bool
func (object *ByteBuf) WriteBool(v bool) *ByteBuf {
	object.EnsureWriteable(ByteSize)
	if v {
		object.buf[object.wIndex] = 1
	} else {
		object.buf[object.wIndex] = 0
	}
	object.wIndex++
	return object
}

// WriteByte 写字节
func (object *ByteBuf) WriteByte(v byte) *ByteBuf {
	object.EnsureWriteable(ByteSize)
	object.buf[object.wIndex] = v
	object.wIndex++
	return object
}

// WriteUint8 写uint8
func (object *ByteBuf) WriteUint8(v uint8) *ByteBuf {
	object.EnsureWriteable(ByteSize)
	object.buf[object.wIndex] = v
	object.wIndex++
	return object
}

// WriteUint16 写大端uint16
func (object *ByteBuf) WriteUint16(v uint16) *ByteBuf {
	object.EnsureWriteable(Int16Size)
	binary.BigEndian.PutUint16(object.buf[object.wIndex:], v)
	object.wIndex += Int16Size
	return object
}

// WriteUint16LE 写小端uint16
func (object *ByteBuf) WriteUint16LE(v uint16) *ByteBuf {
	object.EnsureWriteable(Int16Size)
	binary.LittleEndian.PutUint16(object.buf[object.wIndex:], v)
	object.wIndex += Int16Size
	return object
}

// WriteMedium 写中值
func (object *ByteBuf) WriteMedium(v int32) *ByteBuf {
	object.EnsureWriteable(Int24Size)
	object.buf[object.wIndex] = byte(v & 0xff)
	object.buf[object.wIndex+1] = byte((v >> 8) & 0xff)
	object.buf[object.wIndex+2] = byte((v >> 16) & 0xff)
	object.wIndex += Int24Size
	return object
}

// WriteMediumLE 写小端中值
func (object *ByteBuf) WriteMediumLE(v int32) *ByteBuf {
	object.EnsureWriteable(Int24Size)
	object.buf[object.wIndex] = byte(v & 0xff)
	object.buf[object.wIndex+1] = byte((v >> 8) & 0xff)
	object.buf[object.wIndex+2] = byte((v >> 16) & 0xff)
	object.wIndex += Int24Size
	return object
}

// WriteUint32 写uint32
func (object *ByteBuf) WriteUint32(v uint32) *ByteBuf {
	object.EnsureWriteable(Int32Size)
	binary.BigEndian.PutUint32(object.buf[object.wIndex:], v)
	object.wIndex += Int32Size
	return object
}

// WriteUint32LE 写小端uint32
func (object *ByteBuf) WriteUint32LE(v uint32) *ByteBuf {
	object.EnsureWriteable(Int32Size)
	binary.LittleEndian.PutUint32(object.buf[object.wIndex:], v)
	object.wIndex += Int32Size
	return object
}

// WriteUint64 写uint64
func (object *ByteBuf) WriteUint64(v uint64) *ByteBuf {
	object.EnsureWriteable(Int64Size)
	binary.BigEndian.PutUint64(object.buf[object.wIndex:], v)
	object.wIndex += Int64Size
	return object
}

// WriteUint64LE 写小端uint64
func (object *ByteBuf) WriteUint64LE(v uint64) *ByteBuf {
	object.EnsureWriteable(Int64Size)
	binary.LittleEndian.PutUint64(object.buf[object.wIndex:], v)
	object.wIndex += Int64Size
	return object
}

// WriteFloat32 写float32
func (object *ByteBuf) WriteFloat32(v float32) *ByteBuf {
	object.EnsureWriteable(Float32Size)
	uv := *(*uint32)(unsafe.Pointer(&v))
	binary.BigEndian.PutUint32(object.buf[object.wIndex:], uv)
	object.wIndex += Float32Size
	return object
}

// WriteFloat32LE 写小端float32
func (object *ByteBuf) WriteFloat32LE(v float32) *ByteBuf {
	object.EnsureWriteable(Float32Size)
	uv := *(*uint32)(unsafe.Pointer(&v))
	binary.LittleEndian.PutUint32(object.buf[object.wIndex:], uv)
	object.wIndex += Float32Size
	return object
}

// WriteFloat64 写float64
func (object *ByteBuf) WriteFloat64(v float64) *ByteBuf {
	object.EnsureWriteable(Float64Size)
	uv := *(*uint64)(unsafe.Pointer(&v))
	binary.BigEndian.PutUint64(object.buf[object.wIndex:], uv)
	object.wIndex += Float64Size
	return object
}

// WriteFloat64LE 写小端float64
func (object *ByteBuf) WriteFloat64LE(v float64) *ByteBuf {
	object.EnsureWriteable(Float64Size)
	uv := *(*uint64)(unsafe.Pointer(&v))
	binary.LittleEndian.PutUint64(object.buf[object.wIndex:], uv)
	object.wIndex += Float64Size
	return object
}

// WriteBytes 写字节
func (object *ByteBuf) WriteBytes(bytes []byte) *ByteBuf {
	object.EnsureWriteable(len(bytes))
	copy(object.buf[object.wIndex:], bytes)
	object.wIndex += len(bytes)
	return object
}

// Slice 获取切片
func (object *ByteBuf) Slice(index, length int) []byte {
	return object.buf[index : index+length]
}

// Duplicate 复制整个ByteBuf
func (object *ByteBuf) Duplicate() *ByteBuf {
	dup := &ByteBuf{
		initCap: object.initCap,
		maxCap:  object.maxCap,
		buf:     make([]byte, len(object.buf)),
		rIndex:  object.rIndex,
		wIndex:  object.wIndex,
		mrIndex: object.mrIndex,
		mwIndex: object.mwIndex,
	}
	copy(dup.buf, object.buf)
	return dup
}

// Skip 略过连续特定字符
func (object *ByteBuf) Skip(v byte) *ByteBuf {
	for object.rIndex < object.wIndex {
		if object.buf[object.rIndex] != v {
			break
		}
		object.rIndex++
	}
	return object
}

// TakeUntil 获取直到遇到字符终止
func (object *ByteBuf) TakeUntil(v byte, setRIndex bool) []byte {
	for rIndex := object.rIndex; rIndex < object.wIndex; rIndex++ {
		if v == object.buf[rIndex] {
			tmp := object.buf[object.rIndex:rIndex]
			if setRIndex {
				object.rIndex = rIndex
			}
			return tmp
		}
	}
	return nil
}

// WithReadBuf
func (object *ByteBuf) WithReadBuf(callback func(p []byte)) {
	callback(object.buf[object.rIndex:object.wIndex])
}

// WithWriteBuf
func (object *ByteBuf) WithWriteBuf(callback func(p []byte)) {
	callback(object.buf[object.wIndex:])
}

// ReadFrom 读取数据
func (object *ByteBuf) ReadFrom(r io.Reader, growSize int) (n int, err error) {
	if growSize > object.WriteableBytes() {
		object.EnsureWriteable(growSize)
	}
	object.WithWriteBuf(func(buf []byte) {
		n, err = r.Read(buf)
		if 0 < n {
			object.wIndex += n
		}
	})
	return
}
