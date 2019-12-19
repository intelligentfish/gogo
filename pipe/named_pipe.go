package pipe

import (
	"encoding/binary"
	"errors"
	"github.com/intelligentfish/gogo/buffer"
	"io"
	"os"
	"sync/atomic"
	"syscall"
)

//TODO 只支持Linux
// NamedPipe 命名管道
type NamedPipe struct {
	name   string
	create bool
	closed int32
	file   *os.File
}

// MakeNamedPipe 创建命名管道
func MakeNamedPipe(name string, create bool) (object *NamedPipe, err error) {
	object = &NamedPipe{
		name: name,
	}
	if create {
		os.Remove(name)
		if err = syscall.Mkfifo(name, 0666); nil != err {
			return
		}
	}
	flag := os.O_RDWR | os.O_CREATE | os.O_APPEND
	object.file, err = os.OpenFile(name, flag, os.ModeNamedPipe)
	return
}

// IsClosed 是否已关闭
func (object *NamedPipe) IsClosed() bool {
	return 1 == atomic.LoadInt32(&object.closed)
}

// Close 关闭命名管道
func (object *NamedPipe) Close() (err error) {
	if !atomic.CompareAndSwapInt32(&object.closed, 0, 1) {
		return
	}
	if nil != object.file {
		if err = object.file.Close(); nil != err {
			return
		}
	}
	if object.create {
		err = os.Remove(object.name)
	}
	return
}

// writeEmpty 写空
func (object *NamedPipe) writeEmpty(raw []byte) (err error) {
	var n int
	for 0 < len(raw) {
		n, err = object.file.Write(raw)
		if nil != err {
			return
		}
		raw = raw[n:]
	}
	return
}

// Write 写入
func (object *NamedPipe) Write(raw []byte) (err error) {
	if object.IsClosed() {
		err = errors.New("pipe closed")
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
func (object *NamedPipe) Read(callback func(data []byte) bool) (err error) {
	if object.IsClosed() {
		err = errors.New("pipe closed")
		return
	}
	flag := true
	readBuf := buffer.NewBuffer(1 << 16)
	var n int
	for flag {
		n, err = object.file.Read(readBuf.Internal[readBuf.GetReadIndex():])
		if nil != err {
			if io.EOF == err {
				flag = false
			} else {
				return
			}
		}
		if 0 == n {
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
			if !flag {
				break
			}
			readBuf.SetReadIndex(readBuf.GetReadIndex() + chunkSize)
			readBuf.DiscardReadBytes()
		}
	}
	return
}
