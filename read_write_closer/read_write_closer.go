package read_write_closer

import (
	"encoding/binary"
	"errors"
	"github.com/intelligentfish/gogo/buffer"
	"io"
	"sync/atomic"
)

// ReadWriteCloser 读写关闭器
type ReadWriteCloser struct {
	closed int32
	impl   io.ReadWriteCloser
}

// New 工厂方法
func New(impl io.ReadWriteCloser) *ReadWriteCloser {
	return &ReadWriteCloser{impl: impl}
}

// writeEmpty 写空
func (object *ReadWriteCloser) writeEmpty(raw []byte) (err error) {
	var n int
	for 0 < len(raw) {
		n, err = object.impl.Write(raw)
		if nil != err {
			return
		}
		raw = raw[n:]
	}
	return
}

// IsClosed 是否已关闭
func (object *ReadWriteCloser) IsClosed() bool {
	return 1 == atomic.LoadInt32(&object.closed)
}

// Close 关闭命名管道
func (object *ReadWriteCloser) Close() (err error) {
	if !atomic.CompareAndSwapInt32(&object.closed, 0, 1) {
		err = errors.New("already closed")
		return
	}
	if nil != object.impl {
		if err = object.impl.Close(); nil != err {
			return
		}
	}
	return
}

// Write 写入
func (object *ReadWriteCloser) Write(raw []byte) (err error) {
	if object.IsClosed() {
		err = errors.New("already closed")
		return
	}
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(raw)))
	if err = object.writeEmpty(header); nil != err {
		return
	}
	if err = object.writeEmpty(raw); nil != err {
		return
	}
	return
}

// Read 读取
func (object *ReadWriteCloser) Read(maxSize int, callback func(data []byte) bool) (err error) {
	if object.IsClosed() {
		err = errors.New("already closed")
		return
	}
	flag := true
	readBuf := buffer.GetPoolInstance().Borrow(maxSize)
	var n int
	for flag {
		n, err = object.impl.Read(readBuf.Internal[readBuf.GetWriteIndex():])
		if nil != err {
			if io.EOF == err {
				flag = false
			} else {
				return
			}
		}
		if 0 >= n {
			break
		}
		readBuf.SetWriteIndex(readBuf.GetWriteIndex() + n)
		for 4 <= readBuf.ReadableBytes() {
			chunkSize := int(binary.BigEndian.Uint32(readBuf.Slice(4)))
			if chunkSize+4 > readBuf.ReadableBytes() {
				break
			}
			readBuf.SetReadIndex(readBuf.GetReadIndex() + 4)
			flag = callback(readBuf.Slice(chunkSize))
			readBuf.SetReadIndex(readBuf.GetReadIndex() + chunkSize)
			readBuf.DiscardReadBytes()
			if !flag {
				break
			}
		}
	}
	if flag {
		callback(nil)
	}
	buffer.GetPoolInstance().Return(readBuf.DiscardReadBytes())
	return
}
