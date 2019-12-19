package pipe

import (
	"encoding/binary"
	"errors"
	"github.com/intelligentfish/gogo/buffer"
	"github.com/intelligentfish/gogo/util"
	"io"
	"os"
	"sync/atomic"
)

// PIPE 管道
type PIPE struct {
	closed    int32
	readPipe  *os.File
	writePipe *os.File
}

// 工厂方法
func NewPIPE() *PIPE {
	object := &PIPE{}
	var err error
	object.readPipe, object.writePipe, err = os.Pipe()
	util.PanicOnError(err)
	return object
}

// GetReadPipe 获取管道
func (object *PIPE) GetReadPipe() *os.File {
	return object.readPipe
}

// SetReadPipe 设置读管道
func (object *PIPE) SetReadPipe(readPipe *os.File) *PIPE {
	object.readPipe = readPipe
	return object
}

// GetWritePipe 获取写管道
func (object *PIPE) GetWritePipe() *os.File {
	return object.writePipe
}

// SetWritePipe 设置写管道
func (object *PIPE) SetWritePipe(writePipe *os.File) *PIPE {
	object.writePipe = writePipe
	return object
}

// IsClosed 是否已关闭
func (object *PIPE) IsClosed() bool {
	return 1 == atomic.LoadInt32(&object.closed)
}

// Close 关闭命名管道
func (object *PIPE) Close() (err error) {
	if !atomic.CompareAndSwapInt32(&object.closed, 0, 1) {
		return
	}
	if nil != object.readPipe {
		if err = object.readPipe.Close(); nil != err {
			return
		}
		object.readPipe = nil
	}
	if nil != object.writePipe {
		if err = object.writePipe.Close(); nil != err {
			return
		}
		object.writePipe = nil
	}
	return
}

// writeEmpty 写空
func (object *PIPE) writeEmpty(raw []byte) (err error) {
	var n int
	for 0 < len(raw) {
		n, err = object.writePipe.Write(raw)
		if nil != err {
			return
		}
		raw = raw[n:]
	}
	return
}

// Write 写入
func (object *PIPE) Write(raw []byte) (err error) {
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
func (object *PIPE) Read(callback func(data []byte) bool) (err error) {
	if object.IsClosed() {
		err = errors.New("pipe closed")
		return
	}
	flag := true
	readBuf := buffer.NewBuffer(1 << 16)
	var n int
	for flag {
		n, err = object.readPipe.Read(readBuf.Internal[readBuf.GetReadIndex():])
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
