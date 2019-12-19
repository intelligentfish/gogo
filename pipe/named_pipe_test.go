// +build !windows

package pipe

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/intelligentfish/gogo/util"
	"os"
	"sync"
	"testing"
)

func TestNamedPipe(t *testing.T) {
	os.Mkdir("logs", 0777)

	flag.Parse()
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "logs")

	writeNamedPipe, err := NewNamedPipe("4feac5ac-21fa-11ea-90a8-00155d0c5004", true)
	util.PanicOnError(err)
	readNamedPipe, err := NewNamedPipe("5bde53cc-2206-11ea-b31d-00155d0c5004", true)
	util.PanicOnError(err)

	defer writeNamedPipe.Close()
	defer readNamedPipe.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		writeNamedPipe, err := NewNamedPipe("5bde53cc-2206-11ea-b31d-00155d0c5004", false)
		util.PanicOnError(err)
		readNamedPipe, err := NewNamedPipe("4feac5ac-21fa-11ea-90a8-00155d0c5004", false)
		util.PanicOnError(err)

		defer writeNamedPipe.Close()
		defer readNamedPipe.Close()

		err = readNamedPipe.Read(func(data []byte) bool {
			req := string(data)
			glog.Info("request: ", req)
			switch req {
			case "EXIT":
				err = writeNamedPipe.Write([]byte("EXIT"))
				if nil != err {
					glog.Error(err)
				}
				return false
			}
			return true
		})
		util.PanicOnError(err)
	}()

	go func() {
		defer wg.Done()

		err = readNamedPipe.Read(func(data []byte) bool {
			req := string(data)
			glog.Info("response: ", req)
			switch req {
			case "EXIT":
				return false
			}
			return true
		})
		util.PanicOnError(err)
	}()

	for i := 0; i < 10; i++ {
		err = writeNamedPipe.Write([]byte(fmt.Sprintf("value=%d", i)))
		util.PanicOnError(err)
	}

	err = writeNamedPipe.Write([]byte("EXIT"))
	util.PanicOnError(err)

	wg.Wait()
}
