package pipe

import (
	"github.com/intelligentfish/gogo/util"
	"os"
)

// PIPE 管道
type PIPE struct {
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

// GetReadPipe
func (object *PIPE) GetReadPipe() *os.File {
	return object.readPipe
}

// GetWritePipe
func (object *PIPE) GetWritePipe() *os.File {
	return object.writePipe
}
