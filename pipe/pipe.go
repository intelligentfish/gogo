package pipe

import (
	"errors"
	"github.com/intelligentfish/gogo/util"
	"os"
	"sync/atomic"
)

// PIPE 管道
type PIPE struct {
	closed    int32
	readPipe  *os.File
	writePipe *os.File
	readUtil  *util.ReadWriterCloserUtil
	writeUtil *util.ReadWriterCloserUtil
}

// 工厂方法
func NewPIPE() *PIPE {
	object := &PIPE{}
	var err error
	object.readPipe, object.writePipe, err = os.Pipe()
	util.PanicOnError(err)
	object.readUtil = util.NewReadWriterCloserUtil(object.readPipe)
	object.writeUtil = util.NewReadWriterCloserUtil(object.writePipe)
	return object
}

// GetReadPipe 获取管道
func (object *PIPE) GetReadPipe() *os.File {
	return object.readPipe
}

// SetReadPipe 设置读管道
func (object *PIPE) SetReadPipe(readPipe *os.File) *PIPE {
	object.readPipe = readPipe
	if nil == object.readUtil {
		object.readUtil = util.NewReadWriterCloserUtil(readPipe)
	}
	return object
}

// GetWritePipe 获取写管道
func (object *PIPE) GetWritePipe() *os.File {
	return object.writePipe
}

// SetWritePipe 设置写管道
func (object *PIPE) SetWritePipe(writePipe *os.File) *PIPE {
	object.writePipe = writePipe
	if nil == object.writeUtil {
		object.writeUtil = util.NewReadWriterCloserUtil(writePipe)
	}
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
	if nil != object.readUtil && !object.readUtil.IsClosed() {
		object.readUtil.Close()
	}
	if nil != object.writeUtil && !object.writeUtil.IsClosed() {
		object.writeUtil.Close()
	}
	return
}

// Write 写入
func (object *PIPE) Write(raw []byte) (err error) {
	if object.IsClosed() {
		err = errors.New("pipe closed")
		return
	}
	object.writeUtil.Write(raw)
	return
}

// Read 读取
func (object *PIPE) Read(callback func(data []byte) bool) (err error) {
	if object.IsClosed() {
		err = errors.New("pipe closed")
		return
	}
	err = object.readUtil.Read(callback)
	return
}
