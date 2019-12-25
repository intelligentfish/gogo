package xcmd

import (
	"github.com/intelligentfish/gogo/pipe"
	"os"
	"os/exec"
)

// XCmd 扩展Cmd
type XCmd struct {
	*exec.Cmd
	nextFd    int
	readPipe  *pipe.PIPE
	writePipe *pipe.PIPE
}

// FromFd 从FD构建
func FromFd(readFd, writeFd int) *XCmd {
	object := &XCmd{
		readPipe:  &pipe.PIPE{},
		writePipe: &pipe.PIPE{},
	}
	object.readPipe.SetReadPipe(os.NewFile(uintptr(readFd), "readPipe"))
	object.writePipe.SetWritePipe(os.NewFile(uintptr(writeFd), "writePipe"))
	object.nextFd = 5
	return object
}

// NewXCmd 工厂方法
func New(name string, arg ...string) *XCmd {
	object := &XCmd{Cmd: exec.Command(name, arg...)}
	object.readPipe = pipe.NewPIPE()
	object.writePipe = pipe.NewPIPE()
	object.ExtraFiles = []*os.File{object.writePipe.GetReadPipe(), object.readPipe.GetWritePipe()}
	object.nextFd = 2 + len(object.ExtraFiles)
	return object
}

// Close 关闭
func (object *XCmd) Close() (err error) {
	if nil != object.readPipe {
		err = object.readPipe.Close()
	}
	if nil != object.writePipe {
		e := object.writePipe.Close()
		if nil == err && nil != e {
			err = e
		}
	}
	return
}

// NextFd 进程下一个可用的Fd
func (object *XCmd) NextFd() int {
	return object.nextFd
}

// AddFile 添加文件
func (object *XCmd) AddFile(f *os.File) *XCmd {
	object.ExtraFiles = append(object.ExtraFiles, f)
	object.nextFd++
	return object
}

// ParentWrite 父进程写
func (object *XCmd) ParentWrite(raw []byte) (err error) {
	err = object.writePipe.Write(raw)
	return
}

// ParentRead 父进程读
func (object *XCmd) ParentRead(maxSize int, callback func(raw []byte) bool) (err error) {
	err = object.readPipe.Read(maxSize, callback)
	return
}

// ChildWrite 子进程写
func (object *XCmd) ChildWrite(raw []byte) (err error) {
	err = object.writePipe.Write(raw)
	return
}

// ChildRead 子进程读
func (object *XCmd) ChildRead(maxSize int, callback func(raw []byte) bool) (err error) {
	err = object.readPipe.Read(maxSize, callback)
	return
}
