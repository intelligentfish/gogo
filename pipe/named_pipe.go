package pipe

import (
	"github.com/intelligentfish/gogo/util"
	"os"
	"syscall"
)

//TODO 只支持Linux
// NamedPipe 命名管道
type NamedPipe struct {
	*util.ReadWriterCloserUtil
	name   string
	create bool
	file   *os.File
}

// MakeNamedPipe 创建命名管道
func MakeNamedPipe(name string, create bool) (object *NamedPipe, err error) {
	if create {
		os.Remove(name)
		if err = syscall.Mkfifo(name, 0666); nil != err {
			return
		}
	}
	flag := os.O_RDWR | os.O_CREATE | os.O_APPEND
	if object.file, err = os.OpenFile(name, flag, os.ModeNamedPipe); nil != err {
		return
	}
	object = &NamedPipe{
		ReadWriterCloserUtil: util.NewReadWriterCloserUtil(object.file),
		name:                 name,
	}
	return
}

// Close 关闭
func (object *NamedPipe) Close() (err error) {
	err = object.ReadWriterCloserUtil.Close()
	if nil != err && object.create {
		err = os.Remove(object.name)
	}
	return
}
